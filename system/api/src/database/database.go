package database

import (
	"pkg-common/logger"
	"sync"

	"gorm.io/driver/postgres"
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
		// Mask password in connection string for logging
		masked := connectionString
		if idx := findPasswordIndex(connectionString); idx != -1 {
			masked = connectionString[:idx] + "***" + connectionString[idx+len(getPassword(connectionString)):]
		}
		logger.Default().Infof("Connecting to database with connection string: %s", masked)
		db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
		if err != nil {
			logger.Default().Panicf(err, "Error connecting to database")
		}
		databaseConnection = db
		initializedDatabaseConn = true

		// Run migrations
		if err := AutoMigrate(db); err != nil {
			logger.Default().Panicf(err, "Error running database migrations")
		}
	})
}

// Helper functions to mask password in connection string
func findPasswordIndex(conn string) int {
	pwKey := "password="
	idx := -1
	if i := indexOf(conn, pwKey); i != -1 {
		idx = i + len(pwKey)
	}
	return idx
}

func getPassword(conn string) string {
	pwKey := "password="
	if i := indexOf(conn, pwKey); i != -1 {
		rest := conn[i+len(pwKey):]
		for j, c := range rest {
			if c == ' ' || c == ';' {
				return rest[:j]
			}
		}
		return rest
	}
	return ""
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
