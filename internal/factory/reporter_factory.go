package factory

import (
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/reporter"
)

// ReporterFactory creates different types of reporters
type ReporterFactory struct {
	logger *logging.Logger
}

// NewReporterFactory creates a new reporter factory
func NewReporterFactory(logger *logging.Logger) *ReporterFactory {
	return &ReporterFactory{
		logger: logger,
	}
}

func (f *ReporterFactory) CreateReporters(cfg *config.Config) ([]service.Reporter, error) {
	var reporters []service.Reporter

	reporterType := cfg.GetReporterType()

	switch reporterType {
	case config.ReporterTypeConsole:
		reporters = append(reporters, reporter.NewConsoleReporter(f.logger))
	case config.ReporterTypeJSON:
		reporters = append(reporters, reporter.NewJSONReporter(f.logger, cfg.GetOutputFile()))
	case config.ReporterTypeBoth:
		reporters = append(reporters, reporter.NewConsoleReporter(f.logger))
		reporters = append(reporters, reporter.NewJSONReporter(f.logger, cfg.GetOutputFile()))
	}
	f.logger.Info("Reporters created successfully")
	return reporters, nil
}

// CreateConsoleReporter creates a console reporter
func (f *ReporterFactory) CreateConsoleReporter(logger *logging.Logger) service.Reporter {
	return reporter.NewConsoleReporter(logger)
}

// CreateJSONReporter creates a JSON reporter
func (f *ReporterFactory) CreateJSONReporter(logger *logging.Logger, outputFile string) service.Reporter {
	return reporter.NewJSONReporter(logger, outputFile)
}
