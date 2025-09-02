package zkpfailed

import (
	"api/src/model"
	"encoding/json"
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
		var resp model.ZeroKnowledgeProofVerificationResponse
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			zkpLogger.Errorf(err, "Failed to unmarshal result")
			return
		}
		// Save to DB, update state, etc
		if err := h.service.Todo(resp); err != nil {
			zkpLogger.Errorf(err, "Failed to process verification result")
		} else {
			zkpLogger.Infof("Processed ZKP verification result for identity: %s", resp.IdentityId)
			zkpLogger.Infof(
				"ZKP Verification Result: identity_id=%s | is_proof_valid=%v | proof_reference=%s | schema=%s | error=%s",
				resp.IdentityId, resp.IsProofValid, resp.ProofReference, resp.Schema, resp.Error,
			)
		}
	})

	zkpLogger.Info("Listening for ZKP verification results...")
}
