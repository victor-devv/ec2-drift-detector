package app

import (
	"context"

	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/container"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/factory"
)

func init() {
	// Register the drift detector service factory function
	container.RegisterDriftDetectorServiceFactory(func(
		awsProvider service.InstanceProvider,
		terraformProvider service.InstanceProvider,
		repository service.DriftRepository,
		reporters []service.Reporter,
		config service.DriftDetectorConfig,
		logger *logging.Logger,
	) service.DriftDetectorProvider {
		return NewDriftDetectorService(
			awsProvider,
			terraformProvider,
			repository,
			reporters,
			config,
			logger,
		)
	})
}

func initializeDriftDetector(
	ctx context.Context,
	cfg *config.Config,
	instanceProviderFactory *factory.InstanceProviderFactory,
	driftDetectorFactory *factory.DriftDetectorFactory,
	repository service.DriftRepository,
	reporters []service.Reporter,
	c *container.Container,
) (service.DriftDetectorProvider, error) {
	awsProvider, err := instanceProviderFactory.CreateAWSProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	terraformProvider, err := instanceProviderFactory.CreateTerraformProvider(cfg)
	if err != nil {
		return nil, err
	}

	// Get the factory function
	serviceFactory, err := c.GetDriftDetectorServiceFactory()
	if err != nil {
		return nil, err
	}

	return driftDetectorFactory.CreateDriftDetector(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		cfg,
		serviceFactory,
	)
}

// InitializeApplication creates and configures the application based on the configuration
func InitializeApplication(ctx context.Context, c *container.Container, cfg *config.Config) (*Application, error) {
	instanceProviderFactory, _ := container.Resolve[*factory.InstanceProviderFactory](c, "instanceProviderFactory")
	driftDetectorFactory, _ := container.Resolve[*factory.DriftDetectorFactory](c, "driftDetectorFactory")
	reporterFactory, _ := container.Resolve[*factory.ReporterFactory](c, "reporterFactory")
	repositoryFactory, _ := container.Resolve[*factory.RepositoryFactory](c, "repositoryFactory")
	repository := repositoryFactory.CreateDriftRepository()

	reporters, err := reporterFactory.CreateReporters(cfg)
	if err != nil {
		return nil, err
	}

	driftDetector, err := initializeDriftDetector(
		ctx,
		cfg,
		instanceProviderFactory,
		driftDetectorFactory,
		repository,
		reporters,
		c,
	)
	if err != nil {
		return nil, err
	}

	return NewApplication(driftDetector), nil
}
