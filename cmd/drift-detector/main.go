/*
Package main serves as the entrypoint of the application.

It bootstraps config, logger, CLI parser, AWS client, Terraform parser, drift detector, and result reporter.

It executes either sequential or concurrent form.
*/
package main

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/app"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/repository"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/terraform"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/cli"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/reporter"
)

func main() {
	if err := run(); err != nil {
		// Create error handler
		logger := logging.New()
		errorHandler := errors.NewErrorHandler(logger)
		errorHandler.HandleWithExit(err)
	}
}

func run() error {
	// close context on graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize logger
	logger := logging.New()

	// Load configuration
	configLoader := config.NewConfigLoader(logger, ".")
	cfg, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize AWS client
	awsClient, err := aws.NewClient(ctx, aws.ClientConfig{
		Region:        cfg.AWS.Region,
		Profile:       cfg.AWS.Profile,
		AccessKey:     cfg.AWS.AccessKeyID,
		SecretKey:     cfg.AWS.SecretAccessKey,
		Endpoint:      cfg.AWS.Endpoint,
		UseLocalstack: strings.ToLower(cfg.App.Env) == "dev" || strings.ToLower(cfg.App.Env) == "development",
	}, logger)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Initialize EC2 service
	ec2Service := aws.NewEC2Service(logger, awsClient)

	// Initialize Terraform client
	terraformClient, err := terraform.NewClient(terraform.ClientConfig{
		StateFile: cfg.Terraform.StateFile,
		HCLDir:    cfg.Terraform.HCLDir,
		UseHCL:    cfg.Terraform.UseHCL,
	}, logger)
	if err != nil {
		return fmt.Errorf("failed to create Terraform parser: %w", err)
	}

	// Initialize repository
	driftRepo := repository.NewInMemoryDriftRepository(logger)

	// Initialize reporters
	var reporters []service.Reporter

	switch cfg.Reporter.Type {
	case "console":
		reporters = append(reporters, reporter.NewConsoleReporter(logger))
	case "json":
		reporters = append(reporters, reporter.NewJSONReporter(logger, cfg.Reporter.OutputFile))
	case "both":
		reporters = append(reporters, reporter.NewConsoleReporter(logger))
		reporters = append(reporters, reporter.NewJSONReporter(logger, cfg.Reporter.OutputFile))
	default:
		logger.Warn("Unknown reporter type: %s, using console reporter", cfg.Reporter.Type)
		reporters = append(reporters, reporter.NewConsoleReporter(logger))
	}

	// Initialize drift detector service
	driftDetector := app.NewDriftDetectorService(
		ec2Service,
		terraformClient,
		driftRepo,
		reporters,
		app.DriftDetectorConfig{
			SourceOfTruth:      model.ResourceOrigin(cfg.Detector.SourceOfTruth),
			AttributePaths:     cfg.Detector.Attributes,
			ParallelChecks:     cfg.Detector.ParallelChecks,
			Timeout:            time.Duration(cfg.Detector.TimeoutSeconds) * time.Second,
			ScheduleExpression: cfg.App.ScheduleExpression,
		},
		logger,
	)

	// Initialize application
	application := app.NewApplication(driftDetector)

	// Initialize CLI handler
	cliHandler := cli.NewHandler(application, configLoader, cfg, logger)

	// Execute CLI
	if err := cliHandler.Execute(); err != nil {
		return err
	}

	return nil
}
