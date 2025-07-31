package database

import (
	"api/src/model"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectToDatabase(connectionString string) *gorm.DB {
	log.Printf("Establishing connection to development database: %s", connectionString)

	db, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Fatal("Cannot establish database connection: ", err)
	}

	log.Println("Running migrations for tables")
	err = db.AutoMigrate(
		&model.Identity{},
		&model.VerifiedSchema{},
		&model.ZeroKnowledgeProof{})
	if err != nil {
		log.Fatal("Migrating database failed: ", err)
	}

	log.Println("All tables created (or already exist).")
	return db
}
