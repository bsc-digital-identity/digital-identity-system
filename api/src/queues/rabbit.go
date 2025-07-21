package queues

import (
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ZkpVerifiedMessage struct {
	IdentityId string `json:"identity_id"`
	BirthDay   int    `json:"birth_day"`
	BirthMonth int    `json:"birth_month"`
	BirthYear  int    `json:"birth_year"`
}

type RabbitPublisher struct {
	Conn       *amqp.Connection
	Channel    *amqp.Channel
	Exchange   string
	Queue      string
	RoutingKey string
}

func NewRabbitPublisher(amqpURL, exchange, queue, routingKey string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Exchange/queue setup (must match blockchain-client)
	_ = ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	_, _ = ch.QueueDeclare(queue, false, false, false, false, nil)
	_ = ch.QueueBind(queue, routingKey, exchange, false, nil)

	return &RabbitPublisher{
		Conn:       conn,
		Channel:    ch,
		Exchange:   exchange,
		Queue:      queue,
		RoutingKey: routingKey,
	}, nil
}

func (r *RabbitPublisher) PublishZkpVerified(msg ZkpVerifiedMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.Channel.Publish(
		r.Exchange,
		r.RoutingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (r *RabbitPublisher) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
