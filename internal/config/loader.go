package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

type rawConfig struct {
	App struct {
		Env                string `mapstructure:"env"`
		LogLevel           string `mapstructure:"log_level"`
		JSONLogs           bool   `mapstructure:"json_logs"`
		ScheduleExpression string `mapstructure:"schedule_expression"`
	} `mapstructure:"app"`

	AWS struct {
		Region          string `mapstructure:"region"`
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
		Attributes     []string `mapstructure:"attributes"`
		SourceOfTruth  string   `mapstructure:"source_of_truth"`
		ParallelChecks int      `mapstructure:"parallel_checks"`
		TimeoutSeconds int      `mapstructure:"timeout_seconds"`
	} `mapstructure:"detector"`

	Reporter struct {
		Type        string `mapstructure:"type"`
		OutputFile  string `mapstructure:"output_file"`
		PrettyPrint bool   `mapstructure:"pretty_print"`
	} `mapstructure:"reporter"`
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(logger *logging.Logger, configDir string) *ConfigLoader {
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

	// First try to load from config file
	if err := l.loadFromFile(); err != nil {
		// If no config file is found, just log a warning
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.NewSystemError("Failed to load configuration from file", err)
		}
		l.logger.Warn("No configuration file found, will check for .envrc and environment variables")
	}

	// Try to load from .envrc file if it exists
	if err := l.loadFromEnvrcFile(); err != nil {
		l.logger.Warn(fmt.Sprintf("Failed to load configuration from .envrc file: %v", err))
		// Continue with other sources even if .envrc fails
	}

	// Load from environment variables
	l.loadFromEnv()

	var raw rawConfig
	if err := l.viper.Unmarshal(&raw); err != nil {
		return nil, errors.NewSystemError("Failed to unmarshal configuration", err)
	}
	applyRawToConfig(raw, l.config)

	// Set up logging based on configuration
	logging.ConfigureLogger(logging.LogConfig{
		Level:      l.config.app.logLevel,
		JSONFormat: l.config.app.jsonLogs,
	})

	l.logger.Info("Configuration loaded successfully")

	return l.config, nil
}

// setDefaults sets default configuration values
func (l *ConfigLoader) setDefaults() {
	v := l.viper

	// App defaults
	v.SetDefault("app.env", AppEnvDev)
	v.SetDefault("app.log_level", LogLevelInfo)
	v.SetDefault("app.json_logs", false)
	v.SetDefault("app.schedule_expression", cronEvery6Hours) // Run every 6 hours by default

	// AWS defaults
	v.SetDefault("aws.region", aWSDefaultRegion)
	v.SetDefault("aws.access_key_id", "")
	v.SetDefault("aws.secret_access_key", "")
	v.SetDefault("aws.profile", "")
	v.SetDefault("aws.endpoint", "")

	// Terraform defaults
	v.SetDefault("terraform.state_file", "")
	v.SetDefault("terraform.hcl_dir", "")
	v.SetDefault("terraform.use_hcl", false)

	// DriftDetection defaults
	v.SetDefault("detector.attributes", []string{"instance_type", "ami", "vpc_security_group_ids", "tags"})
	v.SetDefault("detector.source_of_truth", defaultSourceOfTruth)
	v.SetDefault("detector.parallel_checks", 5)
	v.SetDefault("detector.timeout_seconds", 60)

	// Reporter defaults
	v.SetDefault("reporter.type", ReporterTypeConsole)
	v.SetDefault("reporter.output_file", "")
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

// loadFromEnvrcFile loads configuration from .envrc file
func (l *ConfigLoader) loadFromEnvrcFile() error {
	// Check for .envrc in current directory and parent directories
	envrcPath, err := findEnvrcFile(".")
	if err != nil {
		return err
	}

	if envrcPath == "" {
		return fmt.Errorf(".envrc file not found")
	}

	l.logger.Info(fmt.Sprintf("Loading configuration from .envrc file: %s", envrcPath))

	// Open the .envrc file
	file, err := os.Open(envrcPath)
	if err != nil {
		return errors.NewOperationalError(fmt.Sprintf("Failed to open .envrc file: %s", envrcPath), err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Process export statements
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				value = strings.Trim(value, `"'`)

				// Only process DRIFT_ prefixed variables
				if strings.HasPrefix(key, "DRIFT_") {
					// Set the variable as an environment variable
					os.Setenv(key, value)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.NewOperationalError(fmt.Sprintf("Error reading .envrc file: %s", envrcPath), err)
	}

	return nil
}

// findEnvrcFile searches for a .envrc file in the current and parent directories
func findEnvrcFile(startDir string) (string, error) {
	// Get absolute path of starting directory
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", errors.NewOperationalError("Failed to get absolute path", err)
	}

	// Start with the current directory
	currentDir := absPath

	for {
		// Check if .envrc exists in the current directory
		envrcPath := filepath.Join(currentDir, ".envrc")
		if _, err := os.Stat(envrcPath); err == nil {
			return envrcPath, nil
		}

		// Get the parent directory
		parentDir := filepath.Dir(currentDir)

		// If we're at the root, stop searching
		if parentDir == currentDir {
			break
		}

		currentDir = parentDir
	}

	// No .envrc file found
	return "", nil
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
		case "log-level":
			if logLevel, ok := value.(string); ok && logLevel != "" {
				cfg.SetLogLevel(logging.LogLevel(strings.ToUpper(logLevel)))
				l.logger.SetLogLevel(cfg.app.logLevel)
			}
		case "attributes":
			if attrs, ok := value.([]string); ok && len(attrs) > 0 {
				cfg.SetAttributes(attrs)
			}
		case "source-of-truth":
			if sourceOfTruth, ok := value.(string); ok && sourceOfTruth != "" {
				cfg.SetSourceOfTruth(sourceOfTruth)
			}
		case "parallel-checks":
			if parallelChecks, ok := value.(int); ok && parallelChecks > 0 {
				cfg.SetParallelChecks(parallelChecks)
			}
		case "state-file":
			if stateFile, ok := value.(string); ok && stateFile != "" {
				cfg.SetStateFile(stateFile)
				cfg.SetUseHCL(false)
			}
		case "hcl-dir":
			if hclDir, ok := value.(string); ok && hclDir != "" {
				cfg.SetHCLDir(hclDir)
				cfg.SetUseHCL(true)
			}
		case "output":
			if reporterType, ok := value.(string); ok && reporterType != "" {
				cfg.SetReporterType(reporterType)
			}
		case "output-file":
			if outputFile, ok := value.(string); ok && outputFile != "" {
				cfg.SetOutputFile(outputFile)
			}
		case "aws-region":
			if region, ok := value.(string); ok && region != "" {
				cfg.SetAWSRegion(region)
			}
		case "schedule-expression":
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

	if err := l.loadFromFile(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.NewSystemError("Failed to load configuration from file", err)
		}
		l.logger.Warn("No configuration file found, will check for .envrc and environment variables")
	}

	// Try to load from .envrc file if it exists
	if err := l.loadFromEnvrcFile(); err != nil {
		l.logger.Warn(fmt.Sprintf("Failed to load configuration from .envrc file: %v", err))
	}

	// Load from environment variables
	l.loadFromEnv()

	var raw rawConfig
	if err := l.viper.Unmarshal(&raw); err != nil {
		return nil, errors.NewSystemError("Failed to unmarshal configuration", err)
	}
	applyRawToConfig(raw, l.config)

	if err := l.config.Validate(); err != nil {
		return nil, err
	}

	return l.config, nil
}

func applyRawToConfig(raw rawConfig, c *Config) {
	c.SetEnv(raw.App.Env)
	c.SetLogLevel(logging.LogLevel(strings.ToUpper(raw.App.LogLevel)))
	c.SetJSONLogs(raw.App.JSONLogs)
	c.SetScheduleExpression(raw.App.ScheduleExpression)

	c.SetAWSRegion(raw.AWS.Region)
	c.SetAWSAccessKeyID(raw.AWS.AccessKeyID)
	c.SetAWSSecretAccessKey(raw.AWS.SecretAccessKey)
	c.SetAWSProfile(raw.AWS.Profile)
	c.SetAWSEndpoint(raw.AWS.Endpoint)

	c.SetStateFile(raw.Terraform.StateFile)
	c.SetHCLDir(raw.Terraform.HCLDir)
	c.SetUseHCL(raw.Terraform.UseHCL)

	c.SetAttributes(raw.Detector.Attributes)
	c.SetSourceOfTruth(raw.Detector.SourceOfTruth)
	c.SetParallelChecks(raw.Detector.ParallelChecks)
	c.SetTimeout(time.Duration(raw.Detector.TimeoutSeconds) * time.Second)

	c.SetReporterType(raw.Reporter.Type)
	c.SetOutputFile(raw.Reporter.OutputFile)
	c.SetPrettyPrint(raw.Reporter.PrettyPrint)
}
