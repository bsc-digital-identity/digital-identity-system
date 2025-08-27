package workers

import (
	"blockchain-client/src/types/incoming"
	"encoding/json"
	"fmt"
	dtocommon "pkg-common/dto_common"
	"pkg-common/logger"
	"pkg-common/rabbitmq"

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
		err := json.Unmarshal(d.Body, &message)
		if err != nil {
			result := dtocommon.ZkpProofFailureDto{
				IdentityId: message.IdentityId,
				Schema:     message.SchemaId,
				ReqestBody: d.Body,
				Error:      "unmarshal: " + err.Error(),
			}

			_ = failurePublisher.Publish(result)
		}

		workerLogger.Error(fmt.Errorf("Something went wrong"), "Failed to publish messgae")
	})
}
