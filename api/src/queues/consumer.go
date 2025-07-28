package queues

import amqp "github.com/rabbitmq/amqp091-go"

type RabbitConsumer struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	QueueName string
}

func NewRabbitConsumer(conn *amqp.Connection, queueName string) (*RabbitConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &RabbitConsumer{Conn: conn, Channel: ch, QueueName: queueName}, nil
}

func (c *RabbitConsumer) StartConsume(handler func(amqp.Delivery)) error {
	msgs, err := c.Channel.Consume(c.QueueName, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			handler(d)
		}
	}()
	return nil
}
