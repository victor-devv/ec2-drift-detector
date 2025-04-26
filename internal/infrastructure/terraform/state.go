/*
Parses a standard Terraform .tfstate file or HCL configuration.

Extracts EC2 instance blocks into internal EC2Instance models.

Designed for extensibility to .tf HCL parsing.
*/
package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

// StateParser parses Terraform state files
type StateParser struct {
	logger *logging.Logger
}

// NewStateParser creates a new Terraform state parser
func NewStateParser(logger *logging.Logger) *StateParser {
	return &StateParser{
		logger: logger.WithField("component", "terraform-state"),
	}
}

// ParseStateFile parses a Terraform state file
func (p *StateParser) ParseStateFile(ctx context.Context, filePath string) (*model.TFState, error) {
	p.logger.Info(fmt.Sprintf("Parsing Terraform state file: %s", filePath))

	// Read the state file
	stateData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.NewOperationalError(fmt.Sprintf("Failed to read Terraform state file: %s", filePath), err)
	}

	// Parse the state file
	var state model.TFState
	if err := json.Unmarshal(stateData, &state); err != nil {
		return nil, errors.NewOperationalError("Failed to parse Terraform state JSON", err)
	}

	p.logger.Info(fmt.Sprintf("Successfully parsed Terraform state file with %d resources", len(state.Resources)))
	return &state, nil
}

// GetEC2InstancesFromState extracts EC2 instances from a Terraform state
func (p *StateParser) GetEC2InstancesFromState(state *model.TFState) ([]*model.Instance, error) {
	p.logger.Info("Extracting EC2 instances from Terraform state")

	var instances []*model.Instance

	// Find all aws_instance resources
	for _, resource := range state.Resources {
		if resource.Type == "aws_instance" {
			for _, instance := range resource.Instances {
				// Create a domain model instance from the Terraform instance
				domainInstance, err := p.mapToInstance(resource, instance)
				if err != nil {
					p.logger.Warn(fmt.Sprintf("Failed to map Terraform instance %s: %v", resource.Name, err))
					continue
				}

				instances = append(instances, domainInstance)
			}
		}
	}

	p.logger.Info(fmt.Sprintf("Found %d EC2 instances in Terraform state", len(instances)))
	return instances, nil
}

// GetEC2InstanceByID gets an EC2 instance by ID from a Terraform state
func (p *StateParser) GetEC2InstanceByID(state *model.TFState, instanceID string) (*model.Instance, error) {
	p.logger.Info(fmt.Sprintf("Looking for EC2 instance %s in Terraform state", instanceID))

	// Find the instance with the specified ID
	for _, resource := range state.Resources {
		if resource.Type == "aws_instance" {
			for _, instance := range resource.Instances {
				id, ok := instance.Attributes["id"].(string)
				if !ok {
					continue
				}

				if id == instanceID {
					// Create a domain model instance from the Terraform instance
					domainInstance, err := p.mapToInstance(resource, instance)
					if err != nil {
						return nil, errors.NewOperationalError(fmt.Sprintf("Failed to map Terraform instance %s", instanceID), err)
					}

					return domainInstance, nil
				}
			}
		}
	}

	return nil, errors.NewNotFoundError("EC2 Instance", instanceID)
}

// mapToInstance maps a Terraform instance to a domain model instance
func (p *StateParser) mapToInstance(resource model.TFResource, tfInstance model.TFResourceInstance) (*model.Instance, error) {
	// Extract instance ID
	id, ok := tfInstance.Attributes["id"].(string)
	if !ok {
		return nil, errors.NewOperationalError(fmt.Sprintf("Missing ID for Terraform instance %s", resource.Name), nil)
	}

	// Merge resource values and instance attributes
	attributes := make(map[string]interface{})

	// Copy resource values
	for k, v := range resource.Values {
		attributes[k] = v
	}

	// Copy instance attributes
	for k, v := range tfInstance.Attributes {
		attributes[k] = v
	}

	// Normalize attribute names (Terraform uses underscores, AWS might use camelCase)
	normalizedAttrs := p.normalizeAttributes(attributes)

	return model.NewInstance(id, normalizedAttrs, model.OriginTerraform), nil
}

// normalizeAttributes normalizes attribute names and values
func (p *StateParser) normalizeAttributes(attrs map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range attrs {
		// Snake case to camel case if needed
		key := k

		// Handle special cases
		switch k {
		case "vpc_security_group_ids":
			// Terraform stores this as a list or set
			if list, ok := v.([]interface{}); ok {
				strList := make([]string, 0, len(list))
				for _, item := range list {
					if str, ok := item.(string); ok {
						strList = append(strList, str)
					}
				}
				result[key] = strList
			} else {
				result[key] = v
			}
		case "tags":
			// Terraform stores tags as a map
			result[key] = v
		case "ebs_block_device":
			// Process EBS block devices
			if list, ok := v.([]interface{}); ok {
				result[key] = p.processEBSBlockDevices(list)
			} else {
				result[key] = v
			}
		default:
			// Direct copy for other attributes
			result[key] = v
		}
	}

	return result
}

// processEBSBlockDevices processes EBS block device configurations
func (p *StateParser) processEBSBlockDevices(devices []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(devices))

	for _, device := range devices {
		if deviceMap, ok := device.(map[string]interface{}); ok {
			blockDevice := make(map[string]interface{})

			for k, v := range deviceMap {
				blockDevice[k] = v
			}

			result = append(result, blockDevice)
		}
	}

	return result
}

// GetInstancesFromStateFile parses a Terraform state file and extracts EC2 instances
func (p *StateParser) GetInstancesFromStateFile(ctx context.Context, filePath string) ([]*model.Instance, error) {
	// Parse the state file
	state, err := p.ParseStateFile(ctx, filePath)
	if err != nil {
		return nil, err
	}

	// Extract EC2 instances
	return p.GetEC2InstancesFromState(state)
}

// GetInstanceByIDFromStateFile gets an EC2 instance by ID from a Terraform state file
func (p *StateParser) GetInstanceByIDFromStateFile(ctx context.Context, filePath, instanceID string) (*model.Instance, error) {
	// Parse the state file
	state, err := p.ParseStateFile(ctx, filePath)
	if err != nil {
		return nil, err
	}

	// Get the instance by ID
	return p.GetEC2InstanceByID(state, instanceID)
}
