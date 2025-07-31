package logger

import (
	"io"
	"os"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOW"
	}
}

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Reset  = "\033[0m"
)

type LoggerConfig struct {
	MinLevel       LogLevel
	Output         io.Writer
	EnableColors   bool
	EnableLocation bool
	EnableTime     bool
	ColorMapping   map[LogLevel]string
	TimeFormat     string
	Location       string
}

type Logger struct {
	level        LogLevel
	output       io.Writer
	colorMapping map[LogLevel]string
	timeFormat   string
	location     string
}

func DefaultConfig() LoggerConfig {
	return LoggerConfig{
		MinLevel:       INFO,
		Output:         os.Stdout,
		EnableColors:   true,
		EnableLocation: true,
		EnableTime:     true,
		TimeFormat:     "01-01-2025 18:00:00",
		Location:       "",
	}
}

func New(config LoggerConfig) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	if !config.EnableColors {
		config.ColorMapping = nil
	}

	if !config.EnableTime {
		config.TimeFormat = ""
	}

	if !config.EnableLocation {
		config.Location = ""
	}

	return &Logger{
		level:        config.MinLevel,
		output:       config.Output,
		colorMapping: config.ColorMapping,
		timeFormat:   config.TimeFormat,
		location:     config.Location,
	}
}

func NewDeafult() *Logger {
	return New(DefaultConfig())
}

func NewWithLocation(config LoggerConfig, location string) *Logger {
	config.Location = location
	return New(config)
}

func (l *Logger) getMessageColor(level LogLevel) string {
	if l.colorMapping != nil {
		return l.colorMapping[level]
	}

	return ""
}

// TODO: finish if it has value
// custom unfified logger
//func formatMessage(level)
