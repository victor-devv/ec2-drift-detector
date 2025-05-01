package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/victor-devv/ec2-drift-detector/internal/app"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/container"
)

func main() {
	// Initialize the dependency container
	c := container.NewContainer()

	if err := run(c); err != nil {
		handler, _ := container.Resolve[*errors.ErrorHandler](c, "errorHandler")
		handler.HandleWithExit(err)
	}
}

func run(c *container.Container) error {
	// Create context & close context on graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load configuration
	configLoader, err := container.Resolve[*config.ConfigLoader](c, "configLoader")
	if err != nil {
		return err
	}

	cfg, err := configLoader.Load()
	if err != nil {
		return err
	}

	application, err := app.InitializeApplication(ctx, c, cfg)
	if err != nil {
		return err
	}

	// Execute CLI
	if err := c.GetCLIHandler(ctx, application.DriftDetector, cfg).Execute(ctx); err != nil {
		return err
	}

	return nil
}
