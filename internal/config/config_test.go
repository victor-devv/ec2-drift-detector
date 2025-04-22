package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

func setEnv(key, value string) func() {
	original := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		os.Setenv(key, original)
	}
}

func unsetEnv(key string) func() {
	original := os.Getenv(key)
	os.Unsetenv(key)
	return func() {
		os.Setenv(key, original)
	}
}

func TestNewConfig_Defaults(t *testing.T) {
	defer unsetEnv("ENV")()
	cfg, err := config.New()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, config.Env_Dev, cfg.Env)
	require.Equal(t, "eu-north-1", cfg.AWS.DefaultRegion)
	require.True(t, cfg.Concurrent)
	require.Contains(t, cfg.Detector.Attributes, "instance_type")
}

func TestNewConfig_ParsesEnvVars(t *testing.T) {
	reset1 := setEnv("ENV", "production")
	reset2 := setEnv("AWS_DEFAULT_REGION", "us-east-1")
	defer reset1()
	defer reset2()

	cfg, err := config.New()
	require.NoError(t, err)
	require.Equal(t, config.Env_Prod, cfg.Env)
	require.Equal(t, "us-east-1", cfg.AWS.DefaultRegion)
}

func TestValidateConfig_MissingRegion(t *testing.T) {
	cfg := &config.Config{}
	err := cfg.Validate()
	require.ErrorIs(t, err, config.ErrMissingRegion)
}

func TestValidateConfig_MissingTerraform(t *testing.T) {
	cfg := &config.Config{
		AWS: config.Config{}.AWS,
		Terraform: struct {
			StateFile  string `mapstructure:"TERRAFORM_STATE_FILE" envDefault:"terraform/terraform.tfstate"`
			ConfigFile string `mapstructure:"TERRAFORM_CONFIG_FILE" envDefault:"terraform/main.tf"`
		}{
			StateFile:  "",
			ConfigFile: "",
		},
	}
	cfg.AWS.DefaultRegion = "us-east-1"

	err := cfg.Validate()
	require.ErrorIs(t, err, config.ErrMissingTerraformInput)
}

func TestConfigError_ErrorMethod(t *testing.T) {
	err := config.ErrMissingRegion
	require.Equal(t, "AWS region is required", err.Error())
}
