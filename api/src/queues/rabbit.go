package queues

import (
	"api/src/model"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	Conn       *amqp.Connection
	Channel    *amqp.Channel
	Exchange   string
	Queue      string
	RoutingKey string
}

// NewRabbitPublisher creates a new RabbitPublisher, ensures the exchange and queue exist, and binds them.
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

	// Declare the exchange (direct, durable)
	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Declare the queue (durable)
	if _, err := ch.QueueDeclare(
		queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Bind the queue to the exchange with routing key
	if err := ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitPublisher{
		Conn:       conn,
		Channel:    ch,
		Exchange:   exchange,
		Queue:      queue,
		RoutingKey: routingKey,
	}, nil
}

func (r *RabbitPublisher) EnsureResultsQueue(queueName string) error {
	_, err := r.Channel.QueueDeclare(
		queueName,
		true, false, false, false, nil,
	)
	return err
}

// PublishZkpVerificationRequest publishes a ZKP verification request to the queue.
func (r *RabbitPublisher) PublishZkpVerificationRequest(req model.ZeroKnowledgeProofVerificationRequest) error {
	body, err := json.Marshal(req)
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

// Close closes the AMQP channel and connection.
func (r *RabbitPublisher) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
