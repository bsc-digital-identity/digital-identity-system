package logger

import "github.com/rs/zerolog"

type LoggerConfigJson struct {
	logLevel int8 `json:"log_level"`
}

type LoggerConfig struct {
	LogLevel zerolog.Level
}

func (lcj LoggerConfigJson) MapToDomain() LoggerConfig {
	return LoggerConfig{
		LogLevel: zerolog.Level(lcj.logLevel),
	}
}
