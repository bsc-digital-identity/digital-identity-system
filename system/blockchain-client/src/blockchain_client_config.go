package main

import (
	"blockchain-client/src/queues"
	"pkg-common/logger"
)

type BlockchainClientConfigJson struct {
	loggerConf   logger.LoggerConfigJson  `json:"logger"`
	rabbitmqConf queues.RabbimqConfigJson `json:"rabbitmq"`
}

type BlockchainClientConfig struct {
	LoggerConf  logger.LoggerConfig
	RabbimqConf queues.RabbitmqConfig
}

func (bccj BlockchainClientConfigJson) ConvertToDomain() BlockchainClientConfig {
	return BlockchainClientConfig{
		LoggerConf:  bccj.loggerConf.ConvertToDomain(),
		RabbimqConf: bccj.rabbitmqConf.ConvertToDomain(),
	}
}
