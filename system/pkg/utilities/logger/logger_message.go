package logger_message

import (
	"pkg-common/utilities"
	"pkg-common/utilities/timeutil"
)

type LoggerMessage struct {
	Level     string           `json:"level"`
	Message   string           `json:"message"`
	Timestamp timeutil.TimeUTC `json:"timestamp"`
}

func (lm LoggerMessage) Serialize() ([]byte, error) {
	return utilities.Serialize[LoggerMessage](lm)
}
