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
	// might need to increase on different machines
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

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"verifiers", // name
		"fanout",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)

	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"verified", // name
		false,      // durable
		false,      // delete when unused
		true,       // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,      // queue name
		"",          // routing key
		"verifiers", // exchange
		false,
		nil,
	)

	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
