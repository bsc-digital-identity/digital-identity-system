package logaudit

import (
	"encoding/json"

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

	repository := NewLogAuditRepository()
	service := NewLogAuditService(repository)

	return &LogSinkWorker{
		service:  service,
		consumer: rabbitmq.GetConsumer(rabbitmq.ConsumerAlias(logConsumerAlias)),
	}
}

func (w *LogSinkWorker) GetServiceName() string {
	return logConsumerAlias
}

func (w *LogSinkWorker) StartService() {
	w.consumer.StartConsuming(func(d amqp.Delivery) {
		var logMessage logger_message.LoggerMessage

		if err := json.Unmarshal(d.Body, &logMessage); err != nil {
			return
		}

		if err := w.service.ProcessLogMessage(logMessage); err != nil {
			return
		}
	})
}
