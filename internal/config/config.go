package config

import (
	"sync"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
)

// Config holds all application configuration
type Config struct {
	// General application configuration
	App struct {
		Env string `mapstructure:"env"`

		LogLevel logging.LogLevel `mapstructure:"log_level"`
		// JSONLogs indicates whether to output logs in JSON format
		JSONLogs bool `mapstructure:"json_logs"`
		// ScheduleExpression is the cron expression for scheduled drift checks
		ScheduleExpression string `mapstructure:"schedule_expression"`
	} `mapstructure:"app"`

	AWS struct {
		DefaultRegion   string `mapstructure:"region"`
		AccessKeyID     string `mapstructure:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key"`
		Profile         string `mapstructure:"profile"`
		Endpoint        string `mapstructure:"endpoint"`
	} `mapstructure:"aws"`

	Terraform struct {
		StateFile string `mapstructure:"state_file"`
		HCLDir    string `mapstructure:"hcl_dir"`
		UseHCL    bool   `mapstructure:"use_hcl"`
	} `mapstructure:"terraform"`

	Detector struct {
		Attributes    []string `mapstructure:"attributes"`
		SourceOfTruth string   `mapstructure:"source_of_truth"`
		// ParallelChecks is the number of parallel checks to run
		ParallelChecks int `mapstructure:"parallel_checks"`
		// TimeoutSeconds is the timeout for drift detection operations
		TimeoutSeconds int `mapstructure:"timeout_seconds"`
	} `mapstructure:"drift_detection"`

	// Reporter configuration
	Reporter struct {
		// Type is the type of reporter to use (json, console, or both)
		Type string `mapstructure:"type"`
		// OutputFile is the path to the output file for JSON reporter
		OutputFile string `mapstructure:"output_file"`
		// PrettyPrint indicates whether to format JSON output
		PrettyPrint bool `mapstructure:"pretty_print"`
	} `mapstructure:"reporter"`

	// mutex for thread safety
	mu sync.RWMutex
}

// Validate validates the configuration
func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.AWS.DefaultRegion == "" {
		return errors.NewValidationError("AWS region cannot be empty")
	}

	if c.Terraform.UseHCL {
		if c.Terraform.HCLDir == "" {
			return errors.NewValidationError("Terraform HCL directory cannot be empty when UseHCL is true")
		}
	} else {
		if c.Terraform.StateFile == "" {
			return errors.NewValidationError("Terraform state file cannot be empty when UseHCL is false")
		}
	}

	if len(c.Detector.Attributes) == 0 {
		return errors.NewValidationError("At least one attribute must be specified for drift detection")
	}

	if c.Detector.SourceOfTruth != "aws" && c.Detector.SourceOfTruth != "terraform" {
		return errors.NewValidationError("Source of truth must be either 'aws' or 'terraform'")
	}

	if c.Detector.ParallelChecks <= 0 {
		return errors.NewValidationError("Parallel checks must be greater than 0")
	}

	if c.Detector.TimeoutSeconds <= 0 {
		return errors.NewValidationError("Timeout seconds must be greater than 0")
	}

	if c.Reporter.Type != "json" && c.Reporter.Type != "console" && c.Reporter.Type != "both" {
		return errors.NewValidationError("Reporter type must be 'json', 'console', or 'both'")
	}

	if (c.Reporter.Type == "json" || c.Reporter.Type == "both") && c.Reporter.OutputFile == "" {
		return errors.NewValidationError("Output file must be specified for JSON reporter")
	}

	// Validate schedule expression if provided
	if c.App.ScheduleExpression != "" {
		if len(c.App.ScheduleExpression) < 9 { // Minimum valid cron is "* * * * *"
			return errors.NewValidationError("Invalid schedule expression format")
		}
	}

	return nil
}

// GetSourceOfTruth returns the source of truth configuration
func (c *Config) GetSourceOfTruth() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Detector.SourceOfTruth
}

// SetSourceOfTruth sets the source of truth configuration
func (c *Config) SetSourceOfTruth(sourceOfTruth string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Detector.SourceOfTruth = sourceOfTruth
}

// GetAttributes returns the attributes to check for drift
func (c *Config) GetAttributes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Detector.Attributes
}

// SetAttributes sets the attributes to check for drift
func (c *Config) SetAttributes(attributes []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Detector.Attributes = attributes
}

// GetParallelChecks returns the number of parallel checks to run
func (c *Config) GetParallelChecks() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Detector.ParallelChecks
}

// SetParallelChecks sets the number of parallel checks to run
func (c *Config) SetParallelChecks(parallelChecks int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Detector.ParallelChecks = parallelChecks
}

// GetTimeout returns the timeout for drift detection operations
func (c *Config) GetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.Detector.TimeoutSeconds) * time.Second
}

// SetTimeout sets the timeout for drift detection operations
func (c *Config) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Detector.TimeoutSeconds = int(timeout.Seconds())
}

// GetReporterType returns the reporter type
func (c *Config) GetReporterType() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Reporter.Type
}

// SetReporterType sets the reporter type
func (c *Config) SetReporterType(reporterType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Reporter.Type = reporterType
}

// GetScheduleExpression returns the schedule expression
func (c *Config) GetScheduleExpression() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.App.ScheduleExpression
}

// SetScheduleExpression sets the schedule expression
func (c *Config) SetScheduleExpression(expression string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.App.ScheduleExpression = expression
}
