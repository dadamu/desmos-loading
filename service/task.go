package service

import (
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/cosmos-go-wallet/client"
	wallettypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/desmos-labs/desmos/v6/app"
)

type Service struct {
	mu     sync.Mutex
	wallet *wallet.Wallet

	size       int
	subspaceID uint64
	gasLimit   uint64
	sequence   uint64

	duration time.Duration
	round    int
}

func NewService(cfg *Config) Service {
	encodingCfg := app.MakeEncodingConfig()
	txConfig, cdc := encodingCfg.TxConfig, encodingCfg.Codec

	walletClient, err := client.NewClient(cfg.Chain, cdc)
	if err != nil {
		panic(fmt.Errorf("error while creating wallet client"))
	}

	wallet, err := wallet.NewWallet(cfg.Account, walletClient, txConfig)
	if err != nil {
		panic(fmt.Errorf("error while creating cosmos wallet: %s", err))
	}

	sequence, err := GetAccountSequence(cdc, walletClient.AuthClient, wallet.AccAddress())

	return Service{
		wallet: wallet,

		size:       cfg.Size,
		subspaceID: cfg.SubspaceID,
		gasLimit:   GetGasLimit(wallet, generatePostMsgs(cfg.SubspaceID, cfg.Size, wallet.AccAddress())),
		sequence:   sequence,

		round:    cfg.Round,
		duration: cfg.Duration,
	}
}

func (s *Service) RunTasks() {
	var wg sync.WaitGroup
	ticker := time.NewTicker(time.Duration(s.duration.Nanoseconds() / int64(s.round)))

	count := 0
	for range ticker.C {
		if count >= s.round {
			break
		}

		wg.Add(1)
		go func() {
			err := s.broadcast(generatePostMsgs(s.subspaceID, s.size, s.wallet.AccAddress()))
			if err != nil {
				fmt.Println(err)
			}

			wg.Done()
		}()

		count += 1
	}
	wg.Wait()
}

func (s *Service) broadcast(msgs []sdk.Msg) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Build the transaction data
	txData := wallettypes.NewTransactionData(msgs...).
		WithGasLimit(s.gasLimit).
		WithFeeAuto().
		WithSequence(s.sequence)
	s.sequence += 1

	// Broadcast the transaction
	response, err := s.wallet.BroadcastTxSync(txData)
	fmt.Println(s.sequence, response.TxHash)
	if err != nil {
		return err
	}

	// Check the response
	if response.Code != 0 {
		return fmt.Errorf("error while broadcasting msg: %s", response.RawLog)
	}

	return nil
}
