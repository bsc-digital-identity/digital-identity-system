package queues

import (
	"math"
	"pkg-common/logger"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqExchangeType string

func (ret RabbitmqExchangeType) String() string {
	return string(ret)
}

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

	queueLogger := logger.Default()

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			return conn, nil
		}
		queueLogger.Warnf("Attempt %d failed: %v. Retrying in %v...", i+1, err, waitTime)
		time.Sleep(waitTime)
		waitTime = time.Duration(math.Pow(2, float64(i+1))) * time.Second
	}
	return nil, err
}

// CreateNewExchange declares an exchange (e.g. "identity", direct)
func CreateNewExchange(ch *amqp.Channel, exchangeConfig RabbitmqExchangeConfig) error {
	return ch.ExchangeDeclare(
		exchangeConfig.ExchangeName,          // name
		exchangeConfig.ExchangeType.String(), // type
		true,                                 // durable
		false,                                // auto-deleted
		false,                                // internal
		false,                                // no-wait
		nil,                                  // arguments
	)
}

// CreateNewQueue declares a queue with given durability/exclusivity
func CreateNewQueue(ch *amqp.Channel, queueConfig RabbitmqQueueConfig) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueConfig.QueueName, // name
		queueConfig.Durable,   // durable
		false,                 // delete when unused
		queueConfig.Exclusive, // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
}

// BindQueueToExchange binds a queue to an exchange with a routing key
func BindQueueToExchange(ch *amqp.Channel, queueConfig RabbitmqQueueConfig) error {
	return ch.QueueBind(
		queueConfig.QueueName,       // queue name
		queueConfig.RoutingKey,      // routing key
		queueConfig.ExchangeBinding, // exchange
		false,
		nil,
	)
}

// SetupIdentityQueues declares both main and result queues with proper bindings
func SetupIdentityQueues(ch *amqp.Channel, rabbimqConfig RabbitmqConfig) error {
	// declare exchanges
	for _, exchangeConf := range rabbimqConfig.Exchanges {
		if err := CreateNewExchange(ch, exchangeConf); err != nil {
			return err
		}
	}

	// declare queues and bind to exchanges
	for _, queueConf := range rabbimqConfig.Queues {
		if _, err := CreateNewQueue(ch, queueConf); err != nil {
			return err
		}

		if err := BindQueueToExchange(ch, queueConf); err != nil {
			return err
		}
	}

	return nil
}
