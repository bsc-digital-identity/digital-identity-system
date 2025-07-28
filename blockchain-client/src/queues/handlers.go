package queues

import (
	"blockchain-client/src/external"
	"blockchain-client/src/utils"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
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

func HandleIncomingMessages(
	solanaClient *external.SolanaClient,
	ch *amqp.Channel,
	queueName,
	consumerTag string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[%s] Recovered from panic for consumer: %s, %v\n", queueName, consumerTag, r)
		}
	}()

	msgs, err := ch.Consume(
		queueName,   // queue
		consumerTag, // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	utils.FailOnError(err, "Failed to register a consumer")

	log.Printf("Waiting for messages in queue: %s", queueName)
	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for d := range msgs {
			log.Printf("[%s] %s", queueName, d.Body)

			var req ZeroKnowledgeProofVerificationRequest
			err := json.Unmarshal(d.Body, &req)
			if err != nil {
				result := ZeroKnowledgeProofVerificationResponse{
					IdentityId:   req.IdentityId,
					IsProofValid: false,
					Error:        "unmarshal: " + err.Error(),
				}
				_ = PublishVerificationResult(ch, "identity", "identity.verified.results", result)
				continue
			}

			result := MockZKPVerification(req)
			_ = PublishVerificationResult(ch, "identity", "identity.verified.results", result)
			log.Printf("Processed ZKP Verification for %s. ProofReference: %s", req.IdentityId, result.ProofReference)
		}
	}()

	waitGroup.Wait()
}

func MockZKPVerification(req ZeroKnowledgeProofVerificationRequest) ZeroKnowledgeProofVerificationResponse {
	ref := uuid.NewString()
	return ZeroKnowledgeProofVerificationResponse{
		IdentityId:     req.IdentityId,
		IsProofValid:   true,
		ProofReference: ref,
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
