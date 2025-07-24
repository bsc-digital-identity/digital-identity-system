package main

import (
	"api/src/identity"
	"api/src/queues"
	"api/src/zkp"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"api/src/database"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from Go Docker multistage")
}

type EnvType string

const (
	Dev  EnvType = "dev"
	Prod EnvType = "prod"
)

func main() {
	// Setup DB
	isDev := Ternary(os.Getenv("ENV_TYPE") == string(Dev), true, false)

	// Ensure the sqlite directory exists before using it
	os.MkdirAll("./sqlite", 0755)

	dbConn := os.Getenv("DB_CONNECTION_STRING")
	if dbConn == "" {
		dbConn = "./sqlite/DigitalIdentity.db"
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

	// Ensure the results queue exists
	resultsQueue := "identity.verified.results"
	if err := rabbit.EnsureResultsQueue(resultsQueue); err != nil {
		log.Fatalf("Failed to declare results queue: %v", err)
	}

	if isDev {
		InitializeDev(db)
	}

	// Init service and handler
	service := identity.NewService(db, rabbit)
	handler := identity.NewHandler(service)

	// Gin routes
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	api := r.Group("/api/v1/")
	identity.RegisterIdentityRoutes(api, handler)

	resultsChannel, err := rabbit.Conn.Channel()

	zkpService := zkp.NewZkpService(db)
	zkpHandler := zkp.NewZkpHandler(zkpService, resultsChannel, "identity.verified.results")
	_ = zkpHandler // just to avoid unused var warning, but handler runs in background

	log.Println("server running at 0.0.0.0:8080")
	r.Run("0.0.0.0:8080")

}
