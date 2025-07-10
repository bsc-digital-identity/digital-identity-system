package database

import (
	"log"

	"api/src/attribute"
	"api/src/identity"
	"api/src/zkp"

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
	err = db.AutoMigrate(&identity.SuperIdentity{}, &attribute.Attribute{}, &zkp.ZKPProof{})
	if err != nil {
		log.Fatal("Migrating database failed: ", err)
	}

	log.Println("All tables created (or already exist).")
	return db
}
