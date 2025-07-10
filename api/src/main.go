package main

import (
	"api/src/database"
	"api/src/identity"
	"api/src/router"
	"log"
	"os"
)

func main() {
	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnectionString == "" {
		dbConnectionString = "./DigitalIdentity.db"
	}

	db := database.ConnectToDatabase(dbConnectionString)
	if db == nil {
		log.Fatal("Database connection failed")
	}

	// Example: Insert admin if not exists
	admin := identity.SuperIdentity{
		IdentityId:   "admin-guid-here",
		IdentityName: "admin",
	}
	result := db.FirstOrCreate(&admin, identity.SuperIdentity{IdentityName: "admin"})
	if result.Error != nil {
		log.Printf("Error inserting admin: %v", result.Error)
	}

	// Prepare Gin router and register ALL endpoints
	r := router.PrepareAppRouter(db)

	// Start the server
	r.Run(":8080")
}
