package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	wallettypes "github.com/desmos-labs/cosmos-go-wallet/types"
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
	EnvRound    = "ROUND"
)

type Config struct {
	Account    *wallettypes.AccountConfig
	Chain      *wallettypes.ChainConfig
	SubspaceID uint64
	Size       int

	Duration time.Duration
	Round    int
}

// ReadEnvConfig reads a Config instance from the env variables values
func ReadEnvConfig() (*Config, error) {
	subspaceID, err := subspacetypes.ParseSubspaceID(GetEnvOr(EnvSubspaceID, ""))
	if err != nil {
		return nil, err
	}

	size, err := strconv.Atoi(GetEnvOr(EnvSize, ""))
	if err != nil {
		return nil, err
	}

	round, err := strconv.Atoi(GetEnvOr(EnvRound, ""))
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(GetEnvOr(EnvDuration, ""))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Account: &wallettypes.AccountConfig{
			Mnemonic: GetEnvOr(EnvAccountRecoveryPhrase, ""),
			HDPath:   app.FullFundraiserPath,
		},
		Chain: &wallettypes.ChainConfig{
			Bech32Prefix:  app.Bech32MainPrefix,
			RPCAddr:       GetEnvOr(EnvRPCAddress, ""),
			GRPCAddr:      GetEnvOr(EnvGRPCAddress, ""),
			GasPrice:      GetEnvOr(EnvGasPrice, "0.02udaric"),
			GasAdjustment: 2,
		},
		SubspaceID: subspaceID,
		Size:       size,
		Round:      round,
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
