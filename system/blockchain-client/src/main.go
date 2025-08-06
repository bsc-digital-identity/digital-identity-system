package main

import (
	"blockchain-client/src/config"
	"blockchain-client/src/external"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	"pkg-common/utilities"
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

	defaultLogger := logger.Default()

	blockchainClientConfig, err := utilities.ReadConfig[BlockchainClientConfigJson]("config.json")
	blockchainLogger := logger.NewFromConfig(blockchainClientConfig.LoggerConf)

	blockchainLogger.Infof("Loaded config %s", blockchainClientConfig)
	solanaConfig, err := config.LoadSolanaKeys()
	if err != nil {
		blockchainLogger.Fatal(err, "Unable to load keypairs for solana")
	}

	conn, err := rabbitmq.ConnectToRabbitmq(
		blockchainClientConfig.RabbimqConf.User,
		blockchainClientConfig.RabbimqConf.Password,
	)

	rabbitmq.InitializeConsumerRegistry(conn, blockchainClientConfig.RabbimqConf.ConsumersConfig)
	rabbitmq.InitializePublisherRegistry(conn, blockchainClientConfig.RabbimqConf.PublishersConfig)

	defer conn.Close()

	solanaClient := external.NewSolanaClient(solanaConfig)

	go solanaClient.StartService()

	defaultLogger.Info("Blockchain client started and listening for messages")

	select {}
}
