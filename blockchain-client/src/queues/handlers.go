package queues

import (
	"blockchain-client/src/api"
	"blockchain-client/src/utils"
	"blockchain-client/src/zkp"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ZkpVerifiedMessage struct {
	IdentityId string `json:"identity_id"`
	SchemaId   string `json:"schema_id"`
	BirthDay   int    `json:"birth_day"`
	BirthMonth int    `json:"birth_month"`
	BirthYear  int    `json:"birth_year"`
}

type VerificationResultMessage struct {
	IdentityId   string `json:"identity_id"`
	Success      bool   `json:"success"`
	BlockchainTx string `json:"blockchain_tx,omitempty"`
	Error        string `json:"error,omitempty"`
}

func HandleIncomingMessages(ch *amqp.Channel, queueName, consumerTag string) {
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

			var msg ZkpVerifiedMessage
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Failed to unmarshal message: %s", err)
				// If possible, try to extract identity_id for reporting
				var partial struct {
					IdentityId string `json:"identity_id"`
				}
				_ = json.Unmarshal(d.Body, &partial)
				_ = PublishVerificationResult(ch, "identity", "identity.verified.results", VerificationResultMessage{
					IdentityId: partial.IdentityId,
					Success:    false,
					Error:      "unmarshal: " + err.Error(),
				})
				continue
			}

			zkpResult, err := zkp.CreateZKP(msg.BirthDay, msg.BirthMonth, msg.BirthYear)
			if err != nil {
				log.Printf("Failed to create ZKP: %s", err)
				_ = PublishVerificationResult(ch, "identity", "identity.verified.results", VerificationResultMessage{
					IdentityId: msg.IdentityId,
					Success:    false,
					Error:      "zkp: " + err.Error(),
				})
				continue
			}

			log.Println(zkpResult)

			// gen new blockchain ref
			blockRef, _ := uuid.NewRandom()

			type ZkpProof struct {
				DigitalIdentitySchemaId string
				SuperIdentityId         string
				ProofReference          string
			}

			validProofRequest := ZkpProof{
				blockRef.String(),
				msg.IdentityId,
				msg.SchemaId,
			}

			responseCh := make(chan struct{})
			errorCh := make(chan error)
			go api.ReqeuestBase[struct{}, ZkpProof]("create", http.MethodPost, errorCh, responseCh, validProofRequest)
			select {
			case <-responseCh:
				log.Println("Succesffuly created new ZKP")
			case err := <-errorCh:
				log.Printf("Failed to create ZKP: %s", err)
			}
		}
	}()

	waitGroup.Wait()
}

func PublishVerificationResult(ch *amqp.Channel, exchange, routingKey string, msg VerificationResultMessage) error {
	body, err := json.Marshal(msg)
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
