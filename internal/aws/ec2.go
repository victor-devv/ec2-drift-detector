package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/sirupsen/logrus"

	cfg "github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// EC2Client is an interface for AWS EC2 operations
type EC2Client interface {
	DescribeInstances(ctx context.Context, instanceIDs []string) ([]*models.EC2Instance, error)
}

type EC2ClientImpl struct {
	client *ec2.Client
	logger *logrus.Logger
}

// NewEC2Client creates a new EC2 client
func NewEC2Client(awsClient *Client, cfg *cfg.Config, logger *logrus.Logger) *EC2ClientImpl {
	ec2Client := ec2.NewFromConfig(awsClient.config, func(options *ec2.Options) {
		if cfg.Env != "production" {
			options.BaseEndpoint = aws.String(cfg.AWS.Ec2Endpoint)
			options.Region = cfg.AWS.DefaultRegion
		}
	})

	return &EC2ClientImpl{
		client: ec2Client,
		logger: logger,
	}
}
