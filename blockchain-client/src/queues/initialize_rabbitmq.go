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

// ConnectToRabbitmq connects with retries
func ConnectToRabbitmq() (*amqp.Connection, error) {
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

// CreateNewExchange declares an exchange (e.g. "identity", direct)
func CreateNewExchange(ch *amqp.Channel, exName string, exType RabbitmqExchangeType) error {
	return ch.ExchangeDeclare(
		exName,         // name
		string(exType), // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
}

// CreateNewQueue declares a queue with given durability/exclusivity
func CreateNewQueue(ch *amqp.Channel, queueName string, durable, exclusive bool) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName, // name
		durable,   // durable
		false,     // delete when unused
		exclusive, // exclusive
		false,     // no-wait
		nil,       // arguments
	)
}

// BindQueueToExchange binds a queue to an exchange with a routing key
func BindQueueToExchange(ch *amqp.Channel, queueName, routingKey, exchangeName string) error {
	return ch.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
}

// SetupIdentityQueues declares both main and result queues with proper bindings
func SetupIdentityQueues(ch *amqp.Channel) error {
	// Declare the direct exchange
	if err := CreateNewExchange(ch, "identity", ExchangeDirect); err != nil {
		return err
	}
	// Main verification job queue: NOT exclusive, NOT durable (unless you want persistence)
	if _, err := CreateNewQueue(ch, "identity.verified", false, false); err != nil {
		return err
	}
	if err := BindQueueToExchange(ch, "identity.verified", "identity.verified", "identity"); err != nil {
		return err
	}
	// Results queue: NOT exclusive, NOT durable
	if _, err := CreateNewQueue(ch, "identity.verified.results", false, false); err != nil {
		return err
	}
	if err := BindQueueToExchange(ch, "identity.verified.results", "identity.verified.results", "identity"); err != nil {
		return err
	}
	return nil
}
