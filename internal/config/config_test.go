package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

func TestConfigAccessors(t *testing.T) {
	cfg := &config.Config{}

	cfg.SetEnv("dev")
	assert.Equal(t, "dev", cfg.GetEnv())

	cfg.SetLogLevel("DEBUG")
	assert.Equal(t, logging.LogLevel("DEBUG"), cfg.GetLogLevel())

	cfg.SetJSONLogs(true)
	assert.True(t, cfg.GetJSONLogs())

	cfg.SetScheduleExpression("0 */6 * * *")
	assert.Equal(t, "0 */6 * * *", cfg.GetScheduleExpression())

	cfg.SetAWSRegion("us-east-1")
	cfg.SetAWSAccessKeyID("key")
	cfg.SetAWSSecretAccessKey("secret")
	cfg.SetAWSProfile("profile")
	cfg.SetAWSEndpoint("http://localhost:4566")
	assert.Equal(t, "us-east-1", cfg.GetAWSRegion())
	assert.Equal(t, "key", cfg.GetAWSAccessKeyID())
	assert.Equal(t, "secret", cfg.GetAWSSecretAccessKey())
	assert.Equal(t, "profile", cfg.GetAWSProfile())
	assert.Equal(t, "http://localhost:4566", cfg.GetAWSEndpoint())

	cfg.SetStateFile("terraform.tfstate")
	cfg.SetUseHCL(true)
	assert.Equal(t, "terraform.tfstate", cfg.GetStateFile())
	assert.True(t, cfg.GetUseHCL())

	cfg.SetSourceOfTruth("terraform")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetParallelChecks(3)
	cfg.SetTimeout(45 * time.Second)
	assert.Equal(t, "terraform", cfg.GetSourceOfTruth())
	assert.Equal(t, []string{"instance_type"}, cfg.GetAttributes())
	assert.Equal(t, 3, cfg.GetParallelChecks())
	assert.Equal(t, 45*time.Second, cfg.GetTimeout())

	cfg.SetReporterType(config.ReporterTypeJSON)
	cfg.SetOutputFile("report.json")
	cfg.SetPrettyPrint(true)
	assert.Equal(t, config.ReporterTypeJSON, cfg.GetReporterType())
	assert.Equal(t, "report.json", cfg.GetOutputFile())
	assert.True(t, cfg.GetPrettyPrint())
}

func TestConfigValidation(t *testing.T) {
	cfg := &config.Config{}

	// Set minimal valid values
	cfg.SetAWSRegion("us-east-1")
	cfg.SetUseHCL(false)
	cfg.SetStateFile("terraform.tfstate")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetSourceOfTruth("aws")
	cfg.SetParallelChecks(1)
	cfg.SetTimeout(10 * time.Second)
	cfg.SetReporterType(config.ReporterTypeConsole)

	err := cfg.Validate()
	assert.NoError(t, err)

	cfg.SetSourceOfTruth("invalid")
	err = cfg.Validate()
	assert.ErrorContains(t, err, "Source of truth must be either")
}
