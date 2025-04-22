package aws_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
)

func TestNewClient_Success(t *testing.T) {
	ctx := context.Background()

	cfg := &config.Config{}
	cfg.Env = config.Env_Dev
	cfg.AWS.DefaultRegion = "eu-north-1"
	cfg.AWS.AccessKeyID = "dummy-access-key"
	cfg.AWS.SecretAccessKey = "dummy-secret"
	cfg.AWS.Ec2Endpoint = "http://localhost:4566"

	logger := logrus.New()

	client, err := aws.NewClient(ctx, cfg, logger)

	require.NoError(t, err, "expected no error creating aws client")
	require.NotNil(t, client, "expected non-nil client")

	require.Equal(t, "eu-north-1", client.Config().Region)
	require.Equal(t, logger, client.Logger())
}

func TestNew_WithEnv(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "eu-north-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	cfg, err := config.New()
	require.NoError(t, err)
	require.Equal(t, "eu-north-1", cfg.AWS.DefaultRegion)

	logger := logrus.New()

	client, err := aws.NewClient(context.Background(), cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, client)
}
