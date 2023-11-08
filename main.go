package main

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/v6/app"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/desmos-loading/service"
)

func init() {
	// load env
	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Error loading .env file")
	}

	// Setup Cosmos-related stuff
	app.SetupConfig(sdk.GetConfig())
}

func main() {
	cfg, err := service.ReadEnvConfig()
	if err != nil {
		panic(err)
	}

	service := service.NewService(cfg)
	service.RunTasks()
}
