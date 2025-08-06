package rabbitmq

import (
	"encoding/json"
	"pkg-common/logger"
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
	onceConsumer.Do(func() {
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
	Publish(body any) error
}

func (rp *RabbitmqPublisher) Publish(body any) error {
	body_json, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return rp.Channel.Publish(
		rp.Exchange,
		rp.RoutingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body_json,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}
