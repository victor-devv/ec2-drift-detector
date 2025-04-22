package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/victor-devv/ec2-drift-detector/internal/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
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

	awsClient, err := aws.NewClient(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	ec2Client := aws.NewEC2Client(awsClient, cfg, logger)

	tfParser, err := terraform.GetParser(logger, cfg.Terraform.StateFile)
	if err != nil {
		return fmt.Errorf("failed to create Terraform parser: %w", err)
	}

	return nil
}
