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

// Convenience getter for CLI handler
func (c *Container) GetCLIHandler(ctx context.Context, application service.DriftDetectorProvider, cfg *config.Config) CLIHandlerProvider {
	logger, _ := Resolve[*logging.Logger](c, "logger")
	configLoader, _ := Resolve[*config.ConfigLoader](c, "configLoader")
	return cli.NewHandler(ctx, application, configLoader, cfg, logger)
}
