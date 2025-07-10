package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"api/src/database"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from Go Docker multistage")
}

func main() {
	// Get DB connection string, default to ./DigitalIdentity.db
	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	if dbConnectionString == "" {
		dbConnectionString = "./DigitalIdentity.db"
	}

	// Connect to database (using your database package with GORM)
	db := database.ConnectToDatabase(dbConnectionString)
	if db == nil {
		log.Fatal("Database connection failed")
	}

	// Example: Insert admin if not exists
	admin := database.SuperIdentity{
		IdentityId:   "admin-guid-here",
		IdentityName: "admin",
	}
	result := db.FirstOrCreate(&admin, database.SuperIdentity{IdentityName: "admin"})
	if result.Error != nil {
		log.Printf("Error inserting admin: %v", result.Error)
	}

	http.HandleFunc("/", handler)
	fmt.Println("server running at 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
