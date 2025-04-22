/*
Package config loads environment variables using caarlos0/env.

Validates required fields like AWS_DEFAULT_REGION and Terraform state file location.

Groups logical config segments (AWS, Terraform, Detector, Logging).
*/
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Env string

const (
	Env_Test    Env = "test"
	Env_Dev     Env = "dev"
	Env_Prod    Env = "production"
	Env_Staging Env = "staging"
)

// Config represents the application configuration
type Config struct {
	Env Env `env:"ENV" envDefault:"dev"`

	Concurrent bool `env:"CONCURRENT" envDefault:"true"`

	AWS struct {
		DefaultRegion   string `env:"AWS_DEFAULT_REGION" envDefault:"eu-north-1"`
		AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY "`
		Ec2Endpoint     string `env:"AWS_EC2_ENDPOINT"`
	}

	Terraform struct {
		StateFile  string `mapstructure:"TERRAFORM_STATE_FILE" envDefault:"terraform/terraform.tfstate"`
		ConfigFile string `mapstructure:"TERRAFORM_CONFIG_FILE" envDefault:"terraform/main.tf"`
	}

	Detector struct {
		Attributes   []string `env:"DRIFT_ATTRIBUTES" envDefault:"instance_type,ami,subnet_id,vpc_security_group_ids,tags"`
		OutputFormat string   `env:"DRIFT_OUTPUT_FORMAT"`
		OutputFile   string   `env:"DRIFT_OUTPUT_FILE"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL"`
	}
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

// checks if the configuration is valid
func (c *Config) Validate() error {
	if c.AWS.DefaultRegion == "" {
		return ErrMissingRegion
	}

	if c.Terraform.StateFile == "" && c.Terraform.ConfigFile == "" {
		return ErrMissingTerraformInput
	}

	return nil
}

// Errors
var (
	ErrMissingRegion         = newConfigError("AWS region is required")
	ErrMissingTerraformInput = newConfigError("Either Terraform state file or config file is required")
)

type configError struct {
	msg string
}

func newConfigError(msg string) error {
	return &configError{msg}
}

func (e *configError) Error() string {
	return e.msg
}
