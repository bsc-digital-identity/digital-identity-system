package main

import (
	"api/src/database"
	"api/src/identity"
	"api/src/zkp"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rest"
)

func main() {
	appbuilder.New[ApiConfigJson, ApiConfig]().
		InitLogger(logger.GlobalLoggerConfig{}).
		LoadConfig("config.json").
		WithOption(func(a *appbuilder.AppBuilder[ApiConfigJson, ApiConfig]) {
			database.ConnectToDatabase(a)
			database.RunMigrations(true)
		}).
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		AddWorkerServices(zkp.NewZeroKnowledgeProofHandler()).
		AddGinRoutes(
			rest.NewRoute(rest.POST, "v1", "", identity.NewHandler().CreateIdentity),
			rest.NewRoute(rest.GET, "v1", "/:id", identity.NewHandler().GetIdentity),
			rest.NewRoute(rest.POST, "v1", "/verify", identity.NewHandler().QueueVerification),
		).
		AddSwagger().
		InitGinRouter().
		Build().
		Start()
}
