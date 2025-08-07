package rabbitmq

import (
	"pkg-common/logger"
	"pkg-common/utilities"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerAlias string

var (
	ConsumerRegistry    map[ConsumerAlias]IRabbitmqConsumer
	onceConsumer        sync.Once
	initializedConsumer bool
)

func GetConsumer(alias ConsumerAlias) IRabbitmqConsumer {
	if !initializedConsumer {
		panic("Consumer registry not initialized: call InitializeConsumerRegistry() first")
	}
	return ConsumerRegistry[alias]
}

func InitializeConsumerRegistry(conn *amqp.Connection, consumerConfig []RabbitmqConsumerConfig) {
	onceConsumer.Do(func() {
		ConsumerRegistry = make(map[ConsumerAlias]IRabbitmqConsumer)

		for _, consumer := range consumerConfig {
			channel, err := conn.Channel()
			if err != nil {
				logger.Default().Panicf(err, "Could not obtain connection for consumer")
			}

			ConsumerRegistry[consumer.ConsumerAlias] = NewConsumer(
				channel,
				consumer.QueueName,
				consumer.ConsumerTag,
			)
		}

		initializedConsumer = true
	})
}

type RabbitmqConsumer struct {
	Channel     *amqp.Channel
	QueueName   string
	ConsumerTag string
}

type IRabbitmqConsumer interface {
	StartConsuming(func(amqp.Delivery))
}

func NewConsumer(ch *amqp.Channel, queueName, consumerTag string) *RabbitmqConsumer {
	return &RabbitmqConsumer{
		Channel:     ch,
		QueueName:   queueName,
		ConsumerTag: consumerTag,
	}
}

func (rc *RabbitmqConsumer) StartConsuming(messageHandler func(amqp.Delivery)) {
	defer func() {
		if r := recover(); r != nil {
			logger.Default().Errorf(
				nil,
				"[%s] Recovered from panic for consumer: %s, %v",
				rc.QueueName,
				rc.ConsumerTag,
				r,
			)
		}
	}()

	msgs, err := rc.Channel.Consume(
		rc.QueueName,   // queue
		rc.ConsumerTag, // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	utilities.FailOnError(err, "Failed to register a consumer")

	consumerLogger := logger.Default()
	consumerLogger.Infof("Waiting for messages in queue: %s", rc.QueueName)
	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		for d := range msgs {
			consumerLogger.Infof("[%s] %s", rc.QueueName, d.Body)
			messageHandler(d)
		}
	}()

	waitGroup.Wait()
}
