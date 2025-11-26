package logger

import (
	"fmt"
	"pkg-common/utilities/timeutil"

	"github.com/rs/zerolog"
)

func AddSinkToLoggerInstance(loggerInstance *Logger, sinkFunction func(string, zerolog.Level, timeutil.TimeUTC)) {
	loggerInstance.sink = sinkFunction
}

func (l *Logger) activateSinkFormatted(format string, level zerolog.Level, timestamp timeutil.TimeUTC, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.activateSink(msg, level, timestamp)
}

func (l *Logger) activateSink(msg string, level zerolog.Level, timestamp timeutil.TimeUTC) {
	if l.sink != nil {
		l.sink(msg, level, timestamp)
	}
}
