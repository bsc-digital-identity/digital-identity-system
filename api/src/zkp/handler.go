package zkp

import (
	"api/src/model"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type ZeroKnowledgeProofHandler struct {
	service      ZkpService
	amqpChannel  *amqp.Channel
	resultsQueue string
}

func NewZeroKnowledgeProofHandler(service ZkpService, amqpChannel *amqp.Channel, resultsQueue string) *ZeroKnowledgeProofHandler {
	h := &ZeroKnowledgeProofHandler{
		service:      service,
		amqpChannel:  amqpChannel,
		resultsQueue: resultsQueue,
	}
	go h.listenResultsQueue() // start listener in background
	return h
}

// Listen for verification results from the queue
func (h *ZeroKnowledgeProofHandler) listenResultsQueue() {
	msgs, err := h.amqpChannel.Consume(
		h.resultsQueue,
		"zkp-result-consumer",
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("Failed to register result queue consumer: %v", err)
		return
	}
	log.Println("Listening for ZKP verification results...")
	for d := range msgs {
		// Handle each result message here
		var resp model.ZeroKnowledgeProofVerificationResponse
		if err := json.Unmarshal(d.Body, &resp); err != nil {
			log.Printf("Failed to unmarshal result: %v", err)
			continue
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
	}
}
