/*
Wraps AWS SDK v2 to describe EC2 instances.

Uses filters if instance IDs are provided.

Converts AWS ec2.Instance data to internal EC2Instance model.

Includes support for test mocking via an injectable interface.
*/
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/sirupsen/logrus"
	cfg "github.com/victor-devv/ec2-drift-detector/internal/config"
)

// Client implements EC2Client interface
type Client struct {
	config aws.Config
	logger *logrus.Logger
}

// NewClient creates a new AWS EC2 client
func NewClient(ctx context.Context, cfg *cfg.Config, logger *logrus.Logger) (*Client, error) {
	var awsConfig aws.Config
	var err error

	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.AWS.DefaultRegion),
	}

	// Load AWS configuration
	awsConfig, err = awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	return &Client{
		config: awsConfig,
		logger: logger,
	}, nil
}

// expose accessors for testing
func (c *Client) Config() aws.Config {
	return c.config
}

func (c *Client) Logger() *logrus.Logger {
	return c.logger
}
