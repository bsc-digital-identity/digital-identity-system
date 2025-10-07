package database

import (
	appbuilder "pkg-common/app_builder"
	"pkg-common/utilities"
)

type DatabaseConfig interface {
	appbuilder.AppConfig
	GetDatabaseConnectionString() string
}

func ConnectToDatabase[T utilities.JsonConfigObj[U], U DatabaseConfig](a *appbuilder.AppBuilder[T, U]) {
	a.Logger.Info("Establishing connection to database...")
	connectionString := a.Config.GetDatabaseConnectionString()

	InitializeDatabaseConnection(connectionString)

	a.Logger.Info("Database connection established successfully.")
}

func RunMigrations(migrateDatabase bool) {
	if migrateDatabase {
		if err := AutoMigrate(GetDatabaseConnection()); err != nil {
			panic(err)
		}
	}
}
