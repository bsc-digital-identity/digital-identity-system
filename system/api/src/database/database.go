package database

import (
	"pkg-common/logger"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	databaseConnection      *gorm.DB
	onceConnectDb           sync.Once
	initializedDatabaseConn bool
)

func GetDatabaseConnection() *gorm.DB {
	if !initializedDatabaseConn {
		panic("Database connection is not established: call InitializeDatabaseConnection() first")
	}
	return databaseConnection
}

func InitializeDatabaseConnection(connectionString string) {
	onceConnectDb.Do(func() {
		db, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
		if err != nil {
			logger.Default().Panicf(err, "Erorr connecting to database")
		}

		databaseConnection = db
		initializedDatabaseConn = true
	})
}
