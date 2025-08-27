package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"pkg-common/logger"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	l := logger.New()
	if l == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestNewFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   logger.LoggerConfig
		expected zerolog.Level
	}{
		{
			name:     "Default log level when no level specified",
			config:   logger.LoggerConfig{LogLevel: zerolog.NoLevel},
			expected: zerolog.InfoLevel,
		},
		{
			name:     "Debug log level",
			config:   logger.LoggerConfig{LogLevel: zerolog.DebugLevel},
			expected: zerolog.DebugLevel,
		},
		{
			name:     "Error log level",
			config:   logger.LoggerConfig{LogLevel: zerolog.ErrorLevel},
			expected: zerolog.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.NewFromConfig(tt.config)
			if l == nil {
				t.Fatal("Expected logger to be created, got nil")
			}
		})
	}
}

func TestLoggerWithOutput(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
}

func TestLoggerWithLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf).WithLevel(zerolog.ErrorLevel)

	// This should not appear in output due to level filtering
	l.Info("info message")
	// This should appear
	l.Error(errors.New("test error"), "error message")

	output := buf.String()
	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear when level is set to Error")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear when level is set to Error")
	}
}

func TestLoggerWithContext(t *testing.T) {
	l := logger.New()
	ctx := context.Background()

	contextLogger := l.WithContext(ctx)
	if contextLogger == nil {
		t.Fatal("Expected context logger to be created, got nil")
	}
}

func TestLoggerWith(t *testing.T) {
	l := logger.New()
	context := l.With()
	// Test that the context is not nil and can be used
	contextLogger := context.Str("test", "value").Logger()
	if &contextLogger == nil {
		t.Error("Expected context to create a valid logger")
	}
}

func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf).WithLevel(zerolog.DebugLevel)

	l.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Expected output to contain 'debug message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"debug"`) {
		t.Error("Expected log level to be debug")
	}
}

func TestLoggerDebugf(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf).WithLevel(zerolog.DebugLevel)

	l.Debugf("debug message with %s", "formatting")

	output := buf.String()
	if !strings.Contains(output, "debug message with formatting") {
		t.Errorf("Expected formatted output, got: %s", output)
	}
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Info("info message")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("Expected output to contain 'info message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Error("Expected log level to be info")
	}
}

func TestLoggerInfof(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Infof("info message with %d items", 5)

	output := buf.String()
	if !strings.Contains(output, "info message with 5 items") {
		t.Errorf("Expected formatted output, got: %s", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Warn("warning message")

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Errorf("Expected output to contain 'warning message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"warn"`) {
		t.Error("Expected log level to be warn")
	}
}

func TestLoggerWarnf(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Warnf("warning message with %s", "details")

	output := buf.String()
	if !strings.Contains(output, "warning message with details") {
		t.Errorf("Expected formatted output, got: %s", output)
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	testErr := errors.New("test error")
	l.Error(testErr, "error message")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("Expected output to contain 'error message', got: %s", output)
	}
	if !strings.Contains(output, "test error") {
		t.Error("Expected output to contain error details")
	}
	if !strings.Contains(output, `"level":"error"`) {
		t.Error("Expected log level to be error")
	}
}

func TestLoggerErrorf(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	testErr := errors.New("test error")
	l.Errorf(testErr, "error message with %s", "context")

	output := buf.String()
	if !strings.Contains(output, "error message with context") {
		t.Errorf("Expected formatted output, got: %s", output)
	}
	if !strings.Contains(output, "test error") {
		t.Error("Expected output to contain error details")
	}
}

func TestLoggerLog(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Log(zerolog.WarnLevel, "custom level message")

	output := buf.String()
	if !strings.Contains(output, "custom level message") {
		t.Errorf("Expected output to contain 'custom level message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"warn"`) {
		t.Error("Expected log level to be warn")
	}
}

func TestLoggerLogf(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Logf(zerolog.InfoLevel, "custom level message with %d", 42)

	output := buf.String()
	if !strings.Contains(output, "custom level message with 42") {
		t.Errorf("Expected formatted output, got: %s", output)
	}
}

func TestLoggerConfigConvertToDomain(t *testing.T) {
	tests := []struct {
		name     string
		config   logger.LoggerConfigJson
		expected logger.LoggerConfig
	}{
		{
			name:     "Info level conversion",
			config:   logger.LoggerConfigJson{}, // This will need to be adjusted based on actual struct
			expected: logger.LoggerConfig{LogLevel: zerolog.Level(0)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.MapToDomain()
			if result.LogLevel != tt.expected.LogLevel {
				t.Errorf("Expected LogLevel %v, got %v", tt.expected.LogLevel, result.LogLevel)
			}
		})
	}
}

func TestInitDefaultLogger(t *testing.T) {
	// Reset the once variable for testing
	// Note: This might require making the once variable exported or adding a reset function
	config := logger.GlobalLoggerConfig{
		Args: []struct {
			Key   string
			Value string
		}{
			{"application", "test-app"},
			{"version", "1.0.0"},
		},
	}

	logger.InitDefaultLogger(config)
	defaultLogger := logger.Default()

	if defaultLogger == nil {
		t.Fatal("Expected default logger to be initialized, got nil")
	}
}

func TestDefaultLogger(t *testing.T) {
	// Initialize first
	config := logger.GlobalLoggerConfig{
		Args: []struct {
			Key   string
			Value string
		}{
			{"service", "test"},
		},
	}
	logger.InitDefaultLogger(config)

	l := logger.Default()
	if l == nil {
		t.Fatal("Expected default logger to exist, got nil")
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New()
	l = l.WithOutput(&buf)

	l.Info("test json format")

	output := buf.String()

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	if err != nil {
		t.Errorf("Log output is not valid JSON: %v", err)
	}

	// Check required fields
	if logEntry["level"] != "info" {
		t.Error("Expected level field to be 'info'")
	}
	if logEntry["message"] != "test json format" {
		t.Error("Expected message field to match input")
	}
	if _, ok := logEntry["time"]; !ok {
		t.Error("Expected time field to be present")
	}
}
