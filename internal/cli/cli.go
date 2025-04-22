/*
Parses CLI arguments using Go's flag package.

Overrides values in config with CLI-supplied ones.

Supports:

	--state-file

	--attributes

	--output, --output-file

	--verbose, --concurrent
*/
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

type CLI struct {
	config *config.Config
	logger *logrus.Logger
}

// NewCLI creates a new CLI instance
func NewCLI(cfg *config.Config, logger *logrus.Logger) *CLI {
	return &CLI{
		config: cfg,
		logger: logger,
	}
}

// Parse parses the command-line arguments and applies them to the configuration
func (c *CLI) Parse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided; use --help for usage")
	}

	flags := flag.NewFlagSet("drift-detector", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var verbose bool
	attributesStr := flags.String("attributes", strings.Join(c.config.Detector.Attributes, ","), "Comma-separated list of attributes to check for drift")
	flags.StringVar(&c.config.Terraform.StateFile, "state-file", c.config.Terraform.StateFile, "Path to Terraform state file or HCL directory (required)")
	flags.BoolVar(&c.config.Concurrent, "concurrent", c.config.Concurrent, "Enable concurrent processing")
	flags.StringVar(&c.config.Detector.OutputFormat, "output", c.config.Detector.OutputFormat, "Output format (console, json)")
	flags.StringVar(&c.config.Detector.OutputFile, "output-file", c.config.Detector.OutputFile, "Output file path (defaults to stdout)")
	flags.BoolVar(&verbose, "verbose", false, "Enable verbose logging")

	if err := flags.Parse(args); err != nil {
		return err
	}

	// Validate required flag
	if strings.TrimSpace(c.config.Terraform.StateFile) == "" {
		return fmt.Errorf("--state-file is required")
	}

	// Parse attribute list
	if attrStr := strings.TrimSpace(*attributesStr); attrStr != "" {
		parts := strings.Split(attrStr, ",")
		c.config.Detector.Attributes = make([]string, 0, len(parts))
		for _, attr := range parts {
			trimmed := strings.TrimSpace(attr)
			if trimmed != "" {
				c.config.Detector.Attributes = append(c.config.Detector.Attributes, trimmed)
			}
		}
	}

	// Set verbose logging
	if verbose {
		c.logger.SetLevel(logrus.DebugLevel)
		c.logger.Debug("Verbose logging enabled")
	}

	return nil
}

// LogStartupInfo logs config summary before execution
func (c *CLI) LogStartupInfo(ctx context.Context) error {
	c.logger.Info("Starting drift detection")
	c.logger.Infof("Terraform path: %s", c.config.Terraform.StateFile)
	c.logger.Infof("Attributes to check: %v", c.config.Detector.Attributes)
	c.logger.Infof("Output format: %s", c.config.Detector.OutputFormat)
	if c.config.Detector.OutputFile != "" {
		c.logger.Infof("Output will be written to: %s", c.config.Detector.OutputFile)
	}
	return nil
}

// Config returns the populated configuration
func (c *CLI) Config() *config.Config {
	return c.config
}
