package container

import (
	"context"
	"fmt"
	"sync"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/factory"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/cli"
)

// CLIHandlerProvider is an interface for CLI handlers
type CLIHandlerProvider interface {
	Execute(ctx context.Context) error
}

// DriftDetectorServiceFactory is a function type that creates a drift detector service
type DriftDetectorServiceFactory func(
	awsProvider service.InstanceProvider,
	terraformProvider service.InstanceProvider,
	repository service.DriftRepository,
	reporters []service.Reporter,
	config service.DriftDetectorConfig,
	logger *logging.Logger,
) service.DriftDetectorProvider

// Container manages dependencies in a type-safe registry
type Container struct {
	registry map[string]any
	mu       sync.RWMutex
}

// NewContainer initializes and registers all dependencies
func NewContainer() *Container {
	c := &Container{registry: make(map[string]any)}

	logger := logging.New()
	c.Register("logger", logger)
	c.Register("errorHandler", errors.NewErrorHandler(logger))
	c.Register("configLoader", config.NewConfigLoader(logger, "."))
	c.Register("driftDetectorServiceFactory", func(
		awsProvider service.InstanceProvider,
		terraformProvider service.InstanceProvider,
		repository service.DriftRepository,
		reporters []service.Reporter,
		config service.DriftDetectorConfig,
		logger *logging.Logger,
	) service.DriftDetectorProvider {
		// This will be implemented in the app package
		// but we register the factory here to avoid circular imports
		return nil // This will be replaced when resolving
	})
	c.Register("instanceProviderFactory", factory.NewInstanceProviderFactory(logger))
	c.Register("driftDetectorFactory", factory.NewDriftDetectorFactory(logger))
	c.Register("reporterFactory", factory.NewReporterFactory(logger))
	c.Register("repositoryFactory", factory.NewRepositoryFactory(logger))

	return c
}

// Register adds a new dependency
func (c *Container) Register(key string, dep any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.registry[key] = dep
}

// resolve retrieves a registered dependency and casts it to the expected type
func resolve[T any](registry map[string]any, key string) (T, error) {
	val, ok := registry[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("dependency %s not registered", key)
	}

	typed, ok := val.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("dependency %s is not of expected type", key)
	}

	return typed, nil
}

// Resolve provides a public API for retrieving dependencies
func Resolve[T any](c *Container, key string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return resolve[T](c.registry, key)
}

// GetDriftDetectorServiceFactory returns the drift detector service factory
func (c *Container) GetDriftDetectorServiceFactory() (DriftDetectorServiceFactory, error) {
	factory := GetRegisteredDriftDetectorServiceFactory()
	if factory == nil {
		var zero DriftDetectorServiceFactory
		return zero, fmt.Errorf("drift detector service factory not registered")
	}
	return factory, nil
}

// Convenience getter for CLI handler
func (c *Container) GetCLIHandler(ctx context.Context, application service.DriftDetectorProvider, cfg *config.Config) CLIHandlerProvider {
	logger, _ := Resolve[*logging.Logger](c, "logger")
	configLoader, _ := Resolve[*config.ConfigLoader](c, "configLoader")
	return cli.NewHandler(ctx, application, configLoader, cfg, logger)
}

var (
	driftDetectorServiceFactory DriftDetectorServiceFactory
	driftFactoryMutex           sync.RWMutex
)

// RegisterDriftDetectorServiceFactory registers a function that creates drift detector services
func RegisterDriftDetectorServiceFactory(factory DriftDetectorServiceFactory) {
	driftFactoryMutex.Lock()
	defer driftFactoryMutex.Unlock()
	driftDetectorServiceFactory = factory
}

// GetRegisteredDriftDetectorServiceFactory returns the registered drift detector service factory
func GetRegisteredDriftDetectorServiceFactory() DriftDetectorServiceFactory {
	driftFactoryMutex.RLock()
	defer driftFactoryMutex.RUnlock()
	return driftDetectorServiceFactory
}
