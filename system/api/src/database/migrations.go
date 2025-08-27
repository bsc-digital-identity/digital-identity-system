package database

import (
	"api/src/model"
	"pkg-common/logger"
)

func RunMigrations() {
	db := GetDatabaseConnection()
	migtionLogger := logger.Default()
	migtionLogger.Info("Running migrations for tables... ")

	err := db.AutoMigrate(
		&model.Identity{},
		&model.VerifiedSchema{},
		&model.ZeroKnowledgeProof{})
	if err != nil {
		migtionLogger.Fatal(err, "Migrating database failed")
	}

	migtionLogger.Info("All tables created (or already exist).")
}
