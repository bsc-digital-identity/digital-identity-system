package queues

import (
	"log"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqExchangeType string

const (
	ExchangeFanout  RabbitmqExchangeType = "fanout"
	ExchangeDirect  RabbitmqExchangeType = "direct"
	ExchangeTopic   RabbitmqExchangeType = "topic"
	ExchangeHeaders RabbitmqExchangeType = "headers"
)

func ConnectToRabbitmq() (*amqp.Connection, error) {
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

func CreateNewExchange(ch *amqp.Channel, ex_name string, ex_type RabbitmqExchangeType) error {
	return ch.ExchangeDeclare(
		ex_name,         // name
		string(ex_type), // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
}

func CreateNewQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
}

func BindQueueToExchange(ch *amqp.Channel, queueName, routingKey, exchangeName string) error {
	return ch.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
}
