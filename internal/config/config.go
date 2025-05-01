package config

import (
	"sync"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
)

// Config holds all application configuration
// All fields are private and accessed via methods only
type Config struct {
	app       appConfig
	aws       awsConfig
	terraform terraformConfig
	detector  detectorConfig
	reporter  reporterConfig

	mu sync.RWMutex
}

type appConfig struct {
	env                string
	logLevel           logging.LogLevel
	jsonLogs           bool
	scheduleExpression string
}

type awsConfig struct {
	region          string
	accessKeyID     string
	secretAccessKey string
	profile         string
	endpoint        string
}

type terraformConfig struct {
	stateFile string
	hclDir    string
	useHCL    bool
}

type detectorConfig struct {
	attributes     []string
	sourceOfTruth  string
	parallelChecks int
	timeoutSeconds int
}

type reporterConfig struct {
	typeVal     string
	outputFile  string
	prettyPrint bool
}

// ------- App Getters/Setters -------
func (c *Config) GetEnv() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.app.env
}

func (c *Config) SetEnv(env string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.app.env = env
}

func (c *Config) GetLogLevel() logging.LogLevel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.app.logLevel
}

func (c *Config) SetLogLevel(level logging.LogLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.app.logLevel = level
}

func (c *Config) GetJSONLogs() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.app.jsonLogs
}

func (c *Config) SetJSONLogs(val bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.app.jsonLogs = val
}

func (c *Config) GetScheduleExpression() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.app.scheduleExpression
}

func (c *Config) SetScheduleExpression(expr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.app.scheduleExpression = expr
}

// ------- AWS Getters/Setters -------
func (c *Config) GetAWSRegion() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.aws.region
}

func (c *Config) SetAWSRegion(region string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aws.region = region
}

func (c *Config) GetAWSAccessKeyID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.aws.accessKeyID
}

func (c *Config) SetAWSAccessKeyID(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aws.accessKeyID = key
}

func (c *Config) GetAWSSecretAccessKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.aws.secretAccessKey
}

func (c *Config) SetAWSSecretAccessKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aws.secretAccessKey = key
}

func (c *Config) GetAWSProfile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.aws.profile
}

func (c *Config) SetAWSProfile(profile string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aws.profile = profile
}

func (c *Config) GetAWSEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.aws.endpoint
}

func (c *Config) SetAWSEndpoint(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aws.endpoint = endpoint
}

// ------- Terraform Getters/Setters -------
func (c *Config) GetStateFile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.terraform.stateFile
}

func (c *Config) SetStateFile(file string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.terraform.stateFile = file
}

func (c *Config) GetUseHCL() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.terraform.useHCL
}

func (c *Config) SetUseHCL(val bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.terraform.useHCL = val
}

func (c *Config) GetHCLDir() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.terraform.hclDir
}

func (c *Config) SetHCLDir(val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.terraform.hclDir = val
}

// ------- Detector Getters/Setters -------
func (c *Config) GetSourceOfTruth() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.detector.sourceOfTruth
}

func (c *Config) SetSourceOfTruth(val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.detector.sourceOfTruth = val
}

func (c *Config) GetAttributes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.detector.attributes
}

func (c *Config) SetAttributes(val []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.detector.attributes = val
}

func (c *Config) GetParallelChecks() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.detector.parallelChecks
}

func (c *Config) SetParallelChecks(val int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.detector.parallelChecks = val
}

func (c *Config) GetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.detector.timeoutSeconds) * time.Second
}

func (c *Config) SetTimeout(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.detector.timeoutSeconds = int(d.Seconds())
}

// ------- Reporter Getters/Setters -------
func (c *Config) GetReporterType() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reporter.typeVal
}

func (c *Config) SetReporterType(val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reporter.typeVal = val
}

func (c *Config) GetOutputFile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reporter.outputFile
}

func (c *Config) SetOutputFile(val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reporter.outputFile = val
}

func (c *Config) GetPrettyPrint() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reporter.prettyPrint
}

func (c *Config) SetPrettyPrint(val bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reporter.prettyPrint = val
}

// ------- Validation -------
func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.aws.region == "" {
		return errors.NewValidationError("AWS region cannot be empty")
	}

	if c.terraform.useHCL {
		if c.terraform.hclDir == "" {
			return errors.NewValidationError("Terraform HCL directory cannot be empty when UseHCL is true")
		}
	} else {
		if c.terraform.stateFile == "" {
			return errors.NewValidationError("Terraform state file cannot be empty when UseHCL is false")
		}
	}

	if len(c.detector.attributes) == 0 {
		return errors.NewValidationError("At least one attribute must be specified for drift detection")
	}

	if c.detector.sourceOfTruth != "aws" && c.detector.sourceOfTruth != "terraform" {
		return errors.NewValidationError("Source of truth must be either 'aws' or 'terraform'")
	}

	if c.detector.parallelChecks <= 0 {
		return errors.NewValidationError("Parallel checks must be greater than 0")
	}

	if c.detector.timeoutSeconds <= 0 {
		return errors.NewValidationError("Timeout seconds must be greater than 0")
	}

	if c.reporter.typeVal != ReporterTypeConsole && c.reporter.typeVal != ReporterTypeJSON && c.reporter.typeVal != ReporterTypeBoth {
		return errors.NewValidationError("Reporter type must be 'json', 'console', or 'both'")
	}

	// if (c.reporter.typeVal == ReporterTypeJSON || c.reporter.typeVal == ReporterTypeBoth) && c.reporter.outputFile == "" {
	// 	return errors.NewValidationError("Output file must be specified for JSON reporter")
	// }

	if c.app.scheduleExpression != "" && len(c.app.scheduleExpression) < 9 {
		return errors.NewValidationError("Invalid schedule expression format")
	}

	return nil
}
