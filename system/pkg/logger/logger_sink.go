package logger

import (
	"fmt"
)

func AddSinkToLoggerInstance(loggerInstance *Logger, sinkFunction func(string)) {
	loggerInstance.sink = sinkFunction
}

func (l *Logger) activateSinkFormatted(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v)
	l.activateSink(msg)
}

func (l *Logger) activateSink(msg string) {
	if l.sink != nil {
		l.sink(msg)
	}
}
