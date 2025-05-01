package factory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/factory"
)

func newMockConfig() *config.Config {
	cfg := &config.Config{}
	cfg.SetEnv("dev")
	cfg.SetAWSRegion("eu-north-1")
	cfg.SetAWSAccessKeyID("dummy")
	cfg.SetAWSSecretAccessKey("dummy")
	cfg.SetAWSEndpoint("http://localhost:4566")
	cfg.SetUseHCL(false)
	cfg.SetStateFile("./testdata/test.tfstate")
	cfg.SetHCLDir("./testdata")
	return cfg
}

func TestNewInstanceProviderFactory(t *testing.T) {
	logger := logging.New()
	factory := factory.NewInstanceProviderFactory(logger)
	assert.NotNil(t, factory)
}

func TestCreateAWSProvider_Success(t *testing.T) {
	logger := logging.New()
	f := factory.NewInstanceProviderFactory(logger)
	cfg := newMockConfig()

	provider, err := f.CreateAWSProvider(context.Background(), cfg)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateTerraformProvider_Success(t *testing.T) {
	logger := logging.New()
	f := factory.NewInstanceProviderFactory(logger)
	cfg := newMockConfig()

	provider, err := f.CreateTerraformProvider(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestCreateAWSProvider_InvalidRegion(t *testing.T) {
	logger := logging.New()
	f := factory.NewInstanceProviderFactory(logger)
	cfg := newMockConfig()
	cfg.SetAWSRegion("") // Force invalid config

	_, err := f.CreateAWSProvider(context.Background(), cfg)
	assert.Error(t, err)
}

func TestCreateTerraformProvider_InvalidStateFile(t *testing.T) {
	logger := logging.New()
	f := factory.NewInstanceProviderFactory(logger)
	cfg := newMockConfig()
	cfg.SetUseHCL(false)
	cfg.SetStateFile("")

	_, err := f.CreateTerraformProvider(cfg)
	assert.Error(t, err)
}
