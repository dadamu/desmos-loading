package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/desmos-labs/cosmos-go-wallet/client"
	wallettypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/desmos-labs/desmos/v6/app"
	poststypes "github.com/desmos-labs/desmos/v6/x/posts/types"
	profilestypes "github.com/desmos-labs/desmos/v6/x/profiles/types"
)

type Service struct {
	Wallet *wallet.Wallet

	size       int
	subspaceID uint64

	duration time.Duration
}

func NewService(cfg *Config, hdPath string) *Service {
	encodingCfg := app.MakeEncodingConfig()
	txConfig, cdc := encodingCfg.TxConfig, encodingCfg.Codec

	walletClient, err := client.NewClient(cfg.Chain, cdc)
	if err != nil {
		panic(fmt.Errorf("error while creating wallet client"))
	}

	account := cfg.Account
	account.HDPath = hdPath
	wallet, err := wallet.NewWallet(account, walletClient, txConfig)
	if err != nil {
		panic(fmt.Errorf("error while creating cosmos wallet: %s", err))
	}

	return &Service{
		Wallet: wallet,

		size:       cfg.Size,
		subspaceID: cfg.SubspaceID,

		duration: cfg.Duration,
	}
}

func (s *Service) getSequence() (uint64, error) {
	res, err := s.Wallet.Client.AuthClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: s.Wallet.AccAddress()})
	if err != nil {
		return 0, fmt.Errorf("error while getting account from node")
	}

	err = res.UnpackInterfaces(s.Wallet.Client.Codec)
	if err != nil {
		return 0, fmt.Errorf("error while unpacking response")
	}

	account, ok := res.Account.GetCachedValue().(authtypes.AccountI)
	if !ok {
		return 0, fmt.Errorf("error while get account from cached value")
	}

	return account.GetSequence(), nil
}

func (s *Service) SendFeesToTester(expRound int, recipients []string) {
	msgs := make([]sdk.Msg, len(recipients))

	for i, recipient := range recipients {
		msgs[i] = banktypes.NewMsgSend(
			sdk.MustAccAddressFromBech32(s.Wallet.AccAddress()),
			sdk.MustAccAddressFromBech32(recipient),
			s.calculateAllFeesDuringTesting(expRound),
		)
	}

	sequence, err := s.getSequence()
	if err != nil {
		panic(err)
	}

	err = s.broadcast(msgs, sequence, s.getGasLimit(msgs))
	if err != nil {
		panic(err)
	}
}

func (s *Service) calculateAllFeesDuringTesting(round int) sdk.Coins {
	allGas := s.getGasLimit(s.generatePostMsgs(1, s.size, s.Wallet.AccAddress())) * uint64(round)
	allGas += s.getGasLimit(s.getSaveProfileMsg())
	return s.Wallet.Client.GetFees(int64(allGas))
}

func (s *Service) SaveProfileIfNotExist() {
	profilesClient := profilestypes.NewQueryClient(s.Wallet.Client.GRPCConn)
	profile, err := profilesClient.Profile(context.Background(), profilestypes.NewQueryProfileRequest(s.Wallet.AccAddress()))
	if err == nil || profile != nil {
		return
	}

	sequence, err := s.getSequence()
	if err != nil {
		panic(err)
	}

	msgs := s.getSaveProfileMsg()
	err = s.broadcast(msgs, sequence, s.getGasLimit(msgs))
	if err != nil {
		panic(err)
	}
}

func (s *Service) RunTasks(round int) {
	fmt.Println(fmt.Sprintf("tester address: %s", s.Wallet.AccAddress()))

	ticker := time.NewTicker(time.Millisecond * 100)
	sequence, err := s.getSequence()
	if err != nil {
		panic(err)
	}

	msgs := s.generatePostMsgs(s.subspaceID, s.size, s.Wallet.AccAddress())
	gas := s.getGasLimit(msgs)

	roundPerTick := math.Floor(float64(round) / float64(s.duration.Milliseconds()) * 100)
	count := 0
	for range ticker.C {
		if count >= round {
			break
		}

		//startTime := time.Now()
		countPerTick := 0
		for {
			if countPerTick >= int(roundPerTick) {
				break
			}

			err := s.broadcast(msgs, sequence, gas)
			if err != nil {
				fmt.Println(err)
				continue
			}

			sequence += 1
			countPerTick += 1
			count += 1
		}
		//fmt.Println(fmt.Sprintf("%s round used", s.Wallet.AccAddress()), time.Now().Sub(startTime))
	}

}

func (s *Service) broadcast(msgs []sdk.Msg, sequence uint64, gasLimit uint64) error {
	// Build the transaction data
	txData := wallettypes.NewTransactionData(msgs...).
		WithFeeAuto().
		WithGasLimit(gasLimit).
		WithSequence(sequence)

	// Broadcast the transaction
	response, err := s.Wallet.BroadcastTxSync(txData)
	if err != nil {
		panic(err)
	}

	// Check the response
	if response.Code != 0 {
		panic(fmt.Errorf("%s error while broadcasting msg: %s", s.Wallet.AccAddress(), response.RawLog))
	}

	return nil
}

func (s *Service) getGasLimit(msgs []sdk.Msg) uint64 {
	// Build the transaction data for simulation
	txData := wallettypes.NewTransactionData(msgs...).
		WithGasAuto().
		WithFeeAuto()

	builder, err := s.Wallet.BuildTx(txData)
	if err != nil {
		panic(err)
	}

	gasLimit, err := s.Wallet.Client.SimulateTx(builder.GetTx())
	if err != nil {
		panic(err)
	}

	return gasLimit
}

func (s *Service) generatePostMsgs(subspaceID uint64, size int, author string) []sdk.Msg {
	msgs := make([]sdk.Msg, size)

	for i := 0; i < size; i++ {
		msgs[i] = poststypes.NewMsgCreatePost(
			subspaceID,
			0,
			"",
			"Lorem ipsum dolor sit amet, id veri scriptorem mei, in pri meliore incorrupte, at repudiandae vituperatoribus duo. Et cum commune qualisque, aperiam voluptua voluptatum mei ad. Eripuit explicari laboramus mel no, vix et causae omnesque, nibh tempor perfecto vis at. Ne quis denique copiosae est, sit et volumus abhorreant dissentiet, malorum inermis intellegebat mea an. Tempor iisque sit ne.",
			0,
			poststypes.REPLY_SETTING_EVERYONE,
			nil,
			nil,
			nil,
			nil,
			author,
		)
	}

	return msgs
}

func (s *Service) getSaveProfileMsg() []sdk.Msg {
	return []sdk.Msg{profilestypes.NewMsgSaveProfile(
		fmt.Sprintf("tester_%s", rand.Str(10)),
		"tester",
		"",
		"",
		"",
		s.Wallet.AccAddress(),
	)}
}
