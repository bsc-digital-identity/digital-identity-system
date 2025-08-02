package database

import (
	"api/src/model"
	"pkg-common/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectToDatabase(connectionString string) *gorm.DB {
	defaultLogger := logger.New()

	defaultLogger.Infof("Establishing connection to development database: %s", connectionString)

	db, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		defaultLogger.Fatal(err, "Cannot establish database connection")
	}

	defaultLogger.Info("Running migrations for tables")
	err = db.AutoMigrate(
		&model.Identity{},
		&model.VerifiedSchema{},
		&model.ZeroKnowledgeProof{})
	if err != nil {
		defaultLogger.Fatal(err, "Migrating database failed")
	}

	defaultLogger.Info("All tables created (or already exist).")
	return db
}
