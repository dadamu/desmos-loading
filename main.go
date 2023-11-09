package main

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/v6/app"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/desmos-loading/service"
	"github.com/desmos-labs/desmos-loading/utils"
)

const (
	EnvRound = "ROUND"
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

	round, err := strconv.Atoi(utils.GetEnvOr(EnvRound, ""))
	if err != nil {
		panic(err)
	}

	fmt.Println("Testing service start...")
	startTime := time.Now()
	service := service.NewService(cfg)
	service.RunTasks(round)

	fmt.Println("time used:", time.Now().Sub(startTime))
}
