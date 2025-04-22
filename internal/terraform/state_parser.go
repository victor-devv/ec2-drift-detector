package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// StateParser parses Terraform state files
type StateParser struct {
	BaseParser
	filePath string
}

// NewStateParser creates a new state parser
func NewStateParser(filePath string, log *logrus.Logger) *StateParser {
	return &StateParser{
		BaseParser: BaseParser{log: log},
		filePath:   filePath,
	}
}

// TFState represents the structure of a Terraform state file
type TFState struct {
	Version          int          `json:"version"`
	TerraformVersion string       `json:"terraform_version"`
	Serial           int          `json:"serial"`
	Lineage          string       `json:"lineage"`
	Resources        []TFResource `json:"resources"`
}

// TFResource represents a resource in a Terraform state file
type TFResource struct {
	Mode      string               `json:"mode"`
	Type      string               `json:"type"`
	Name      string               `json:"name"`
	Provider  string               `json:"provider"`
	Instances []TFResourceInstance `json:"instances"`
}

// TFResourceInstance represents an instance of a resource in a Terraform state file
type TFResourceInstance struct {
	IndexKey   interface{}            `json:"index_key"`
	Attributes map[string]interface{} `json:"attributes"`
	Private    string                 `json:"private"`
}

// GetEC2Instances returns all EC2 instances defined in the Terraform state file
func (p *StateParser) GetEC2Instances(ctx context.Context) ([]models.EC2Instance, error) {
	// Read and parse state file
	stateBytes, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Terraform state file: %w", err)
	}

	var state TFState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Terraform state file: %w", err)
	}

	// Find EC2 instances
	var instances []models.EC2Instance

	for _, resource := range state.Resources {
		if resource.Type == "aws_instance" {
			for _, instance := range resource.Instances {
				// Extract EC2 instance details
				ec2Instance, err := p.parseEC2Instance(instance)
				if err != nil {
					p.log.Warn(fmt.Sprintf("Failed to parse EC2 instance: %v", err))
					continue
				}
				instances = append(instances, ec2Instance)
			}
		}
	}

	return instances, nil
}

// parseEC2Instance extracts EC2 instance details from a Terraform resource instance
func (p *StateParser) parseEC2Instance(instance TFResourceInstance) (models.EC2Instance, error) {
	attrs := instance.Attributes

	// Extract instance ID
	instanceID, ok := attrs["id"].(string)
	if !ok || instanceID == "" {
		return models.EC2Instance{}, fmt.Errorf("instance id not found or invalid")
	}

	// Extract instance type
	instanceType, _ := attrs["instance_type"].(string)

	// Extract AMI
	ami, _ := attrs["ami"].(string)

	// Extract subnet ID
	subnetID, _ := attrs["subnet_id"].(string)

	// Extract security group IDs
	var securityGroupIDs []string
	if sgIDs, ok := attrs["vpc_security_group_ids"].([]interface{}); ok {
		for _, sgID := range sgIDs {
			if id, ok := sgID.(string); ok {
				securityGroupIDs = append(securityGroupIDs, id)
			}
		}
	}

	// Extract tags
	tags := make(map[string]string)
	for k, v := range attrs {
		if strings.HasPrefix(k, "tags.") && k != "tags.%" {
			tagName := strings.TrimPrefix(k, "tags.")
			if tagVal, ok := v.(string); ok {
				tags[tagName] = tagVal
			}
		}
	}

	// For state version 4+, tags might be stored in a different format
	if tagsMap, ok := attrs["tags"].(map[string]interface{}); ok {
		for k, v := range tagsMap {
			if tagVal, ok := v.(string); ok {
				tags[k] = tagVal
			}
		}
	}

	return models.EC2Instance{
		ID:               instanceID,
		InstanceType:     instanceType,
		AMI:              ami,
		SubnetID:         subnetID,
		SecurityGroupIDs: securityGroupIDs,
		Tags:             tags,
	}, nil
}
