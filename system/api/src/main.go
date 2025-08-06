package main

import (
	"api/src/identity"
	"api/src/queues"
	"api/src/zkp"
	"fmt"
	"net/http"
	"os"
	"pkg-common/logger"
	"pkg-common/utilities"

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
	logger.InitDefaultLogger(logger.GlobalLoggerConfig{
		Args: []struct {
			Key   string
			Value string
		}{
			{"application", "api"},
			{"version", "1.0.0"},
		},
	})
	defaultLogger := logger.Default()

	apiConfig, err := utilities.ReadConfig[ApiConfigJson]("config.json")
	if err != nil {
		defaultLogger.Fatalf(err, "Loading config failed")
	}

	apiLogger := logger.NewFromConfig(apiConfig.LoggerConf)
	// Setup DB
	isDev := utilities.Ternary(os.Getenv("ENV_TYPE") == string(Dev), true, false)

	// Ensure the sqlite directory exists before using it
	os.MkdirAll("./sqlite", 0755)

	dbConn := os.Getenv("DB_CONNECTION_STRING")
	if dbConn == "" {
		dbConn = "./sqlite/DigitalIdentity.db"
	}
	db := database.ConnectToDatabase(dbConn)
	if db == nil {
		apiLogger.Fatal(nil, "Database connection failed")
	}

	if isDev {
		InitializeDev(db)
	}

	// Setup RabbitMQ
	rabbitPublisher, err := queues.NewRabbitPublisher(
		"amqp://guest:guest@rabbitmq:5672/",
		"identity", "identity.verified", "identity.verified",
	)
	if err != nil {
		apiLogger.Fatalf(err, "RabbitMQ setup error")
	}
	defer rabbitPublisher.Close()

	// Ensure the results queue exists
	resultsQueue := "identity.verified.results"
	if err := rabbitPublisher.EnsureResultsQueue(resultsQueue); err != nil {
		apiLogger.Fatalf(err, "Failed to declare results queue")
	}

	rabbitConsumer, err := queues.NewRabbitConsumer(rabbitPublisher.Conn, resultsQueue)
	if err != nil {
		apiLogger.Fatalf(err, "Failed to setup result consumer: %v")
	}

	// Init identityRepo / identityService / identityHandler
	identityHandler, _ := identity.Build(db, rabbitPublisher)
	// Init zkpRepo / zkpService / zkpHandler
	zkpHandler, _ := zkp.Build(db, rabbitConsumer)

	// Gin routes
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	api := r.Group("/api/v1/")
	identity.RegisterIdentityRoutes(api, identityHandler)
	_ = zkpHandler // just to avoid unused var warning, but identityHandler runs in background

	apiLogger.Info("server running at 0.0.0.0:8080")
	r.Run("0.0.0.0:8080")

}
