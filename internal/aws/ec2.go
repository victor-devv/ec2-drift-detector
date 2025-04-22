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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"

	cfg "github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// EC2Client is an interface for AWS EC2 operations
type EC2Client interface {
	DescribeInstances(ctx context.Context, instanceIDs []string) ([]*models.EC2Instance, error)
}

// for tests
type EC2API interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

type EC2ClientImpl struct {
	client EC2API
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

// DescribeInstances retrieves information about multiple EC2 instances
func (c *EC2ClientImpl) DescribeInstances(ctx context.Context, instanceIDs []string) ([]models.EC2Instance, error) {
	var instances []models.EC2Instance
	var nextToken *string

	// Convert string IDs to AWS string pointers
	var awsInstanceIDs []string
	if len(instanceIDs) > 0 {
		awsInstanceIDs = instanceIDs
	}

	// Define filter for instances if IDs are provided
	var filters []types.Filter
	if len(awsInstanceIDs) > 0 {
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-id"),
			Values: awsInstanceIDs,
		})
	}

	// Paginate through results
	for {
		resp, err := c.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: awsInstanceIDs,
			Filters:     filters,
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}

		// Process each reservation
		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				// Map AWS instance to our model
				ec2Instance := models.EC2Instance{
					ID:           *instance.InstanceId,
					InstanceType: string(instance.InstanceType),
					AMI:          *instance.ImageId,
				}

				// Extract subnet ID if available
				if instance.SubnetId != nil {
					ec2Instance.SubnetID = *instance.SubnetId
				}

				// Extract security groups
				for _, sg := range instance.SecurityGroups {
					ec2Instance.SecurityGroupIDs = append(ec2Instance.SecurityGroupIDs, *sg.GroupId)
				}

				// Extract tags
				ec2Instance.Tags = make(map[string]string)
				for _, tag := range instance.Tags {
					ec2Instance.Tags[*tag.Key] = *tag.Value
				}

				instances = append(instances, ec2Instance)
			}
		}

		// Check if there are more results
		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return instances, nil
}

func NewTestEC2Client(mockClient EC2API, logger *logrus.Logger) *EC2ClientImpl {
	return &EC2ClientImpl{
		client: mockClient,
		logger: logger,
	}
}
