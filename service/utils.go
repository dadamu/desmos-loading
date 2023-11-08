package service

import (
	"context"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	wallettypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	poststypes "github.com/desmos-labs/desmos/v6/x/posts/types"
	"github.com/rs/zerolog/log"
)

// GetEnvOr returns the value of the ENV variable having the given key, or the provided orValue
func GetEnvOr(envKey string, orValue string) string {
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return orValue
}

func GetAccountSequence(cdc codec.Codec, authClient authtypes.QueryClient, address string) (uint64, error) {
	log.Info().Msg(fmt.Sprintf("tester address: %s", address))
	res, err := authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return 0, fmt.Errorf("error while getting account from node")
	}

	err = res.UnpackInterfaces(cdc)
	if err != nil {
		return 0, fmt.Errorf("error while unpacking response")
	}

	account, ok := res.Account.GetCachedValue().(authtypes.AccountI)
	if !ok {
		return 0, fmt.Errorf("error while get account from cached value")
	}

	return account.GetSequence(), nil
}

func generatePostMsgs(subspaceID uint64, size int, author string) []sdk.Msg {
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

func GetGasLimit(wallet *wallet.Wallet, msgs []sdk.Msg) uint64 {
	account, err := wallet.Client.GetAccount(wallet.AccAddress())
	if err != nil {
		panic(err)
	}

	// Build the transaction data for simulation
	txData := wallettypes.NewTransactionData(msgs...).
		WithGasAuto().
		WithFeeAuto().
		WithSequence(account.GetSequence())

	builder, err := wallet.BuildTx(txData)
	if err != nil {
		panic(err)
	}

	gasLimit, err := wallet.Client.SimulateTx(builder.GetTx())
	if err != nil {
		panic(err)
	}

	return gasLimit
}
