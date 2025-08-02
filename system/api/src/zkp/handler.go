package zkp

import (
	"api/src/model"
	"api/src/queues"
	"encoding/json"
	"pkg-common/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ZeroKnowledgeProofHandler struct {
	service  ZkpService
	consumer *queues.RabbitConsumer
}

func NewZeroKnowledgeProofHandler(service ZkpService, consumer *queues.RabbitConsumer) *ZeroKnowledgeProofHandler {
	h := &ZeroKnowledgeProofHandler{
		service:  service,
		consumer: consumer,
	}
	// Start consuming in background
	go h.listenResultsQueue()
	return h
}

func (h *ZeroKnowledgeProofHandler) listenResultsQueue() {
	zkpLogger := logger.Default()
	err := h.consumer.StartConsume(func(d amqp.Delivery) {
		var resp model.ZeroKnowledgeProofVerificationResponse
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			zkpLogger.Errorf(err, "Failed to unmarshal result")
			return
		}
		// Save to DB, update state, etc
		if err := h.service.ProcessVerificationResult(resp); err != nil {
			zkpLogger.Errorf(err, "Failed to process verification result")
		} else {
			zkpLogger.Infof("Processed ZKP verification result for identity: %s", resp.IdentityId)
			zkpLogger.Infof(
				"ZKP Verification Result: identity_id=%s | is_proof_valid=%v | proof_reference=%s | schema=%s | error=%s",
				resp.IdentityId, resp.IsProofValid, resp.ProofReference, resp.Schema, resp.Error,
			)
		}
	})
	if err != nil {
		zkpLogger.Errorf(err, "Failed to register result queue consumer")
	}
	zkpLogger.Info("Listening for ZKP verification results...")
}
