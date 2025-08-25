package main

import (
	"blockchain-client/src/external"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rest"
)

func main() {
	appbuilder.New[BlockchainClientConfigJson]().
		InitLogger(logger.GlobalLoggerConfig{}).
		LoadConfig("config.json").
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		AddWorkerServices(external.NewSolanaClient()).
		AddGinRoutes(rest.NewRoute(
			rest.GET,
			"v1",
			"verify",
			external.NewSolanaReader().Verify,
		)).
		InitGinRouter().
		Build().
		Start()
}
