package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	wallettypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/desmos-loading/utils"
	"github.com/desmos-labs/desmos/v6/app"
	subspacetypes "github.com/desmos-labs/desmos/v6/x/subspaces/types"
)

const (
	EnvAccountRecoveryPhrase = "ACCOUNT_RECOVERY_PHRASE"
	EnvRPCAddress            = "RPC_ADDRESS"
	EnvGRPCAddress           = "GRPC_ADDRESS"
	EnvGasPrice              = "GAS_PRICE"

	EnvSubspaceID = "SUBSPACE_ID"
	EnvSize       = "MSG_SIZE"

	EnvDuration = "DURATION"
)

type Config struct {
	Account    *wallettypes.AccountConfig
	Chain      *wallettypes.ChainConfig
	SubspaceID uint64
	Size       int

	Duration time.Duration
}

// ReadEnvConfig reads a Config instance from the env variables values
func ReadEnvConfig() (*Config, error) {
	subspaceID, err := subspacetypes.ParseSubspaceID(utils.GetEnvOr(EnvSubspaceID, ""))
	if err != nil {
		return nil, err
	}

	size, err := strconv.Atoi(utils.GetEnvOr(EnvSize, ""))
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(utils.GetEnvOr(EnvDuration, ""))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Account: &wallettypes.AccountConfig{
			Mnemonic: utils.GetEnvOr(EnvAccountRecoveryPhrase, ""),
			HDPath:   app.FullFundraiserPath,
		},
		Chain: &wallettypes.ChainConfig{
			Bech32Prefix:  app.Bech32MainPrefix,
			RPCAddr:       utils.GetEnvOr(EnvRPCAddress, ""),
			GRPCAddr:      utils.GetEnvOr(EnvGRPCAddress, ""),
			GasPrice:      utils.GetEnvOr(EnvGasPrice, "0.02udaric"),
			GasAdjustment: 2,
		},
		SubspaceID: subspaceID,
		Size:       size,
		Duration:   duration,
	}
	return cfg, cfg.Validate()
}

// Validate validates the given configuration returning any error
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Account.Mnemonic) == "" {
		return fmt.Errorf("missing account mnemonic")
	}

	return nil
}
