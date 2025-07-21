package main

import (
	"api/src/identity"
	"api/src/queues"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"

	"api/src/database"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from Go Docker multistage")
}

func main() {
	// Setup DB
	dbConn := os.Getenv("DB_CONNECTION_STRING")
	if dbConn == "" {
		dbConn = "./DigitalIdentity.db"
	}
	db := database.ConnectToDatabase(dbConn)
	if db == nil {
		log.Fatal("Database connection failed")
	}

	// Setup RabbitMQ
	rabbit, err := queues.NewRabbitPublisher(
		"amqp://guest:guest@rabbitmq:5672/",
		"identity", "identity.verified", "identity.verified",
	)
	if err != nil {
		log.Fatalf("RabbitMQ setup error: %v", err)
	}
	defer rabbit.Close()

	// Example: Insert admin if not exists
	admin := identity.SuperIdentity{
		IdentityId:   "admin-guid-here",
		IdentityName: "admin",
	}
	result := db.FirstOrCreate(&admin, identity.SuperIdentity{IdentityName: "admin"})
	if result.Error != nil {
		log.Printf("Error inserting admin: %v", result.Error)
	}

	// Init service and handler
	service := identity.NewService(db, rabbit)
	handler := identity.NewHandler(service)

	// Gin routes
	r := gin.Default()
	api := r.Group("/identity")
	identity.RegisterIdentityRoutes(api, handler)

	log.Println("server running at 0.0.0.0:8080")
	r.Run("0.0.0.0:8080")

}
