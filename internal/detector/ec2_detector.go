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
func NewEC2Detector(ec2Client aws.EC2ClientImpl, tfParser terraform.Parser, log *logrus.Logger) *EC2Detector {
	return &EC2Detector{
		ec2Client:  ec2Client,
		tfParser:   tfParser,
		log:        log,
		comparator: utils.NewComparator(),
	}
}

func existenceDrift(id string, inAWS, inTF bool) models.DriftResult {
	awsVal := "not_exists"
	tfVal := "not_exists"
	if inAWS {
		awsVal = "exists"
	}
	if inTF {
		tfVal = "exists"
	}
	return models.DriftResult{
		ResourceID:   id,
		ResourceType: "aws_instance",
		InTerraform:  inTF,
		InAWS:        inAWS,
		Drifted:      true,
		DriftDetails: []models.AttributeDiff{{
			Attribute:      "existence",
			AWSValue:       awsVal,
			TerraformValue: tfVal,
		}},
	}
}

// DetectDrift checks for differences between AWS EC2 instances and their Terraform definitions
func (d *EC2Detector) DetectDrift(ctx context.Context, attributes []string) ([]models.DriftResult, error) {
	instanceConfigs, err := d.tfParser.GetEC2Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance configurations from Terraform: %w", err)
	}
	if len(instanceConfigs) == 0 {
		return []models.DriftResult{}, nil
	}

	instanceIDToConfig := make(map[string]models.EC2Instance)
	instanceIDs := make([]string, 0, len(instanceConfigs))
	for _, config := range instanceConfigs {
		instanceIDs = append(instanceIDs, config.ID)
		instanceIDToConfig[config.ID] = config
	}

	awsInstances, err := d.ec2Client.DescribeInstances(ctx, instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to describe AWS instances: %w", err)
	}

	awsInstanceMap := make(map[string]models.EC2Instance)
	for _, inst := range awsInstances {
		awsInstanceMap[inst.ID] = inst
	}

	results := make([]models.DriftResult, 0)
	for _, awsInstance := range awsInstances {
		tfInstance, exists := instanceIDToConfig[awsInstance.ID]
		if !exists {
			results = append(results, existenceDrift(awsInstance.ID, true, false))
			continue
		}
		results = append(results, d.compareEC2Instance(awsInstance, tfInstance, attributes))
	}

	for id, tfInstance := range instanceIDToConfig {
		if _, found := awsInstanceMap[id]; !found {
			results = append(results, existenceDrift(tfInstance.ID, false, true))
		}
	}

	d.log.Infof("Drift detection complete: %d instance(s) compared", len(results))
	return results, nil
}

// DetectDriftConcurrent checks for differences concurrently for multiple EC2 instances
func (d *EC2Detector) DetectDriftConcurrent(ctx context.Context, attributes []string) ([]models.DriftResult, error) {
	instanceConfigs, err := d.tfParser.GetEC2Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance configurations from Terraform: %w", err)
	}
	if len(instanceConfigs) == 0 {
		return []models.DriftResult{}, nil
	}

	instanceIDToConfig := make(map[string]models.EC2Instance)
	instanceIDs := make([]string, 0, len(instanceConfigs))
	for _, config := range instanceConfigs {
		instanceIDs = append(instanceIDs, config.ID)
		instanceIDToConfig[config.ID] = config
	}

	awsInstances, err := d.ec2Client.DescribeInstances(ctx, instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to describe AWS instances: %w", err)
	}

	tasks := []comparisonTask{}
	awsInstanceMap := make(map[string]models.EC2Instance)
	for _, awsInstance := range awsInstances {
		awsInstanceMap[awsInstance.ID] = awsInstance
		if tfInstance, exists := instanceIDToConfig[awsInstance.ID]; exists {
			tasks = append(tasks, comparisonTask{
				awsInstance:  awsInstance,
				tfInstance:   tfInstance,
				attributeSet: attributes,
			})
		}
	}

	results, err := RunConcurrent(ctx, tasks, runtime.NumCPU(), func(ctx context.Context, task comparisonTask) (models.DriftResult, error) {
		return d.compareEC2Instance(task.awsInstance, task.tfInstance, task.attributeSet), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error during concurrent drift detection: %w", err)
	}

	for id := range awsInstanceMap {
		if _, found := instanceIDToConfig[id]; !found {
			results = append(results, existenceDrift(id, true, false))
		}
	}

	for id := range instanceIDToConfig {
		if _, found := awsInstanceMap[id]; !found {
			results = append(results, existenceDrift(id, false, true))
		}
	}

	d.log.Infof("Concurrent drift detection complete: %d instance(s) compared", len(results))
	return results, nil
}

type comparisonTask struct {
	awsInstance  models.EC2Instance
	tfInstance   models.EC2Instance
	attributeSet []string
}

func (d *EC2Detector) compareEC2Instance(awsInstance, tfInstance models.EC2Instance, attributes []string) models.DriftResult {
	result := models.DriftResult{
		ResourceID:   awsInstance.ID,
		ResourceType: "aws_instance",
		InTerraform:  true,
		InAWS:        true,
		Drifted:      false,
		DriftDetails: []models.AttributeDiff{},
	}

	compareAttrs := attributes
	if len(compareAttrs) == 0 {
		compareAttrs = []string{"instance_type", "ami", "subnet_id", "vpc_security_group_ids", "tags"}
	}

	for _, attr := range compareAttrs {
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
			d.log.Warnf("Unknown attribute for comparison: %s", attr)
		}
	}

	return result
}
