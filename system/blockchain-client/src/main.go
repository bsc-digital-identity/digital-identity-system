package main

import (
	"blockchain-client/src/config"
	"blockchain-client/src/external"
	"blockchain-client/src/queues"
	"blockchain-client/src/utils"
	"pkg-common/logger"

	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	// Initialize logger
	logger.InitDefaultLogger(logger.GlobalLoggerConfig{
		Args: []struct {
			Key   string
			Value string
		}{
			{"application", "blockchain-client"},
			{"version", "1.0.0"},
		},
	})

	mainLogger := logger.Default()

	solanaConfig, err := config.LoadSolanaKeys()
	if err != nil {
		mainLogger.Fatal(err, "Unable to load keypairs for solana")
	}

	rpcClient := rpc.New("http://host.docker.internal:8899")
	solanaClient := &external.SolanaClient{
		Config:    solanaConfig,
		RpcClient: rpcClient,
	}

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
	go queues.HandleIncomingMessages(solanaClient, ch, "identity.verified", "")

	mainLogger.Info("Blockchain client started and listening for messages")

	// 5. Keep alive
	select {}
}
