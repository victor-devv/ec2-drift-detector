package logger_test

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/pkg/logger"
)

func TestLogger_ValidLevels(t *testing.T) {
	tests := []struct {
		level    string
		expected logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
	}

	for _, tt := range tests {
		cfg := &config.Config{Log: config.Config{}.Log}
		cfg.Log.Level = tt.level

		l := logger.New(cfg)
		assert.Equal(t, tt.expected, l.GetLevel(), "logger level should be set to %s", tt.level)

		// Sanity check: logger should be configured to stdout
		assert.Equal(t, os.Stdout, l.Out)
	}
}

func TestLogger_UsesTextFormatter(t *testing.T) {
	cfg := &config.Config{Log: config.Config{}.Log}
	cfg.Log.Level = "info"

	l := logger.New(cfg)
	formatter, ok := l.Formatter.(*logrus.TextFormatter)
	assert.True(t, ok, "logger should use logrus.TextFormatter")
	assert.True(t, formatter.FullTimestamp, "FullTimestamp should be true")
}
