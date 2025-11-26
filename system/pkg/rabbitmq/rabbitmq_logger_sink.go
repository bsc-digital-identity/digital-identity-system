package rabbitmq

import (
	"fmt"
	logger_message "pkg-common/utilities/logger"
	"pkg-common/utilities/timeutil"

	"github.com/rs/zerolog"
)

func CreateRabbitmqLoggerSink(publisher IRabbitmqPublisher) func(string, zerolog.Level, timeutil.TimeUTC) {
	return func(msg string, level zerolog.Level, timestamp timeutil.TimeUTC) {
		loggerMessage := logger_message.LoggerMessage{
			Level:     level.String(),
			Message:   msg,
			Timestamp: timestamp,
		}

		err := publisher.Publish(loggerMessage)
		if err != nil {
			// Avoid infinite recursion by not using the logger here
			fmt.Printf("Failed to publish log message to RabbitMQ: %v\n", err)
		}
	}
}
