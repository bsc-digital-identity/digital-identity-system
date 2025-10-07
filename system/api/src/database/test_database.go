package database

import (
	"os"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	testDb     *gorm.DB
	testDbOnce sync.Once
)

// GetTestDB returns a database connection for testing
func GetTestDB(t *testing.T) *gorm.DB {
	testDbOnce.Do(func() {
		// Use environment variable for test database if provided
		dbConnString := os.Getenv("TEST_DB_CONNECTION_STRING")
		if dbConnString == "" {
			// Default test database connection string
			dbConnString = "host=localhost user=api_user password=api_password dbname=digital_identity_test port=5432 sslmode=disable"
		}

		db, err := gorm.Open(postgres.Open(dbConnString), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to test database: %v", err)
		}
		testDb = db
	})
	return testDb
}

// SetupTestDB prepares the test database
func SetupTestDB(t *testing.T) *gorm.DB {
	db := GetTestDB(t)

	// Clean up existing tables
	err := db.Exec(`DROP SCHEMA public CASCADE;
					CREATE SCHEMA public;
					GRANT ALL ON SCHEMA public TO api_user;
					GRANT ALL ON SCHEMA public TO public;`).Error
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Run migrations
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database after tests
func CleanupTestDB(t *testing.T) {
	if testDb != nil {
		sqlDB, err := testDb.DB()
		if err != nil {
			t.Errorf("Failed to get underlying *sql.DB: %v", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			t.Errorf("Failed to close database connection: %v", err)
		}
	}
}
