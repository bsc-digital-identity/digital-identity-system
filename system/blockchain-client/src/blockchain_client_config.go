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

func (bccj BlockchainClientConfigJson) MapToDomain() BlockchainClientConfig {
	return BlockchainClientConfig{
		LoggerConf:   bccj.LoggerConf.MapToDomain(),
		RabbitmqConf: bccj.RabbitmqConf.MapToDomain(),
		RestConf:     bccj.RestConf.MapToDomain(),
	}
}

type BlockchainClientConfig struct {
	LoggerConf   logger.LoggerConfig
	RabbitmqConf rabbitmq.RabbitmqConfig
	RestConf     BlockchainClientRestConfig
}

func (bcc BlockchainClientConfig) GetLoggerConfig() logger.LoggerConfig {
	return bcc.LoggerConf
}

func (bcc BlockchainClientConfig) GetRabbitmqConfig() rabbitmq.RabbitmqConfig {
	return bcc.RabbitmqConf
}

func (bcc BlockchainClientConfig) GetRestApiPort() uint16 {
	return bcc.RestConf.Port
}

type BlockchainClientRestConfigJson struct {
	Port uint16 `json:"port"`
}

type BlockchainClientRestConfig struct {
	Port uint16
}

func (bcrcj BlockchainClientRestConfigJson) MapToDomain() BlockchainClientRestConfig {
	return BlockchainClientRestConfig{
		Port: bcrcj.Port,
	}
}
