package factory

import (
	"fmt"

	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
)

// DriftDetectorFactory creates drift detector services
type DriftDetectorFactory struct {
	logger *logging.Logger
}

// NewDriftDetectorFactory creates a new drift detector factory
func NewDriftDetectorFactory(logger *logging.Logger) *DriftDetectorFactory {
	return &DriftDetectorFactory{
		logger: logger,
	}
}

// CreateDriftDetector creates a drift detector service based on configuration
func (f *DriftDetectorFactory) CreateDriftDetector(
	awsProvider service.InstanceProvider,
	terraformProvider service.InstanceProvider,
	repository service.DriftRepository,
	reporters []service.Reporter,
	cfg *config.Config,
	serviceFactory func(
		awsProvider service.InstanceProvider,
		terraformProvider service.InstanceProvider,
		repository service.DriftRepository,
		reporters []service.Reporter,
		config service.DriftDetectorConfig,
		logger *logging.Logger,
	) service.DriftDetectorProvider,
) (service.DriftDetectorProvider, error) {
	if serviceFactory == nil {
		return nil, fmt.Errorf("drift detector service factory is nil")
	}

	f.logger.Info(fmt.Sprintf("Creating drift detector with source of truth: %s", cfg.GetSourceOfTruth()))

	detectorConfig := service.DriftDetectorConfig{
		SourceOfTruth:      model.ResourceOrigin(cfg.GetSourceOfTruth()),
		AttributePaths:     cfg.GetAttributes(),
		ParallelChecks:     cfg.GetParallelChecks(),
		Timeout:            cfg.GetTimeout(),
		ScheduleExpression: cfg.GetScheduleExpression(),
	}

	f.logger.Debug("Drift detector configuration:")
	f.logger.Debug("  - Source of truth: %s", detectorConfig.SourceOfTruth)
	f.logger.Debug("  - Attribute paths: %v", detectorConfig.AttributePaths)
	f.logger.Debug("  - Parallel checks: %d", detectorConfig.ParallelChecks)
	f.logger.Debug("  - Timeout: %s", detectorConfig.Timeout)
	f.logger.Debug("  - Schedule expression: %s", detectorConfig.ScheduleExpression)

	driftDetector := serviceFactory(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		detectorConfig,
		f.logger,
	)

	f.logger.Info("Drift detector created successfully")
	return driftDetector, nil
}

// CreateDriftDetectorWithCustomConfig creates a drift detector with a custom configuration
func (f *DriftDetectorFactory) CreateDriftDetectorWithCustomConfig(
	awsProvider service.InstanceProvider,
	terraformProvider service.InstanceProvider,
	repository service.DriftRepository,
	reporters []service.Reporter,
	cfg service.DriftDetectorConfig,
	serviceFactory func(
		awsProvider service.InstanceProvider,
		terraformProvider service.InstanceProvider,
		repository service.DriftRepository,
		reporters []service.Reporter,
		config service.DriftDetectorConfig,
		logger *logging.Logger,
	) service.DriftDetectorProvider,
) (service.DriftDetectorProvider, error) {
	if serviceFactory == nil {
		return nil, fmt.Errorf("drift detector service factory is nil")
	}

	f.logger.Info("Creating drift detector with custom configuration")

	driftDetector := serviceFactory(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		cfg,
		f.logger,
	)

	f.logger.Info("Drift detector created successfully")
	return driftDetector, nil
}
