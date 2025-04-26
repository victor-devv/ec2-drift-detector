package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
)

// Client encapsulates AWS SDK client for EC2 operations
type Client struct {
	EC2Client *ec2.Client
	logger    *logging.Logger
	region    string
	endpoint  string
}

// ClientConfig holds AWS client configuration options
type ClientConfig struct {
	Region        string
	Profile       string
	AccessKey     string
	SecretKey     string
	Endpoint      string
	UseLocalstack bool
}

// NewClient creates a new AWS client
func NewClient(ctx context.Context, cfg ClientConfig, logger *logging.Logger) (*Client, error) {
	logger = logger.WithField("component", "aws-client")

	// Start with default AWS SDK configuration options
	var optFns []func(*config.LoadOptions) error

	// Apply AWS region if specified
	if cfg.Region != "" {
		optFns = append(optFns, config.WithRegion(cfg.Region))
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	// Apply AWS profile if specified
	if cfg.Profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(cfg.Profile))
	}

	// Load AWS SDK configuration
	awsConfig, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, errors.NewSystemError("Failed to load AWS configuration", err)
	}

	client := &Client{
		logger: logger,
		region: cfg.Region,
	}

	// Set custom endpoint for LocalStack if dev
	ec2Options := []func(*ec2.Options){}

	if cfg.UseLocalstack {
		if cfg.Endpoint == "" {
			cfg.Endpoint = "http://localhost:4566"
		}
		client.endpoint = cfg.Endpoint
		ec2Options = append(ec2Options, func(o *ec2.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.Region = cfg.Region
		})
		logger.Info(fmt.Sprintf("Using LocalStack endpoint: %s", cfg.Endpoint))
	} else if cfg.Endpoint != "" {
		client.endpoint = cfg.Endpoint
		ec2Options = append(ec2Options, func(o *ec2.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.Region = cfg.Region
		})
		logger.Info(fmt.Sprintf("Using custom endpoint: %s", cfg.Endpoint))
	}

	// Create EC2 client
	client.EC2Client = ec2.NewFromConfig(awsConfig, ec2Options...)

	// Test connection to AWS
	if err := client.testConnection(ctx); err != nil {
		return nil, err
	}

	logger.Info("AWS client initialized successfully")
	return client, nil
}

// testConnection tests the connection to AWS
func (c *Client) testConnection(ctx context.Context) error {
	c.logger.Debug("Testing connection to AWS EC2 service")

	_, err := c.EC2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return errors.NewSystemError(fmt.Sprintf("Failed to connect to AWS EC2 service: %v", err), err)
	}

	return nil
}

// GetRegion returns the AWS region
func (c *Client) GetRegion() string {
	return c.region
}

// GetEndpoint returns the AWS endpoint
func (c *Client) GetEndpoint() string {
	return c.endpoint
}
