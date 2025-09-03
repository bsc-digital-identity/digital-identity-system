package zkpfailed

import (
	"encoding/json"
	dtocommon "pkg-common/dto_common"
	"pkg-common/logger"
	"pkg-common/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	proofFailuresConsumerAlias = "ProofFailuresConsumer"
)

type ZeroKnowledgeProofFailedHandler struct {
	service  ZkpFailedService
	consumer rabbitmq.IRabbitmqConsumer
}

func NewZeroKnowledgeProofFailedHandler() *ZeroKnowledgeProofFailedHandler {
	return &ZeroKnowledgeProofFailedHandler{
		service:  newFailedZkpService(),
		consumer: rabbitmq.GetConsumer(proofFailuresConsumerAlias),
	}
}

func (h *ZeroKnowledgeProofFailedHandler) GetServiceName() string {
	return proofFailuresConsumerAlias
}

func (h *ZeroKnowledgeProofFailedHandler) StartService() {
	zkpLogger := logger.Default()
	h.consumer.StartConsuming(func(d amqp.Delivery) {
		var resp dtocommon.ZkpProofFailureDto
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			zkpLogger.Errorf(err, "Failed to unmarshal result")
			return
		}
		// Save to DB, update state, etc
		if err := h.service.SaveFailedAndUpdateOutbox(resp); err != nil {
			zkpLogger.Errorf(err, "Failed to process verification result")
		}
	})

	zkpLogger.Info("Listening for ZKP verification results...")
}
