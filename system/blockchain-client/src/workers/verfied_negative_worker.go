package workers

import (
	"blockchain-client/src/types/incoming"
	"encoding/json"
	dtocommon "pkg-common/dto_common"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	reasoncodes "pkg-common/reason_codes"

	amqp "github.com/rabbitmq/amqp091-go"
)

type VerifiedNegativeWorker struct {
	Consumer rabbitmq.IRabbitmqConsumer
}

func NewVerifiedNegativeWorker() rabbitmq.WorkerService {
	return &VerifiedNegativeWorker{
		Consumer: rabbitmq.GetConsumer(verifiedNegativeConsumerAlias),
	}
}

func (vnw *VerifiedNegativeWorker) GetServiceName() string {
	return verifiedNegativeConsumerAlias
}

func (vnw *VerifiedNegativeWorker) StartService() {
	workerLogger := logger.Default()
	failurePublisher := rabbitmq.GetPublisher(failureQueuePublisherAlias)

	vnw.Consumer.StartConsuming(func(d amqp.Delivery) {
		var message incoming.ZkpVerifiedNegativeDto
		responseFactory := dtocommon.NewZkpProofFailureFactory("", d.Body)

		if err := json.Unmarshal(d.Body, &message); err != nil {
			workerLogger.Error(err, "Unmarshaling failed.")

			result := responseFactory.CreateErrorDto(err, reasoncodes.ErrUnmarshal)

			_ = failurePublisher.Publish(result)
			workerLogger.Info("Message processed sucessfully.")
			return
		}

		responseFactory = dtocommon.NewZkpProofFailureFactory(message.EventId, d.Body)
		_ = failurePublisher.Publish(responseFactory.CreateInfoDto(reasoncodes.ErrVerifierResolution))

		workerLogger.Info("Message processed sucessfully.")
	})
}
