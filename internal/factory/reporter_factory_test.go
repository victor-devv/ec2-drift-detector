package factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/factory"
)

func newTestConfig(reporterType, outputFile string) *config.Config {
	cfg := &config.Config{}
	cfg.SetReporterType(reporterType)
	cfg.SetOutputFile(outputFile)
	return cfg
}

func TestCreateConsoleReporter(t *testing.T) {
	logger := logging.New()
	factory := factory.NewReporterFactory(logger)
	r := factory.CreateConsoleReporter(logger)
	assert.NotNil(t, r)
}

func TestCreateJSONReporter(t *testing.T) {
	logger := logging.New()
	factory := factory.NewReporterFactory(logger)
	r := factory.CreateJSONReporter(logger, "output.json")
	assert.NotNil(t, r)
}

func TestCreateReporters_ConsoleOnly(t *testing.T) {
	logger := logging.New()
	factory := factory.NewReporterFactory(logger)
	cfg := newTestConfig("console", "")

	reporters, err := factory.CreateReporters(cfg)
	assert.NoError(t, err)
	assert.Len(t, reporters, 1)
}

func TestCreateReporters_JSONOnly(t *testing.T) {
	logger := logging.New()
	factory := factory.NewReporterFactory(logger)
	cfg := newTestConfig("json", "report.json")

	reporters, err := factory.CreateReporters(cfg)
	assert.NoError(t, err)
	assert.Len(t, reporters, 1)
}

func TestCreateReporters_Both(t *testing.T) {
	logger := logging.New()
	factory := factory.NewReporterFactory(logger)
	cfg := newTestConfig("both", "report.json")

	reporters, err := factory.CreateReporters(cfg)
	assert.NoError(t, err)
	assert.Len(t, reporters, 2)
}

func TestCreateReporters_JSONMissingFile(t *testing.T) {
	logger := logging.New()
	factory := factory.NewReporterFactory(logger)
	cfg := newTestConfig("json", "")

	reporters, err := factory.CreateReporters(cfg)
	assert.NoError(t, err)
	assert.Len(t, reporters, 1)
}
