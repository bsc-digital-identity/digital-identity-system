package logger

import "sync"

type GlobalLoggerConfig struct {
	Args []struct {
		Key   string
		Value string
	}
}

var (
	defaultLogger *Logger
	once          sync.Once
)

func InitDefaultLogger(config GlobalLoggerConfig) {
	once.Do(func() {
		defaultLogger = New()
		for _, arg := range config.Args {
			defaultLogger.With().Str(arg.Key, arg.Value)
		}
	})
}

func Default() *Logger {
	return defaultLogger
}
