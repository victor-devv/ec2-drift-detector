package detector

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
	"github.com/victor-devv/ec2-drift-detector/internal/terraform"
	"github.com/victor-devv/ec2-drift-detector/pkg/utils"
)

// EC2Detector handles drift detection for EC2 instances
type EC2Detector struct {
	BaseDetector
	ec2Client  aws.EC2ClientImpl
	tfParser   terraform.Parser
	log        *logrus.Logger
	comparator *utils.Comparator
}

// NewEC2Detector creates a new EC2 detector
func NewEC2Detector(ec2Client *aws.EC2ClientImpl, tfParser terraform.Parser, log *logrus.Logger) *EC2Detector {
	return &EC2Detector{
		ec2Client:  *ec2Client,
		tfParser:   tfParser,
		log:        log,
		comparator: utils.NewComparator(),
	}
}

// DetectDrift checks for differences between AWS EC2 instances and their Terraform definitions
func (d *EC2Detector) DetectDrift(ctx context.Context, attributes []string) ([]models.DriftResult, error) {
	// Get instance IDs from Terraform
	instanceConfigs, err := d.tfParser.GetEC2Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance configurations from Terraform: %w", err)
	}

	if len(instanceConfigs) == 0 {
		return []models.DriftResult{}, nil
	}

	// Extract instance IDs
	var instanceIDs []string
	instanceIDToConfig := make(map[string]models.EC2Instance)
	for _, config := range instanceConfigs {
		instanceIDs = append(instanceIDs, config.ID)
		instanceIDToConfig[config.ID] = config
	}

	// Get actual instances from AWS
	awsInstances, err := d.ec2Client.DescribeInstances(ctx, instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to describe AWS instances: %w", err)
	}

	// Compare instances
	results := make([]models.DriftResult, 0, len(instanceIDs))
	for _, awsInstance := range awsInstances {
		tfInstance, exists := instanceIDToConfig[awsInstance.ID]
		if !exists {
			// Instance exists in AWS but not in Terraform
			results = append(results, models.DriftResult{
				ResourceID:   awsInstance.ID,
				ResourceType: "aws_instance",
				InTerraform:  false,
				InAWS:        true,
				Drifted:      true,
				DriftDetails: []models.AttributeDiff{
					{
						Attribute:      "existence",
						AWSValue:       "exists",
						TerraformValue: "not_exists",
					},
				},
			})
			continue
		}

		// Check for drift between AWS and Terraform
		driftResult := d.compareEC2Instance(awsInstance, tfInstance, attributes)
		results = append(results, driftResult)
	}

	// Check for instances in Terraform but not in AWS
	for id, tfInstance := range instanceIDToConfig {
		found := false
		for _, awsInstance := range awsInstances {
			if awsInstance.ID == id {
				found = true
				break
			}
		}

		if !found {
			results = append(results, models.DriftResult{
				ResourceID:   tfInstance.ID,
				ResourceType: "aws_instance",
				InTerraform:  true,
				InAWS:        false,
				Drifted:      true,
				DriftDetails: []models.AttributeDiff{
					{
						Attribute:      "existence",
						AWSValue:       "not_exists",
						TerraformValue: "exists",
					},
				},
			})
		}
	}

	return results, nil
}

// DetectDriftConcurrent checks for differences concurrently for multiple EC2 instances
func (d *EC2Detector) DetectDriftConcurrent(ctx context.Context, attributes []string) ([]models.DriftResult, error) {
	// Get instance IDs from Terraform
	instanceConfigs, err := d.tfParser.GetEC2Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance configurations from Terraform: %w", err)
	}

	if len(instanceConfigs) == 0 {
		return []models.DriftResult{}, nil
	}

	// Extract instance IDs
	var instanceIDs []string
	instanceIDToConfig := make(map[string]models.EC2Instance)
	for _, config := range instanceConfigs {
		instanceIDs = append(instanceIDs, config.ID)
		instanceIDToConfig[config.ID] = config
	}

	// Get actual instances from AWS
	awsInstances, err := d.ec2Client.DescribeInstances(ctx, instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to describe AWS instances: %w", err)
	}

	// Create a list of comparison tasks
	type comparisonTask struct {
		awsInstance  models.EC2Instance
		tfInstance   models.EC2Instance
		attributeSet []string
	}

	// Build comparison tasks
	var tasks []comparisonTask
	missingInAWS := make([]string, 0)
	missingInTF := make([]string, 0)

	// First, check instances in AWS
	for _, awsInstance := range awsInstances {
		tfInstance, exists := instanceIDToConfig[awsInstance.ID]
		if exists {
			tasks = append(tasks, comparisonTask{
				awsInstance:  awsInstance,
				tfInstance:   tfInstance,
				attributeSet: attributes,
			})
		} else {
			missingInTF = append(missingInTF, awsInstance.ID)
		}
	}

	// Check for instances in Terraform but not in AWS
	for id, tfInstance := range instanceIDToConfig {
		found := false
		for _, awsInstance := range awsInstances {
			if awsInstance.ID == id {
				found = true
				break
			}
		}
		if !found {
			missingInAWS = append(missingInAWS, tfInstance.ID)
		}
	}

	// Run comparisons concurrently
	concurrency := runtime.NumCPU()
	results, err := runConcurrent(ctx, tasks, concurrency, func(ctx context.Context, task comparisonTask) (models.DriftResult, error) {
		return d.compareEC2Instance(task.awsInstance, task.tfInstance, task.attributeSet), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error during concurrent drift detection: %w", err)
	}

	// Add missing instances
	for _, id := range missingInTF {
		results = append(results, models.DriftResult{
			ResourceID:   id,
			ResourceType: "aws_instance",
			InTerraform:  false,
			InAWS:        true,
			Drifted:      true,
			DriftDetails: []models.AttributeDiff{
				{
					Attribute:      "existence",
					AWSValue:       "exists",
					TerraformValue: "not_exists",
				},
			},
		})
	}

	for _, id := range missingInAWS {
		results = append(results, models.DriftResult{
			ResourceID:   id,
			ResourceType: "aws_instance",
			InTerraform:  true,
			InAWS:        false,
			Drifted:      true,
			DriftDetails: []models.AttributeDiff{
				{
					Attribute:      "existence",
					AWSValue:       "not_exists",
					TerraformValue: "exists",
				},
			},
		})
	}

	return results, nil
}

// compareEC2Instance compares AWS and Terraform EC2 instance configurations
func (d *EC2Detector) compareEC2Instance(awsInstance, tfInstance models.EC2Instance, attributes []string) models.DriftResult {
	result := models.DriftResult{
		ResourceID:   awsInstance.ID,
		ResourceType: "aws_instance",
		InTerraform:  true,
		InAWS:        true,
		Drifted:      false,
		DriftDetails: []models.AttributeDiff{},
	}

	// If attributes list is empty, check all attributes
	if len(attributes) == 0 {
		attributes = []string{
			"instance_type",
			"ami",
			"subnet_id",
			"vpc_security_group_ids",
			"tags",
		}
	}

	// Check each attribute
	for _, attr := range attributes {
		switch attr {
		case "instance_type":
			if awsInstance.InstanceType != tfInstance.InstanceType {
				result.Drifted = true
				result.DriftDetails = append(result.DriftDetails, models.AttributeDiff{
					Attribute:      "instance_type",
					AWSValue:       awsInstance.InstanceType,
					TerraformValue: tfInstance.InstanceType,
				})
			}
		case "ami":
			if awsInstance.AMI != tfInstance.AMI {
				result.Drifted = true
				result.DriftDetails = append(result.DriftDetails, models.AttributeDiff{
					Attribute:      "ami",
					AWSValue:       awsInstance.AMI,
					TerraformValue: tfInstance.AMI,
				})
			}
		case "subnet_id":
			if awsInstance.SubnetID != tfInstance.SubnetID {
				result.Drifted = true
				result.DriftDetails = append(result.DriftDetails, models.AttributeDiff{
					Attribute:      "subnet_id",
					AWSValue:       awsInstance.SubnetID,
					TerraformValue: tfInstance.SubnetID,
				})
			}
		case "vpc_security_group_ids":
			if !reflect.DeepEqual(awsInstance.SecurityGroupIDs, tfInstance.SecurityGroupIDs) {
				result.Drifted = true
				result.DriftDetails = append(result.DriftDetails, models.AttributeDiff{
					Attribute:      "vpc_security_group_ids",
					AWSValue:       fmt.Sprintf("%v", awsInstance.SecurityGroupIDs),
					TerraformValue: fmt.Sprintf("%v", tfInstance.SecurityGroupIDs),
				})
			}
		case "tags":
			tagDrifts := d.comparator.CompareMaps(awsInstance.Tags, tfInstance.Tags)
			if len(tagDrifts) > 0 {
				result.Drifted = true
				for k, drift := range tagDrifts {
					result.DriftDetails = append(result.DriftDetails, models.AttributeDiff{
						Attribute:      "tags." + k,
						AWSValue:       drift.FirstValue,
						TerraformValue: drift.SecondValue,
					})
				}
			}
		default:
			// Handle unknown attribute
			d.log.Warn(fmt.Sprintf("Unknown attribute for comparison: %s", attr))
		}
	}

	return result
}
