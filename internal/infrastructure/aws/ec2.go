package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

// EC2Service handles AWS EC2 operations
type EC2Service struct {
	client *Client
	logger *logging.Logger
}

// NewEC2Service creates a new EC2 service
func NewEC2Service(logger *logging.Logger, client *Client) *EC2Service {
	return &EC2Service{
		client: client,
		logger: logger.WithField("component", "aws-ec2"),
	}
}

// GetInstance retrieves instance configuration by ID
func (s *EC2Service) GetInstance(ctx context.Context, instanceID string) (*model.Instance, error) {
	s.logger.Info(fmt.Sprintf("Retrieving EC2 instance: %s", instanceID))

	// Describe the EC2 instance
	resp, err := s.client.EC2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to retrieve EC2 instance %s", instanceID), err)
	}

	// Check if the instance was found
	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return nil, errors.NewNotFoundError("EC2 Instance", instanceID)
	}

	// Map the EC2 instance to our domain model
	instance := s.mapToInstance(resp.Reservations[0].Instances[0])
	return instance, nil
}

// ListInstances retrieves all available instances
func (s *EC2Service) ListInstances(ctx context.Context) ([]*model.Instance, error) {
	s.logger.Info("Listing all EC2 instances")

	var instances []*model.Instance
	var nextToken *string

	// Paginate through all instances
	for {
		resp, err := s.client.EC2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, errors.NewOperationalError("Failed to list EC2 instances", err)
		}

		// Process each reservation and instance
		for _, reservation := range resp.Reservations {
			for _, inst := range reservation.Instances {
				// Skip terminated instances
				if inst.State != nil && inst.State.Name == types.InstanceStateNameTerminated {
					continue
				}

				instances = append(instances, s.mapToInstance(inst))
			}
		}

		// Check if there are more instances
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}

	s.logger.Info(fmt.Sprintf("Found %d EC2 instances", len(instances)))
	return instances, nil
}

// ListInstancesParallel retrieves all available instances in parallel
func (s *EC2Service) ListInstancesParallel(ctx context.Context, maxConcurrency int) ([]*model.Instance, error) {
	s.logger.Info("Listing all EC2 instances in parallel")

	// First, get all instance IDs
	var instanceIDs []string
	var nextToken *string

	for {
		resp, err := s.client.EC2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, errors.NewOperationalError("Failed to list EC2 instance IDs", err)
		}

		// Extract instance IDs
		for _, reservation := range resp.Reservations {
			for _, inst := range reservation.Instances {
				// Skip terminated instances
				if inst.State != nil && inst.State.Name == types.InstanceStateNameTerminated {
					continue
				}

				if inst.InstanceId != nil {
					instanceIDs = append(instanceIDs, *inst.InstanceId)
				}
			}
		}

		// Check if there are more instances
		nextToken = resp.NextToken
		if nextToken == nil {
			break
		}
	}

	// Now fetch instance details in parallel
	instances := make([]*model.Instance, len(instanceIDs))
	errs := make([]error, len(instanceIDs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for i, id := range instanceIDs {
		wg.Add(1)
		go func(idx int, instanceID string) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Fetch the instance
			instance, err := s.GetInstance(ctx, instanceID)
			instances[idx] = instance
			errs[idx] = err
		}(i, id)
	}

	wg.Wait()

	// Check for errors
	var errsFound []error
	for i, err := range errs {
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error retrieving instance %s: %v", instanceIDs[i], err))
			errsFound = append(errsFound, err)
		}
	}

	// Filter out nil instances
	var validInstances []*model.Instance
	for _, inst := range instances {
		if inst != nil {
			validInstances = append(validInstances, inst)
		}
	}

	if len(errsFound) > 0 {
		return validInstances, errors.NewOperationalError(fmt.Sprintf("Failed to retrieve %d of %d instances", len(errsFound), len(instanceIDs)), nil)
	}

	s.logger.Info(fmt.Sprintf("Found %d EC2 instances", len(validInstances)))
	return validInstances, nil
}

// mapToInstance maps an EC2 instance to our domain model
func (s *EC2Service) mapToInstance(instance types.Instance) *model.Instance {
	attrs := make(map[string]interface{})

	// Only add non-nil values
	if instance.InstanceId != nil {
		attrs["id"] = *instance.InstanceId
	}

	if instance.InstanceType != "" {
		attrs["instance_type"] = string(instance.InstanceType)
	}

	if instance.ImageId != nil {
		attrs["ami"] = *instance.ImageId
	}

	if instance.Placement != nil {
		placement := make(map[string]interface{})
		if instance.Placement.AvailabilityZone != nil {
			placement["availability_zone"] = *instance.Placement.AvailabilityZone
		}
		if instance.Placement.Tenancy != "" {
			placement["tenancy"] = string(instance.Placement.Tenancy)
		}
		attrs["placement"] = placement
	}

	if len(instance.SecurityGroups) > 0 {
		sgIDs := make([]string, 0, len(instance.SecurityGroups))
		sgMap := make(map[string]string)

		for _, sg := range instance.SecurityGroups {
			if sg.GroupId != nil {
				sgIDs = append(sgIDs, *sg.GroupId)
			}
			if sg.GroupId != nil && sg.GroupName != nil {
				sgMap[*sg.GroupId] = *sg.GroupName
			}
		}

		attrs["vpc_security_group_ids"] = sgIDs
		attrs["security_groups"] = sgMap
	}

	if instance.SubnetId != nil {
		attrs["subnet_id"] = *instance.SubnetId
	}

	if instance.VpcId != nil {
		attrs["vpc_id"] = *instance.VpcId
	}

	if instance.PrivateIpAddress != nil {
		attrs["private_ip"] = *instance.PrivateIpAddress
	}

	if instance.PublicIpAddress != nil {
		attrs["public_ip"] = *instance.PublicIpAddress
	}

	if instance.KeyName != nil {
		attrs["key_name"] = *instance.KeyName
	}

	if instance.EbsOptimized != nil {
		attrs["ebs_optimized"] = *instance.EbsOptimized
	}

	if len(instance.BlockDeviceMappings) > 0 {
		blockDevices := make([]map[string]interface{}, 0, len(instance.BlockDeviceMappings))

		for _, blockDevice := range instance.BlockDeviceMappings {
			bd := make(map[string]interface{})

			if blockDevice.DeviceName != nil {
				bd["device_name"] = *blockDevice.DeviceName
			}

			if blockDevice.Ebs != nil {
				ebs := make(map[string]interface{})

				if blockDevice.Ebs.VolumeId != nil {
					ebs["volume_id"] = *blockDevice.Ebs.VolumeId
				}

				// if blockDevice.Ebs.VolumeSize != nil {
				// 	ebs["volume_size"] = *blockDevice.Ebs.VolumeSize
				// }

				// if blockDevice.Ebs.VolumeType != "" {
				// 	ebs["volume_type"] = string(blockDevice.Ebs.VolumeType)
				// }

				// if blockDevice.Ebs.DeleteOnTermination != nil {
				// 	ebs["delete_on_termination"] = *blockDevice.Ebs.DeleteOnTermination
				// }

				// if blockDevice.Ebs.Encrypted != nil {
				// 	ebs["encrypted"] = *blockDevice.Ebs.Encrypted
				// }

				bd["ebs"] = ebs
			}

			blockDevices = append(blockDevices, bd)
		}

		attrs["block_device_mappings"] = blockDevices
	}

	if len(instance.Tags) > 0 {
		tags := make(map[string]string)

		for _, tag := range instance.Tags {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}

		attrs["tags"] = tags
	}

	if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
		attrs["iam_instance_profile"] = *instance.IamInstanceProfile.Arn
	}

	if instance.State != nil {
		stateMap := make(map[string]interface{})

		if instance.State.Name != "" {
			stateMap["name"] = string(instance.State.Name)
		}

		if instance.State.Code != nil {
			stateMap["code"] = *instance.State.Code
		}

		attrs["state"] = stateMap
	}

	if instance.Monitoring != nil {
		attrs["monitoring"] = string(instance.Monitoring.State)
	}

	// Create the instance with the extracted attributes
	var instanceID string
	if instance.InstanceId != nil {
		instanceID = *instance.InstanceId
	}

	return model.NewInstance(instanceID, attrs, model.OriginAWS)
}
