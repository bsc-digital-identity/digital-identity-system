package outbox

import (
	"pkg-common/logger"
	"pkg-common/rabbitmq"

	"github.com/robfig/cron"
)

const outboxWorkerName = "OutboxCronWorker"

type OutboxWorker struct {
	publisher  rabbitmq.IRabbitmqPublisher
	repository OutboxRepository
	cron       *cron.Cron
}

func NewOutboxWorker() rabbitmq.WorkerService {
	return &OutboxWorker{
		publisher:  rabbitmq.GetPublisher("VerifiersUnverifiedPublisher"),
		repository: NewRepo(),
		cron:       cron.New(),
	}
}

func (ow *OutboxWorker) GetServiceName() string {
	return outboxWorkerName
}

func (ow *OutboxWorker) StartService() {
	// change based on env for dev minute for prod hour
	err := ow.cron.AddFunc("@every 1m", func() { ow.processOutboxEvents() })
	if err != nil {
		logger.Default().Errorf(err, "Could not add function to %s", outboxWorkerName)
	}

	ow.cron.Start()
}

func (ow *OutboxWorker) processOutboxEvents() {
	outboxLogger := logger.Default()

	events, err := ow.repository.GetUnprocessedEvents()
	if err != nil {
		outboxLogger.Error(err, "Could not read events from database")
	}

	for _, e := range events {
		err := ow.publisher.Publish(e.MapToZkpVerifcationRequest())
		if err != nil {
			outboxLogger.Error(err, "Can't publish to queue")
		}
	}
}
