package queues

import (
	"blockchain-client/src/utils"
	"blockchain-client/src/zkp"
	"encoding/json"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ZkpVerifiedMessage struct {
	BirthDay   int `json:"birth_day"`
	BirthMonth int `json:"birth_month"`
	BirthYear  int `json:"birth_year"`
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
				continue
			}

			zkpResult, err := zkp.CreateZKP(msg.BirthDay, msg.BirthMonth, msg.BirthYear)
			if err != nil {
				log.Printf("Failed to create ZKP: %s", err)
				continue
			}

			log.Println(zkpResult)
		}
	}()

	waitGroup.Wait()
}
