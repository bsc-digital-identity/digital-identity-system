package database

import (
	"api/src/model"
	"pkg-common/logger"

	"gorm.io/gorm"
)

// AutoMigrate runs all database migrations
func AutoMigrate(db *gorm.DB) error {
	migrationLogger := logger.Default()
	migrationLogger.Info("Running migrations for tables... ")

	// Define the order of migrations to handle dependencies
	models := []interface{}{
		&model.Identity{},
		&model.VerifiedSchema{},
		&model.ZeroKnowledgeProof{},
		&model.ZkpProofFailure{},
		&model.OutboxEvent{},
	}

	// Run migrations in order
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			migrationLogger.Errorf(err, "Failed migrating %T", model)
			return err
		}
		migrationLogger.Infof("Migrated %T", model)
	}

	migrationLogger.Info("All tables created (or already exist).")
	return nil
}
