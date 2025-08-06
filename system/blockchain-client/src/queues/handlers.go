package queues

import (
	"encoding/json"
	"time"

	"github.com/gagliardetto/solana-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ZeroKnowledgeProofVerificationRequest struct {
	IdentityId string `json:"identity_id"`
	Schema     string `json:"schema"` // schema as JSON string
}

type ZeroKnowledgeProofVerificationResponse struct {
	IdentityId     string `json:"identity_id"`
	IsProofValid   bool   `json:"is_proof_valid"`
	ProofReference string `json:"proof_reference"`
	Schema         string `json:"schema"`
	Error          string `json:"error,omitempty"`
}

func MockZKPVerification(
	req ZeroKnowledgeProofVerificationRequest,
	sig solana.Signature) ZeroKnowledgeProofVerificationResponse {
	return ZeroKnowledgeProofVerificationResponse{
		IdentityId:     req.IdentityId,
		IsProofValid:   true,
		ProofReference: sig.String(),
		Schema:         req.Schema,
	}
}

func PublishVerificationResult(
	ch *amqp.Channel,
	exchange, routingKey string,
	resp ZeroKnowledgeProofVerificationResponse,
) error {
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return ch.Publish(
		exchange,
		routingKey,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}
