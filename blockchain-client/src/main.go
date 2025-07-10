package main

import (
	"log"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatal("%s: %s", msg, err)
	}
}

func connectWithRetry() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	maxRetries := 7
	waitTime := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			return conn, nil
		}

		log.Printf("Attempt %d failed: %v. Retrying in %v...", i+1, err, waitTime)
		time.Sleep(waitTime)

		waitTime = time.Duration(math.Pow(2, float64(i+1))) * time.Second
	}

	return nil, err
}

func main() {
	conn, err := connectWithRetry()
	failOnError(err, "Failed to connect to RabbitMQ after retries")
	defer conn.Close()

	log.Println("Successfully connected to RabbitMQ")
}
