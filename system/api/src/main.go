package main

import (
	"api/src/database"
	"api/src/identity"
	"api/src/middleware"
	"api/src/zkp"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
	"pkg-common/rest"

	_ "api/src/docs"
)

// @title           Digital Identity System API
// @version         1.0
// @description     API to manage identities and verify ZKP proofs
// @host localhost:9000
// @BasePath /v1
func main() {
	appbuilder.New[ApiConfigJson, ApiConfig]().
		InitLogger(logger.GlobalLoggerConfig{}).
		ResolveEnvironment().
		LoadConfig("config.json").
		WithOption(func(a *appbuilder.AppBuilder[ApiConfigJson, ApiConfig]) {
			database.ConnectToDatabase(a)
			database.RunMigrations(true)
		}).
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		AddWorkerServices(zkp.NewZeroKnowledgeProofHandler()).
		AddGinMiddleware(
			rest.NewMiddleware("v1", middleware.PublicAuthMiddleware()),
			rest.NewMiddleware("v1/internal", rest.InternalAuthMiddleware()),
		).
		AddGinRoutes(
			rest.NewRoute(rest.POST, "v1", "identity", identity.NewHandler().CreateIdentity),
			rest.NewRoute(rest.GET, "v1", "identity/:id", identity.NewHandler().GetIdentity),
			rest.NewRoute(rest.POST, "v1", "identity/verify", identity.NewHandler().QueueVerification),
		).
		AddSwagger().
		InitGinRouter().
		Build().
		Start()
}
