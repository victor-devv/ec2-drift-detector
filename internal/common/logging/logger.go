package logging

import (
	"github.com/hashicorp/go-hclog"
	"io"
	"os"
	"sync"
)

// Logger wraps the hclog.Logger to provide application-specific logging
type Logger struct {
	hclog.Logger
	mu sync.Mutex
}

var (
	// globalLogger is the singleton instance of Logger
	globalLogger *Logger
	// once ensures the global logger is initialized only once
	once sync.Once
)

// LogLevel represents the log level
type LogLevel string

const (
	// Debug level for detailed debugging information
	Debug LogLevel = "DEBUG"
	// Info level for informational messages
	Info LogLevel = "INFO"
	// Warn level for warning messages
	Warn LogLevel = "WARN"
	// Error level for error messages
	Error LogLevel = "ERROR"
)

// LogConfig defines the configuration for the logger
type LogConfig struct {
	Level      LogLevel
	Output     io.Writer
	JSONFormat bool
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config LogConfig) *Logger {
	level := hclog.Info

	switch config.Level {
	case Debug:
		level = hclog.Debug
	case Info:
		level = hclog.Info
	case Warn:
		level = hclog.Warn
	case Error:
		level = hclog.Error
	}

	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "drift-detector",
		Level:      level,
		Output:     output,
		JSONFormat: config.JSONFormat,
	})

	return &Logger{
		Logger: logger,
	}
}

// New returns the global logger instance, initializing it if necessary
func New() *Logger {
	once.Do(func() {
		globalLogger = NewLogger(LogConfig{
			Level:      Info,
			JSONFormat: false,
		})
	})

	return globalLogger
}

// SetLogger sets the global logger instance
func SetLogger(logger *Logger) {
	if logger != nil {
		globalLogger = logger
	}
}

// ConfigureLogger configures the global logger with the given configuration
func ConfigureLogger(config LogConfig) {
	SetLogger(NewLogger(config))
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: l.With(key, value).(hclog.Logger),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.Logger

	for k, v := range fields {
		newLogger = newLogger.With(k, v).(hclog.Logger)
	}

	return &Logger{
		Logger: newLogger,
	}
}

// GetLogLevel returns the current log level as a string
func (l *Logger) GetLogLevel() LogLevel {
	level := l.GetLevel()

	switch level {
	case hclog.Debug:
		return Debug
	case hclog.Info:
		return Info
	case hclog.Warn:
		return Warn
	case hclog.Error:
		return Error
	default:
		return Info
	}
}

// SetLogLevel sets the log level
func (l *Logger) SetLogLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var hcLevel hclog.Level

	switch level {
	case Debug:
		hcLevel = hclog.Debug
	case Info:
		hcLevel = hclog.Info
	case Warn:
		hcLevel = hclog.Warn
	case Error:
		hcLevel = hclog.Error
	default:
		hcLevel = hclog.Info
	}

	l.SetLevel(hcLevel)
}
