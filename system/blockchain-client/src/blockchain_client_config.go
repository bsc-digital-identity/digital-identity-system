package main

import (
	"pkg-common/logger"
	"pkg-common/rabbitmq"
)

type BlockchainClientConfigJson struct {
	LoggerConf   logger.LoggerConfigJson        `json:"logger"`
	RabbitmqConf rabbitmq.RabbimqConfigJson     `json:"rabbitmq"`
	RestConf     BlockchainClientRestConfigJson `json:"rest"`
}

type BlockchainClientConfig struct {
	LoggerConf  logger.LoggerConfig
	RabbimqConf rabbitmq.RabbitmqConfig
	RestConf    BlockchainClientRestConfig
}

func (bccj BlockchainClientConfigJson) ConvertToDomain() BlockchainClientConfig {
	return BlockchainClientConfig{
		LoggerConf:  bccj.LoggerConf.ConvertToDomain(),
		RabbimqConf: bccj.RabbitmqConf.ConvertToDomain(),
		RestConf:    bccj.RestConf.ConvertToDomain(),
	}
}

type BlockchainClientRestConfigJson struct {
	Port uint16 `json:"port"`
}

type BlockchainClientRestConfig struct {
	Port uint16
}

func (bcrcj BlockchainClientRestConfigJson) ConvertToDomain() BlockchainClientRestConfig {
	return BlockchainClientRestConfig{
		Port: bcrcj.Port,
	}
}
