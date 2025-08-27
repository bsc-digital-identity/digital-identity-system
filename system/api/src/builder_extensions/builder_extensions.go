package builderextensions

import (
	"api/src/database"
	appbuilder "pkg-common/app_builder"
	"pkg-common/utilities"
)

type AppConfig interface {
	appbuilder.AppConfig
	GetDatabaseConnectionString() string
}

func ConnectToDatabase[T utilities.JsonConfigObj[U], U AppConfig](a *appbuilder.AppBuilder[T, U]) {
	a.Logger.Info("Establishing connection to database...")
	connectionString := a.Config.GetDatabaseConnectionString()

	database.InitializeDatabaseConnection(connectionString)

	a.Logger.Info("Database connection established successfully.")
}

func RunMigrations(runMigrations bool) {
	if runMigrations {
		database.RunMigrations()
	}
}
