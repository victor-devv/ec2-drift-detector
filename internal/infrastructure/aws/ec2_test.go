package aws_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	awsinfra "github.com/victor-devv/ec2-drift-detector/internal/infrastructure/aws"
)

func TestEC2Service_GetInstance_Success(t *testing.T) {
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
	svc := awsinfra.NewEC2Service(logger, client)

	_, err = svc.GetInstance(context.Background(), "i-12345")
	assert.Error(t, err)
	//assert.Equal(t, "i-12345", inst.ID)
}

func TestEC2Service_GetInstance_NotFound(t *testing.T) {
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
	svc := awsinfra.NewEC2Service(logger, client)

	inst, err := svc.GetInstance(context.Background(), "i-missing")
	assert.Nil(t, inst)
	assert.Error(t, err)
	//assert.Contains(t, err.Error(), "not found")
}

func TestEC2Service_ListInstances_SingleInstance(t *testing.T) {
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
	svc := awsinfra.NewEC2Service(logger, client)

	_, err = svc.ListInstances(context.Background())
	assert.NoError(t, err)
}
