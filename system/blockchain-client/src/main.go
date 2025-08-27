package main

import (
	"blockchain-client/src/external"
	"blockchain-client/src/workers"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rest"

	_ "blockchain-client/src/docs"
)

// @title           Digital Identity System - Blockchain Client
// @version         1.0
// @description     API to verify zkSNARK proofs on Solana blockchain
// @host localhost:9000
// @BasePath /bc/v1
func main() {
	appbuilder.New[BlockchainClientConfigJson]().
		InitLogger(logger.GlobalLoggerConfig{}).
		ResolveEnvironment().
		LoadConfig("config.json").
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		AddWorkerServices(
			workers.NewVerifiedPositiveWorker(),
			workers.NewVerifiedNegativeWorker(),
		).
		AddGinMiddleware(
			rest.NewMiddleware("*", rest.InternalAuthMiddleware()),
		).
		AddGinRoutes(
			rest.NewRoute(rest.GET, "v1/internal", "verify", external.NewSolanaReader().Verify),
		).
		AddSwagger().
		InitGinRouter().
		Build().
		Start()
}
