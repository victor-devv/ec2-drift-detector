package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/victor-devv/ec2-drift-detector/internal/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/cli"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/detector"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
	"github.com/victor-devv/ec2-drift-detector/internal/reporter"
	"github.com/victor-devv/ec2-drift-detector/internal/terraform"
	"github.com/victor-devv/ec2-drift-detector/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// close context on graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	logger := logger.New(cfg)

	app := cli.NewCLI(cfg, logger)
	if err := app.Parse(os.Args[1:]); err != nil {
		return err
	}

	// refresh config
	cfg = app.Config()
	app.PreRun(ctx)

	awsClient, err := aws.NewClient(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	ec2Client := aws.NewEC2Client(awsClient, cfg, logger)

	tfParser, err := terraform.GetParser(logger, cfg.Terraform.StateFile)
	if err != nil {
		return fmt.Errorf("failed to create Terraform parser: %w", err)
	}

	ec2Detector := detector.NewEC2Detector(*ec2Client, tfParser, logger)

	// Detect drift
	var results []models.DriftResult
	if cfg.Concurrent {
		logger.Info("Running concurrent drift detection")
		results, err = ec2Detector.DetectDriftConcurrent(ctx, cfg.Detector.Attributes)
	} else {
		logger.Info("Running sequential drift detection")
		results, err = ec2Detector.DetectDrift(ctx, cfg.Detector.Attributes)
	}
	if err != nil {
		return fmt.Errorf("drift detection failed: %w", err)
	}

	// Initialize reporter
	var rep reporter.Reporter
	switch cfg.Detector.OutputFormat {
	case "json":
		if cfg.Detector.OutputFile != "" {
			rep = reporter.NewJSONReporter(logger).WithOutputFile(cfg.Detector.OutputFile)
		} else {
			rep = reporter.NewJSONReporter(logger)
		}
	default:
		rep = reporter.NewConsoleReporter(logger)
	}

	if err := rep.Report(ctx, results); err != nil {
		return fmt.Errorf("reporting failed: %w", err)
	}

	logger.Info("Drift detection completed")

	return nil
}
