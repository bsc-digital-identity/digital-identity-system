package main

import (
	"blockchain-client/src/queues"
	"blockchain-client/src/utils"
)

func main() {
	// 1. Connect to RabbitMQ
	conn, err := queues.ConnectToRabbitmq()
	utils.FailOnError(err, "Failed to connect to RabbitMQ after retries")
	defer conn.Close()

	// 2. Open channel
	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	// 3. Declare exchange and both queues, and bind
	err = queues.SetupIdentityQueues(ch)
	utils.FailOnError(err, "Failed to setup exchange/queues")

	// 4. Start consuming from the job queue ("identity.verified")
	go queues.HandleIncomingMessages(ch, "identity.verified", "")

	// 5. Keep alive
	select {}
}
