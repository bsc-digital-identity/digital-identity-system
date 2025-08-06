package rabbitmq

import (
	"fmt"
	"math"
	"pkg-common/logger"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectToRabbitmq(user, password string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	maxRetries := 7
	waitTime := 1 * time.Second

	queueLogger := logger.Default()

	for i := 0; i < maxRetries; i++ {
		connectionString := fmt.Sprintf("amqp://%s:%s@rabbitmq:5672/", user, password)
		conn, err = amqp.Dial(connectionString)
		if err == nil {
			return conn, nil
		}
		queueLogger.Warnf("Attempt %d failed: %v. Retrying in %v...", i+1, err, waitTime)
		time.Sleep(waitTime)
		waitTime = time.Duration(math.Pow(2, float64(i+1))) * time.Second
	}
	return nil, err
}
