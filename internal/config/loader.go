package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
)

// ConfigLoader is responsible for loading application configuration
type ConfigLoader struct {
	viper     *viper.Viper
	config    *Config
	logger    *logging.Logger
	configDir string
	mu        sync.Mutex
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(configDir string, logger *logging.Logger) *ConfigLoader {
	return &ConfigLoader{
		viper:     viper.New(),
		config:    &Config{},
		logger:    logger,
		configDir: configDir,
	}
}

// Load loads the configuration from files, environment variables, and defaults
func (l *ConfigLoader) Load() (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.setDefaults()

	if err := l.loadFromFile(); err != nil {
		// If no config file is found, just log a warning
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.NewSystemError("Failed to load configuration from file", err)
		}
		l.logger.Warn("No configuration file found, using defaults and environment variables")
	}

	l.loadFromEnv()

	if err := l.viper.Unmarshal(l.config); err != nil {
		return nil, errors.NewSystemError("Failed to unmarshal configuration", err)
	}

	if err := l.config.Validate(); err != nil {
		return nil, err
	}

	// Set up logging based on configuration
	logging.ConfigureLogger(logging.LogConfig{
		Level:      l.config.App.LogLevel,
		JSONFormat: l.config.App.JSONLogs,
	})

	return l.config, nil
}

// setDefaults sets default configuration values
func (l *ConfigLoader) setDefaults() {
	v := l.viper

	// App defaults
	v.SetDefault("app.env", "Dev")
	v.SetDefault("app.log_level", "INFO")
	v.SetDefault("app.json_logs", false)
	v.SetDefault("app.schedule_expression", "0 */6 * * *") // Run every 6 hours by default

	// AWS defaults
	v.SetDefault("aws.region", "eu-north-1")
	v.SetDefault("aws.use_localstack", false)

	// Terraform defaults
	v.SetDefault("terraform.use_hcl", false)

	// DriftDetection defaults
	v.SetDefault("drift_detection.attributes", []string{"instance_type", "ami", "vpc_security_group_ids", "tags"})
	v.SetDefault("drift_detection.source_of_truth", "terraform")
	v.SetDefault("drift_detection.parallel_checks", 5)
	v.SetDefault("drift_detection.timeout_seconds", 60)

	// Reporter defaults
	v.SetDefault("reporter.type", "console")
	v.SetDefault("reporter.pretty_print", true)
}

// loadFromFile loads configuration from file
func (l *ConfigLoader) loadFromFile() error {
	v := l.viper

	// Config file search paths
	configDirs := []string{
		l.configDir,
		".",
		"./config",
		"/etc/drift-detector",
		filepath.Join(getUserHomeDir(), ".drift-detector"),
	}

	// Supported config file names
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add search paths
	for _, dir := range configDirs {
		if dir != "" {
			v.AddConfigPath(dir)
		}
	}

	// Read the config file
	return v.ReadInConfig()
}

// loadFromEnv loads configuration from environment variables
func (l *ConfigLoader) loadFromEnv() {
	v := l.viper

	// Set environment variable prefix
	v.SetEnvPrefix("DRIFT")

	// Replace dots with underscores in env vars (app.log_level -> DRIFT_APP_LOG_LEVEL)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Automatically bind env vars to config keys
	v.AutomaticEnv()
}

// UpdateConfig updates the configuration with command-line flags
func (l *ConfigLoader) UpdateConfig(cfg *Config, cliOpts map[string]interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Update specific fields based on CLI options
	for key, value := range cliOpts {
		if value == nil {
			continue
		}

		switch key {
		case "log_level":
			if logLevel, ok := value.(string); ok && logLevel != "" {
				cfg.App.LogLevel = logging.LogLevel(strings.ToUpper(logLevel))

				l.logger.SetLogLevel(cfg.App.LogLevel)
			}
		case "attributes":
			if attrs, ok := value.([]string); ok && len(attrs) > 0 {
				cfg.SetAttributes(attrs)
			}
		case "source_of_truth":
			if sourceOfTruth, ok := value.(string); ok && sourceOfTruth != "" {
				cfg.SetSourceOfTruth(sourceOfTruth)
			}
		case "parallel_checks":
			if parallelChecks, ok := value.(int); ok && parallelChecks > 0 {
				cfg.SetParallelChecks(parallelChecks)
			}
		case "state_file":
			if stateFile, ok := value.(string); ok && stateFile != "" {
				cfg.Terraform.StateFile = stateFile
				cfg.Terraform.UseHCL = false
			}
		case "hcl_dir":
			if hclDir, ok := value.(string); ok && hclDir != "" {
				cfg.Terraform.HCLDir = hclDir
				cfg.Terraform.UseHCL = true
			}
		case "reporter_type":
			if reporterType, ok := value.(string); ok && reporterType != "" {
				cfg.SetReporterType(reporterType)
			}
		case "output_file":
			if outputFile, ok := value.(string); ok && outputFile != "" {
				cfg.Reporter.OutputFile = outputFile
			}
		case "aws_region":
			if region, ok := value.(string); ok && region != "" {
				cfg.AWS.DefaultRegion = region
			}
		case "schedule_expression":
			if expr, ok := value.(string); ok && expr != "" {
				cfg.SetScheduleExpression(expr)
			}
		}
	}

	// Validate the updated configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	return nil
}

// getUserHomeDir returns the current user's home directory
func getUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return homeDir
}

// ReloadConfig reloads the configuration from file
func (l *ConfigLoader) ReloadConfig() (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.viper.ReadInConfig(); err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("Failed to reload configuration: %v", err), err)
	}

	if err := l.viper.Unmarshal(l.config); err != nil {
		return nil, errors.NewSystemError("Failed to unmarshal configuration", err)
	}

	if err := l.config.Validate(); err != nil {
		return nil, err
	}

	return l.config, nil
}
