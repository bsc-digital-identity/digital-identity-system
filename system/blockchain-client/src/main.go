package main

import (
	"blockchain-client/src/external"
	"blockchain-client/src/workers"
	"fmt"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	"pkg-common/rest"
	"pkg-common/utilities"

	"blockchain-client/src/docs"
)

// @title           Digital Identity System - Blockchain Client
// @version         1.0
// @description     API to verify zkSNARK proofs on Solana blockchain
// @host localhost:9000
// @BasePath /bc/v1
func main() {
	lanHost := utilities.ResolveLanHost()
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:9000", lanHost)

	appbuilder.New[BlockchainClientConfigJson]().
		InitLogger(logger.GlobalLoggerConfig{}).
		ResolveEnvironment().
		LoadConfig("config.json").
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		WithOption(func(a *appbuilder.AppBuilder[BlockchainClientConfigJson, BlockchainClientConfig]) {
			// ----- RABBITMQ LOGGING SINK -----
			logPublisher := rabbitmq.GetPublisher("LogPublisher")
			loggerInstance := logger.Default()
			logSink := rabbitmq.CreateRabbitmqLoggerSink(logPublisher)
			logger.AddSinkToLoggerInstance(loggerInstance, logSink)
		}).
		AddWorkerServices(
			workers.NewVerifiedPositiveWorker(),
			workers.NewVerifiedNegativeWorker(),
			workers.NewBlockchainClientLogSink(),
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
