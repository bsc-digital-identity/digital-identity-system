package zkp

import (
	"api/src/model"
	"api/src/queues"
	"encoding/json"
	"log"

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
	err := h.consumer.StartConsume(func(d amqp.Delivery) {
		var resp model.ZeroKnowledgeProofVerificationResponse
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			log.Printf("Failed to unmarshal result: %v", err)
			return
		}
		// Save to DB, update state, etc
		if err := h.service.ProcessVerificationResult(resp); err != nil {
			log.Printf("Failed to process verification result: %v", err)
		} else {
			log.Printf("Processed ZKP verification result for identity: %s", resp.IdentityId)
			log.Printf(
				"ZKP Verification Result: identity_id=%s | is_proof_valid=%v | proof_reference=%s | schema=%s | error=%s",
				resp.IdentityId, resp.IsProofValid, resp.ProofReference, resp.Schema, resp.Error,
			)
		}
	})
	if err != nil {
		log.Printf("Failed to register result queue consumer: %v", err)
	}
	log.Println("Listening for ZKP verification results...")
}
