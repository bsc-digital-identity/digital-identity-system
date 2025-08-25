package logger

import "sync"

type LoggerArg struct {
	Key   string
	Value string
}

type GlobalLoggerConfig struct {
	Args []LoggerArg
}

var (
	defaultLogger     *Logger
	onceLogger        sync.Once
	initializedLogger bool
)

func InitDefaultLogger(config GlobalLoggerConfig) {
	onceLogger.Do(func() {
		defaultLogger = New()
		for _, arg := range config.Args {
			defaultLogger.With().Str(arg.Key, arg.Value)
		}

		initializedLogger = true
	})
}

func Default() *Logger {
	if !initializedLogger {
		panic("Deafult logger not initialized: call InitDefaultLogger() first")
	}
	return defaultLogger
}
