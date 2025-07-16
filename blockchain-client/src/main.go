package main

import (
	"blockchain-client/src/queues"
	"blockchain-client/src/utils"
)

func main() {
	conn, err := queues.ConnectToRabbitmq()
	utils.FailOnError(err, "Failed to connect to RabbitMQ after retries")
	defer conn.Close()

	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = queues.CreateNewExchange(ch, "verifiers", queues.ExchangeFanout)
	utils.FailOnError(err, "Failed to declare an exchange")

	q, err := queues.CreateNewQueue(ch, "verified")
	utils.FailOnError(err, "Failed to declare a queue")

	err = queues.BindQueueToExchange(ch, q.Name, "verifiers", "verifiers")
	utils.FailOnError(err, "Failed to bind a queue")

	go queues.HandleIncomingMessages(ch, q.Name, "")

	select {}
}
