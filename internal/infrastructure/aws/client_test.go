package aws_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	awsinfra "github.com/victor-devv/ec2-drift-detector/internal/infrastructure/aws"
)

func TestNewClient_UsesLocalstack(t *testing.T) {
	logger := logging.New()
	ctx := context.Background()

	client, err := awsinfra.NewClient(ctx, awsinfra.ClientConfig{
		Region:        "us-west-2",
		AccessKey:     "test",
		SecretKey:     "secret",
		UseLocalstack: true,
		Endpoint:      "http://localhost:4566",
	}, logger)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:4566", client.GetEndpoint())
}

func TestNewClient_CustomEndpoint(t *testing.T) {
	logger := logging.New()
	ctx := context.Background()

	_, err := awsinfra.NewClient(ctx, awsinfra.ClientConfig{
		Region:    "us-east-1",
		AccessKey: "foo",
		SecretKey: "bar",
		Endpoint:  "http://custom-endpoint:1234",
	}, logger)

	require.Error(t, err) // should fail, custom endpoint doesn't exist
}

func TestNewClient_InvalidProfile(t *testing.T) {
	logger := logging.New()
	ctx := context.Background()

	_, err := awsinfra.NewClient(ctx, awsinfra.ClientConfig{
		Profile: "nonexistent-profile",
	}, logger)

	assert.Error(t, err)
}

func TestClient_GetRegionAndEndpoint(t *testing.T) {
	logger := logging.New()
	ctx := context.Background()

	client, err := awsinfra.NewClient(ctx, awsinfra.ClientConfig{
		Region:    "us-west-2",
		AccessKey: "test",
		SecretKey: "secret",
		Endpoint:  "http://localhost:4566",
	}, logger)

	require.NoError(t, err)
	assert.Equal(t, "us-west-2", client.GetRegion())
	assert.Equal(t, "http://localhost:4566", client.GetEndpoint())
}
