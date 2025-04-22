package cli_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/victor-devv/ec2-drift-detector/internal/cli"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

func setupTestCLI() (*cli.CLI, *config.Config, *logrus.Logger) {
	cfg := &config.Config{
		Terraform: config.Config{}.Terraform,
		Detector:  config.Config{}.Detector,
	}
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel) // default

	return cli.NewCLI(cfg, logger), cfg, logger
}

func TestCLI_Parse_MinimumValidArgs(t *testing.T) {
	cliApp, cfg, _ := setupTestCLI()

	args := []string{
		"--state-file=terraform.tfstate",
	}
	err := cliApp.Parse(args)

	require.NoError(t, err)
	require.Equal(t, "terraform.tfstate", cfg.Terraform.StateFile)
}

func TestCLI_Parse_CustomAttributes(t *testing.T) {
	cliApp, cfg, _ := setupTestCLI()

	args := []string{
		"--state-file=terraform.tfstate",
		"--attributes=instance_type,tags,ami",
	}
	err := cliApp.Parse(args)

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"instance_type", "tags", "ami"}, cfg.Detector.Attributes)
}

func TestCLI_Parse_TrimsAttributeSpaces(t *testing.T) {
	cliApp, cfg, _ := setupTestCLI()

	args := []string{
		"--state-file=terraform.tfstate",
		"--attributes= instance_type , tags , ami ",
	}
	err := cliApp.Parse(args)

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"instance_type", "tags", "ami"}, cfg.Detector.Attributes)
}

func TestCLI_Parse_MissingRequiredStateFile(t *testing.T) {
	cliApp, _, _ := setupTestCLI()

	err := cliApp.Parse([]string{"--attributes=instance_type"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "--state-file is required")
}

func TestCLI_Parse_EnablesVerboseLogging(t *testing.T) {
	cliApp, _, logger := setupTestCLI()

	args := []string{
		"--state-file=terraform.tfstate",
		"--verbose",
	}
	err := cliApp.Parse(args)
	require.NoError(t, err)
	require.Equal(t, logrus.DebugLevel, logger.Level)
}

func TestCLI_LogStartupInfo(t *testing.T) {
	cliApp, cfg, logger := setupTestCLI()
	cfg.Terraform.StateFile = "main.tf"
	cfg.Detector.Attributes = []string{"ami", "tags"}
	cfg.Detector.OutputFormat = "json"

	logger.SetLevel(logrus.DebugLevel)
	err := cliApp.LogStartupInfo(context.Background())
	require.NoError(t, err)
}
