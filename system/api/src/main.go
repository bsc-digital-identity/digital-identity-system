package main

import (
	builderextensions "api/src/builder_extensions"
	appbuilder "pkg-common/app_builder"
	"pkg-common/logger"
)

func main() {
	appbuilder.New[ApiConfigJson, ApiConfig]().
		InitLogger(logger.GlobalLoggerConfig{}).
		LoadConfig("config.json").
		WithOption(func(a *appbuilder.AppBuilder[ApiConfigJson, ApiConfig]) {
			builderextensions.ConnectToDatabase(a)
			builderextensions.RunMigrations(true)
		}).
		InitRabbitmqConnection().
		InitRabbitmqRegistries().
		AddWorkerServices().
		AddGinRoutes().
		AddSwagger().
		InitGinRouter().
		Build().
		Start()
}
