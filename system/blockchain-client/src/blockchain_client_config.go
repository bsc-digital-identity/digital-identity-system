package main

import (
	"pkg-common/logger"
	"pkg-common/rabbitmq"
)

type BlockchainClientConfigJson struct {
	LoggerConf   logger.LoggerConfigJson    `json:"logger"`
	RabbitmqConf rabbitmq.RabbimqConfigJson `json:"rabbitmq"`
}

type BlockchainClientConfig struct {
	LoggerConf  logger.LoggerConfig
	RabbimqConf rabbitmq.RabbitmqConfig
}

func (bccj BlockchainClientConfigJson) ConvertToDomain() BlockchainClientConfig {
	return BlockchainClientConfig{
		LoggerConf:  bccj.LoggerConf.ConvertToDomain(),
		RabbimqConf: bccj.RabbitmqConf.ConvertToDomain(),
	}
}
