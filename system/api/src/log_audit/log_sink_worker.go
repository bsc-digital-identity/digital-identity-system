package logaudit

import (
	"encoding/json"
	"os"

	"pkg-common/logger"
	"pkg-common/rabbitmq"
	logger_message "pkg-common/utilities/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	logConsumerAlias = "LogConsumer"
)

type LogSinkWorker struct {
	service  LogAuditService
	consumer rabbitmq.IRabbitmqConsumer
	logger   *logger.Logger
}

func NewLogSinkWorker() *LogSinkWorker {
	dedicatedLogger := logger.New().WithOutput(os.Stdout)

	repository := NewLogAuditRepository()
	service := NewLogAuditService(repository)

	return &LogSinkWorker{
		service:  service,
		consumer: rabbitmq.GetConsumer(rabbitmq.ConsumerAlias(logConsumerAlias)),
		logger:   dedicatedLogger,
	}
}

func (w *LogSinkWorker) GetServiceName() string {
	return logConsumerAlias
}

func (w *LogSinkWorker) StartService() {
	w.logger.Info("Starting API Log Sink Worker")

	w.consumer.StartConsuming(func(d amqp.Delivery) {
		var logMessage logger_message.LoggerMessage

		if err := json.Unmarshal(d.Body, &logMessage); err != nil {
			w.logger.Errorf(err, "Failed to unmarshal log message")
			return
		}

		w.logger.Debugf("Processing log message: Level=%s, Message=%s", logMessage.Level, logMessage.Message)

		if err := w.service.ProcessLogMessage(logMessage); err != nil {
			w.logger.Errorf(err, "Failed to save log message to database")
			return
		}

		w.logger.Debugf("Successfully saved log message to database: %s", logMessage.Message)
	})
}
