package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectToDatabase(connectionString string) *gorm.DB {
	log.Printf("Establising connection to development database: %s", connectionString)

	db, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Println("Cannot establish database connection")
		return nil
	}

	log.Println("Running migrations tables")
	err = db.AutoMigrate(&SuperIdentity{}, &Attribute{}, &ZKPProof{})
	if err != nil {
		log.Fatal("Migrating database failed")
	}

	if !db.Migrator().HasTable(&SuperIdentity{}) {
		log.Fatal("SuperIdentity table was not created")
	} else {
		log.Println("SuperIdentity table exists")
	}

	return db
}
