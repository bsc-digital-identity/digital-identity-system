package queues

import (
	"blockchain-client/src/external"
	"blockchain-client/src/utils"
	"blockchain-client/src/zkp"
	"encoding/json"
	"log"
	"sync"
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

			// TODO: mock data read from reqesut
			// replace to read from request
			zkpResult, err := zkp.CreateZKP(10, 10, 1990)
			if err != nil {
				log.Printf("[ERROR]: Failed to create ZKP with user provided data: %s \n error: %s", 10, err)
				return
			}

			signatureChan := make(chan solana.Signature)
			errChan := make(chan error)

			go solanaClient.PublishZkpToSolana(*zkpResult, errChan, signatureChan)

			var signature solana.Signature
			select {
			case signature = <-signatureChan:
				log.Printf("[INFO]: Saved zkp to blockchain with signature: %s", signature.String())
			case err := <-errChan:
				log.Printf("[ERROR]: Unable to save the ZKP to the blockchain %s", err)
				continue
			}

			result := MockZKPVerification(req, signature)

			_ = PublishVerificationResult(ch, "identity", "identity.verified.results", result)
			log.Printf("Processed ZKP Verification for %s. ProofReference: %s", req.IdentityId, result.ProofReference)
		}
	}()

	waitGroup.Wait()
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
