package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Test case 1: Create logger with minimum configuration
	config := LogConfig{
		Level: Info,
	}
	logger := NewLogger(config)

	assert.NotNil(t, logger)
	assert.Equal(t, hclog.Info, logger.GetLevel())

	// Test case 2: Create logger with custom configuration
	config = LogConfig{
		Level:      Debug,
		Output:     &buf,
		JSONFormat: true,
	}
	logger = NewLogger(config)

	assert.NotNil(t, logger)
	assert.Equal(t, hclog.Debug, logger.GetLevel())

	// Test logging
	logger.Info("test message", "key", "value")

	// Verify output is JSON format
	output := buf.String()
	assert.True(t, strings.HasPrefix(output, "{"))

	// Try to parse as JSON
	var logData map[string]interface{}
	err := json.Unmarshal([]byte(output), &logData)
	assert.NoError(t, err)
	assert.Contains(t, logData, "@message")
	assert.Equal(t, "test message", logData["@message"])
	assert.Contains(t, logData, "key")
	assert.Equal(t, "value", logData["key"])
}

func TestLoggerWithField(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a logger
	config := LogConfig{
		Level:  Debug,
		Output: &buf,
	}
	logger := NewLogger(config)

	// Add a field to the logger
	newLogger := logger.WithField("component", "test")

	// Test logging with the new logger
	newLogger.Info("test message")

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "test message")
}

func TestLoggerWithFields(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a logger
	config := LogConfig{
		Level:  Debug,
		Output: &buf,
	}
	logger := NewLogger(config)

	// Add multiple fields to the logger
	fields := map[string]interface{}{
		"component": "test",
		"version":   "1.0",
	}
	newLogger := logger.WithFields(fields)

	// Test logging with the new logger
	newLogger.Info("test message")

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "version=1.0")
	assert.Contains(t, output, "test message")
}

func TestLogLevels(t *testing.T) {
	// Create buffers to capture log output
	var buf bytes.Buffer

	// Create logger with Debug level
	config := LogConfig{
		Level:  Debug,
		Output: &buf,
	}
	logger := NewLogger(config)

	// Test each log level
	logger.Debug("debug message")
	assert.Contains(t, buf.String(), "debug message")
	buf.Reset()

	logger.Info("info message")
	assert.Contains(t, buf.String(), "info message")
	buf.Reset()

	logger.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message")
	buf.Reset()

	logger.Error("error message")
	assert.Contains(t, buf.String(), "error message")
	buf.Reset()

	// Now change the log level to Error
	logger.SetLogLevel(Error)
	assert.Equal(t, Error, logger.GetLogLevel())

	// Debug and Info should not be logged
	logger.Debug("debug message")
	assert.Empty(t, buf.String())
	buf.Reset()

	logger.Info("info message")
	assert.Empty(t, buf.String())
	buf.Reset()

	logger.Warn("warn message")
	assert.Empty(t, buf.String())
	buf.Reset()

	// Error should still be logged
	logger.Error("error message")
	assert.Contains(t, buf.String(), "error message")
}

func TestGetLogger(t *testing.T) {
	// Test GetLogger singleton
	logger1 := New()
	logger2 := New()

	// Both should be the same instance
	assert.Same(t, logger1, logger2)
}

func TestSetLogger(t *testing.T) {
	// Get the current global logger
	originalLogger := New()

	// Create a new logger
	config := LogConfig{
		Level: Debug,
	}
	newLogger := NewLogger(config)

	// Set the new logger as the global logger
	SetLogger(newLogger)

	// Get the global logger again
	currentLogger := New()

	// Should be the new logger, not the original
	assert.Same(t, newLogger, currentLogger)
	assert.NotSame(t, originalLogger, currentLogger)

	// Try setting nil (should not change)
	SetLogger(nil)
	assert.Same(t, newLogger, New())

	// Reset the original logger for other tests
	SetLogger(originalLogger)
}

func TestConfigureLogger(t *testing.T) {
	// Get the current global logger
	originalLogger := New()

	// Configure with new settings
	ConfigureLogger(LogConfig{
		Level:      Debug,
		JSONFormat: true,
	})

	// Get the new logger
	newLogger := New()

	// Should be a different logger with the new settings
	assert.NotSame(t, originalLogger, newLogger)
	assert.Equal(t, Debug, newLogger.GetLogLevel())

	// Reset the original logger for other tests
	SetLogger(originalLogger)
}
