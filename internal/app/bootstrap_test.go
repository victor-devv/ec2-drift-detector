package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/app"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/container"
)

func TestInitializeApplication_ReturnsApp(t *testing.T) {
	ctx := context.Background()

	c := container.NewContainer()

	cfg := &config.Config{}
	cfg.SetSourceOfTruth("aws")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetParallelChecks(2)
	cfg.SetTimeout(30 * time.Second)
	cfg.SetReporterType("console")
	cfg.SetAWSEndpoint("http://localhost:4566")
	cfg.SetAWSRegion("eu-north-1")
	cfg.SetAWSSecretAccessKey("dummy")
	cfg.SetAWSAccessKeyID("dummy")
	cfg.SetStateFile("./testdata/test.tfstate")

	appInstance, err := app.InitializeApplication(ctx, c, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, appInstance)
	assert.NotNil(t, appInstance.DriftDetector)
}

func TestInitializeApplication_InvalidConfig(t *testing.T) {
	ctx := context.Background()

	c := container.NewContainer()

	cfg := &config.Config{} // Missing required config
	// Do not set source of truth to force error

	_, err := app.InitializeApplication(ctx, c, cfg)
	assert.Error(t, err)
}
