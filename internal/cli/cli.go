package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

type CLI struct {
	config *config.Config
	logger *logrus.Logger
}

// NewCLI creates a new CLI
func NewCLI(cfg *config.Config, logger *logrus.Logger) *CLI {
	return &CLI{
		config: cfg,
		logger: logger,
	}
}

// Parse parses the command-line arguments
func (c *CLI) Parse(args []string) error {
	var verbose bool

	flags := flag.NewFlagSet("drift-detector", flag.ContinueOnError)

	// Terraform flags
	flags.StringVar(&c.config.Terraform.StateFile, "state-file", "", "Path to Terraform state file or HCL directory (required)")

	// Drift detection flags
	attributesStr := flags.String("attributes", "instance_type,ami,subnet_id,vpc_security_group_ids,tags", "Comma-separated list of attributes to check for drift")
	flags.BoolVar(&c.config.Concurrent, "concurrent", true, "Enable concurrent processing")

	// Output flags
	flags.StringVar(&c.config.Detector.OutputFormat, "output", "console", "Output format (console, json)")
	flags.StringVar(&c.config.Detector.OutputFile, "output-file", "", "Output file path (defaults to stdout)")

	flags.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	// Parse the flags
	if err := flags.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if c.config.Terraform.StateFile == "" {
		return fmt.Errorf("terraform-path is required")
	}

	// Parse the attributes
	if *attributesStr != "" {
		c.config.Detector.Attributes = strings.Split(*attributesStr, ",")
		// Trim whitespace
		for i, attr := range c.config.Detector.Attributes {
			c.config.Detector.Attributes[i] = strings.TrimSpace(attr)
		}
	}

	if verbose {
		c.logger.SetLevel(logrus.DebugLevel)
		c.logger.Debug("Verbose logging enabled")
	}

	return nil
}

// Run executes the CLI
func (c *CLI) PreRun(ctx context.Context) error {
	c.logger.Info("Starting drift detection")
	c.logger.Info(fmt.Sprintf("Terraform path: %s", c.config.Terraform.StateFile))
	c.logger.Info(fmt.Sprintf("Attributes to check: %v", c.config.Detector.Attributes))
	c.logger.Info(fmt.Sprintf("Output format: %s", c.config.Detector.OutputFormat))

	// Return config for the application to use
	return nil
}

// Config returns the configuration
func (c *CLI) Config() *config.Config {
	return c.config
}
