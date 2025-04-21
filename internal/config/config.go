package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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

	AWS struct {
		DefaultRegion   string `mapstructure:"default_region"`
		AccessKeyID     string `mapstructure:"access_key_id"`
		SecretAccessKey string `mapstructure:"secret_access_key"`
		Endpoint        string `mapstructure:"endpoint"` // For LocalStack
	} `mapstructure:"aws"`

	Terraform struct {
		StateFile  string `mapstructure:"state_file"`
		ConfigFile string `mapstructure:"config_file"`
	} `mapstructure:"terraform"`

	Detector struct {
		Attributes []string `mapstructure:"attributes"`
	} `mapstructure:"detector"`

	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
}

func setDefaults() {
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.endpoint", "")

	viper.SetDefault("terraform.state_file", "terraform.tfstate")
	viper.SetDefault("terraform.config_file", "main.tf")

	viper.SetDefault("detector.attributes", []string{
		"instance_type",
		"tags",
		"security_groups.group_ids",
	})

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
}

// LoadConfig loads configuration from file and environment variables
func New(path string) (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // or viper.SetConfigType("json")

	if path != "" {
		// Use config file from the specified path
		viper.SetConfigFile(path)
	} else {
		// Search config in home directory and current directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		viper.AddConfigPath(filepath.Join(home, ".ec2-drift-detector"))
		viper.AddConfigPath(".")
	}

	// Set environment variable prefix and enable automatic env var binding
	viper.SetEnvPrefix("EC2_DRIFT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults()

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
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
