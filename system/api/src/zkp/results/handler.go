package zkpresult

import (
	"encoding/json"
	dtocommon "pkg-common/dto_common"
	"pkg-common/logger"
	"pkg-common/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	proofResultsConsumerAlias = "ProofResultsConsumer"
)

type ZeroKnowledgeProofHandler struct {
	service  ZkpService
	consumer rabbitmq.IRabbitmqConsumer
}

func NewZeroKnowledgeProofHandler() *ZeroKnowledgeProofHandler {
	return &ZeroKnowledgeProofHandler{
		service:  NewZkpService(),
		consumer: rabbitmq.GetConsumer(proofResultsConsumerAlias),
	}
}

func (h *ZeroKnowledgeProofHandler) GetServiceName() string {
	return proofResultsConsumerAlias
}

func (h *ZeroKnowledgeProofHandler) StartService() {
	zkpLogger := logger.Default()
	h.consumer.StartConsuming(func(d amqp.Delivery) {
		var resp dtocommon.ZkpProofResultDto
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			zkpLogger.Errorf(err, "Failed to unmarshal result")
			return
		}
		// Save to DB, update state, etc
		if err := h.service.ProcessVerificationResult(resp); err != nil {
			zkpLogger.Errorf(err, "Failed to process verification result")
		}
	})

	zkpLogger.Info("Listening for ZKP verification results...")
}
