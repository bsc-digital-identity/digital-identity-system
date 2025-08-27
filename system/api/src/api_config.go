package main

import (
	"pkg-common/logger"
	"pkg-common/rabbitmq"
)

type ApiConfigJson struct {
	LoggerConf   logger.LoggerConfigJson     `json:"logger"`
	RabbitmqConf rabbitmq.RabbimqConfigJson  `json:"rabbitmq"`
	RestConf     ApiClientRestConfigJson     `json:"rest"`
	DatabaseConf ApiClientDatabaseConfigJson `json:"database"`
}

func (acj ApiConfigJson) ConvertToDomain() ApiConfig {
	return ApiConfig{
		LoggerConf:   acj.LoggerConf.ConvertToDomain(),
		RabbitmqConf: acj.RabbitmqConf.ConvertToDomain(),
		RestConf:     acj.RestConf.ConvertToDomain(),
	}
}

type ApiConfig struct {
	LoggerConf   logger.LoggerConfig
	RabbitmqConf rabbitmq.RabbitmqConfig
	RestConf     ApiClientRestConfig
	DatabaseConf ApiClientDatabaseConfig
}

func (ac ApiConfig) GetLoggerConfig() logger.LoggerConfig {
	return ac.LoggerConf
}

func (ac ApiConfig) GetRabbitmqConfig() rabbitmq.RabbitmqConfig {
	return ac.RabbitmqConf
}

func (ac ApiConfig) GetRestApiPort() uint16 {
	return ac.RestConf.Port
}

func (ac ApiConfig) GetDatabaseConnectionString() string {
	return ac.DatabaseConf.ConnectionString
}

type ApiClientRestConfigJson struct {
	Port uint16 `json:"port"`
}

type ApiClientRestConfig struct {
	Port uint16
}

func (acrcj ApiClientRestConfigJson) ConvertToDomain() ApiClientRestConfig {
	return ApiClientRestConfig{
		Port: acrcj.Port,
	}
}

type ApiClientDatabaseConfigJson struct {
	ConnectionString string `json:"connection_string"`
}

type ApiClientDatabaseConfig struct {
	ConnectionString string
}

func (acdcj ApiClientDatabaseConfigJson) ConvertToDomain() ApiClientDatabaseConfig {
	return ApiClientDatabaseConfig{
		ConnectionString: acdcj.ConnectionString,
	}
}
