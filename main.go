package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/v6/app"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/desmos-loading/service"
	"github.com/desmos-labs/desmos-loading/utils"
)

const (
	EnvAccountAmount = "ACCOUNT_AMOUNT"
	EnvRound         = "ROUND"
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
	accountAmount, err := strconv.Atoi(utils.GetEnvOr(EnvAccountAmount, ""))
	if err != nil {
		panic(err)
	}

	round, err := strconv.Atoi(utils.GetEnvOr(EnvRound, ""))
	if err != nil {
		panic(err)
	}

	cfg, err := service.ReadEnvConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println("Initializing services...")
	services := make([]*service.Service, accountAmount)
	accounts := make([]string, accountAmount)
	for i := 0; i < accountAmount; i++ {
		services[i] = service.NewService(cfg, getHDPath(i))
		accounts[i] = services[i].Wallet.AccAddress()
	}

	if accountAmount > 1 {
		adjustment := 1.2
		expRoundPerService := int(float64(round) * adjustment / float64(accountAmount))
		services[0].SendFeesToTester(expRoundPerService, accounts[1:])
	}
	time.Sleep(5 * time.Second)

	start(services, round)
}

func start(services []*service.Service, round int) {
	fmt.Println("Saving services profiles...")
	for _, service := range services {
		service.SaveProfileIfNotExist()
	}
	time.Sleep(5 * time.Second)

	startTime := time.Now()
	var wg sync.WaitGroup
	for _, service := range services {
		service := service

		wg.Add(1)
		go func() {
			service.RunTasks(round / len(services))
			wg.Done()
		}()
	}
	fmt.Println("Services running...")

	wg.Wait()
	fmt.Printf("time used: %s\n", time.Now().Sub(startTime))
}

func getHDPath(id int) string {
	return fmt.Sprintf("44'/852'/0'/0/%d", id)
}
