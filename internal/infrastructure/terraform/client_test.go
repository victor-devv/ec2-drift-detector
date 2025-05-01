package terraform_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/terraform"
)

func TestNewClient_StateFileSuccess(t *testing.T) {
	logger := logging.New()
	tempFile, err := os.CreateTemp(".", "test-*.tfstate")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	client, err := terraform.NewClient(terraform.ClientConfig{
		StateFile: tempFile.Name(),
		UseHCL:    false,
	}, logger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, tempFile.Name(), client.GetStateFile())
	assert.False(t, client.IsUsingHCL())
}

func TestNewClient_HCLDirSuccess(t *testing.T) {
	logger := logging.New()
	tempDir := t.TempDir()

	client, err := terraform.NewClient(terraform.ClientConfig{
		HCLDir: tempDir,
		UseHCL: true,
	}, logger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, tempDir, client.GetHCLDir())
	assert.True(t, client.IsUsingHCL())
}

func TestNewClient_MissingStateFile(t *testing.T) {
	logger := logging.New()

	_, err := terraform.NewClient(terraform.ClientConfig{
		StateFile: "nonexistent.tfstate",
		UseHCL:    false,
	}, logger)

	assert.Error(t, err)
}

func TestNewClient_MissingHCLDir(t *testing.T) {
	logger := logging.New()

	_, err := terraform.NewClient(terraform.ClientConfig{
		HCLDir: "nonexistent-dir",
		UseHCL: true,
	}, logger)

	assert.Error(t, err)
}

func TestGetSourceType(t *testing.T) {
	logger := logging.New()
	tempDir := t.TempDir()

	client, err := terraform.NewClient(terraform.ClientConfig{
		HCLDir: tempDir,
		UseHCL: true,
	}, logger)
	assert.NoError(t, err)
	assert.Equal(t, "terraform", string(client.GetSourceType()))
}

func TestListInstances_HCL(t *testing.T) {
	logger := logging.New()
	client, err := terraform.NewClient(terraform.ClientConfig{
		HCLDir: "testdata",
		UseHCL: true,
	}, logger)
	assert.NoError(t, err)

	instances, err := client.ListInstances(context.Background())
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "tf-aws_instance-web", instances[0].ID)
}

func TestListInstances_StateFile(t *testing.T) {
	logger := logging.New()
	client, err := terraform.NewClient(terraform.ClientConfig{
		StateFile: "./testdata/test.tfstate",
		UseHCL:    false,
	}, logger)
	assert.NoError(t, err)

	instances, err := client.ListInstances(context.Background())
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "i-1234567890abcdef0", instances[0].ID)
}

func TestGetInstance_StateFile(t *testing.T) {
	logger := logging.New()
	client, err := terraform.NewClient(terraform.ClientConfig{
		StateFile: "./testdata/test.tfstate",
		UseHCL:    false,
	}, logger)
	assert.NoError(t, err)

	instance, err := client.GetInstance(context.Background(), "i-1234567890abcdef0")
	assert.NoError(t, err)
	assert.Equal(t, "i-1234567890abcdef0", instance.ID)
}
