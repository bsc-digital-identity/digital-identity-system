package rabbitmq

import (
	"pkg-common/logger"
	"pkg-common/utilities"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherAlias string

var (
	PublisherRegistry map[PublisherAlias]IRabbitmqPublisher
	oncePublisher     sync.Once
)

func GetPublisher(alias PublisherAlias) IRabbitmqPublisher {
	return PublisherRegistry[alias]
}

func InitializePublisherRegistry(conn *amqp.Connection, publisherConfig []RabbitmqPublishersConfig) {
	oncePublisher.Do(func() {
		PublisherRegistry = make(map[PublisherAlias]IRabbitmqPublisher)

		for _, publisher := range publisherConfig {
			channel, err := conn.Channel()
			if err != nil {
				logger.Default().Panicf(err, "Could not obtain connection for publisher")
			}

			PublisherRegistry[publisher.PublisherAlias] = NewPublisher(
				channel,
				publisher.Exchange,
				publisher.RoutingKey,
			)
		}
	})
}

type RabbitmqPublisher struct {
	Channel    *amqp.Channel
	Exchange   string
	RoutingKey string
}

func NewPublisher(ch *amqp.Channel, exchange, routingKey string) *RabbitmqPublisher {
	return &RabbitmqPublisher{
		Channel:    ch,
		Exchange:   exchange,
		RoutingKey: routingKey,
	}
}

type IRabbitmqPublisher interface {
	Publish(body utilities.Serializable) error
}

func (rp *RabbitmqPublisher) Publish(body utilities.Serializable) error {
	json, err := body.Serialize()
	if err != nil {
		return err
	}

	return rp.Channel.Publish(
		rp.Exchange,
		rp.RoutingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         json,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}
