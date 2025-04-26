package terraform

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

func TestStateParser_ParseStateFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "terraform-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample Terraform state file
	sampleState := model.TFState{
		Version:          4,
		TerraformVersion: "1.0.0",
		Resources: []model.TFResource{
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "test_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-12345",
							"instance_type": "t2.micro",
							"ami":           "ami-12345",
							"tags": map[string]interface{}{
								"Name": "test-instance",
								"Env":  "test",
							},
						},
					},
				},
			},
		},
		Outputs: map[string]interface{}{
			"instance_id": map[string]interface{}{
				"value": "i-12345",
			},
		},
	}

	// Write sample state to file
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	stateData, err := json.MarshalIndent(sampleState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	err = os.WriteFile(stateFile, stateData, 0644)
	if err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test parsing state file
	state, err := parser.ParseStateFile(context.Background(), stateFile)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, 4, state.Version)
	assert.Equal(t, "1.0.0", state.TerraformVersion)
	assert.Len(t, state.Resources, 1)
	assert.Equal(t, "aws_instance", state.Resources[0].Type)
	assert.Equal(t, "test_instance", state.Resources[0].Name)

	// Test parsing non-existent file
	_, err = parser.ParseStateFile(context.Background(), filepath.Join(tempDir, "non-existent.tfstate"))
	assert.Error(t, err)

	// Test parsing invalid JSON
	invalidFile := filepath.Join(tempDir, "invalid.tfstate")
	err = os.WriteFile(invalidFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	_, err = parser.ParseStateFile(context.Background(), invalidFile)
	assert.Error(t, err)
}

func TestStateParser_GetEC2InstancesFromState(t *testing.T) {
	// Create a sample Terraform state
	sampleState := &model.TFState{
		Version:          4,
		TerraformVersion: "1.0.0",
		Resources: []model.TFResource{
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "test_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-12345",
							"instance_type": "t2.micro",
							"ami":           "ami-12345",
							"tags": map[string]interface{}{
								"Name": "test-instance",
								"Env":  "test",
							},
							"vpc_security_group_ids": []interface{}{
								"sg-12345",
								"sg-67890",
							},
							"ebs_block_device": []interface{}{
								map[string]interface{}{
									"device_name": "/dev/sdf",
									"volume_size": float64(10),
									"volume_type": "gp2",
								},
							},
						},
					},
				},
			},
			{
				Mode:     "managed",
				Type:     "aws_s3_bucket", // Not an EC2 instance
				Name:     "test_bucket",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":   "test-bucket",
							"acl":  "private",
							"tags": map[string]interface{}{},
						},
					},
				},
			},
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "another_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-67890",
							"instance_type": "t2.small",
							"ami":           "ami-67890",
						},
					},
				},
			},
		},
	}

	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test extracting EC2 instances
	instances, err := parser.GetEC2InstancesFromState(sampleState)
	assert.NoError(t, err)
	assert.Len(t, instances, 2) // Should find two aws_instance resources

	// Verify instance properties
	assert.Equal(t, "i-12345", instances[0].ID)
	assert.Equal(t, "t2.micro", instances[0].InstanceType)
	assert.Equal(t, model.OriginTerraform, instances[0].Origin)
	assert.Equal(t, "ami-12345", instances[0].Attributes["ami"])

	// Check nested attributes
	tags, ok := instances[0].Attributes["tags"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test-instance", tags["Name"])
	assert.Equal(t, "test", tags["Env"])

	secGroups, ok := instances[0].Attributes["vpc_security_group_ids"].([]string)
	assert.True(t, ok)
	assert.Len(t, secGroups, 2)
	assert.Equal(t, "sg-12345", secGroups[0])
	assert.Equal(t, "sg-67890", secGroups[1])

	// Check second instance
	assert.Equal(t, "i-67890", instances[1].ID)
	assert.Equal(t, "t2.small", instances[1].InstanceType)
	assert.Equal(t, model.OriginTerraform, instances[1].Origin)
	assert.Equal(t, "ami-67890", instances[1].Attributes["ami"])

	// Test with no aws_instance resources
	emptyState := &model.TFState{
		Version:          4,
		TerraformVersion: "1.0.0",
		Resources: []model.TFResource{
			{
				Mode:     "managed",
				Type:     "aws_s3_bucket",
				Name:     "test_bucket",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":  "test-bucket",
							"acl": "private",
						},
					},
				},
			},
		},
	}

	instances, err = parser.GetEC2InstancesFromState(emptyState)
	assert.NoError(t, err)
	assert.Len(t, instances, 0) // Should find no aws_instance resources
}

func TestStateParser_GetEC2InstanceByID(t *testing.T) {
	// Create a sample Terraform state
	sampleState := &model.TFState{
		Version:          4,
		TerraformVersion: "1.0.0",
		Resources: []model.TFResource{
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "test_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-12345",
							"instance_type": "t2.micro",
							"ami":           "ami-12345",
						},
					},
				},
			},
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "another_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-67890",
							"instance_type": "t2.small",
							"ami":           "ami-67890",
						},
					},
				},
			},
		},
	}

	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test getting an instance by ID
	instance, err := parser.GetEC2InstanceByID(sampleState, "i-12345")
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "i-12345", instance.ID)
	assert.Equal(t, "t2.micro", instance.InstanceType)
	assert.Equal(t, model.OriginTerraform, instance.Origin)
	assert.Equal(t, "ami-12345", instance.Attributes["ami"])

	// Test getting an instance by a different ID
	instance, err = parser.GetEC2InstanceByID(sampleState, "i-67890")
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "i-67890", instance.ID)
	assert.Equal(t, "t2.small", instance.InstanceType)

	// Test getting a non-existent instance
	instance, err = parser.GetEC2InstanceByID(sampleState, "i-nonexistent")
	assert.Error(t, err)
	assert.Nil(t, instance)
}

func TestStateParser_MapToInstance(t *testing.T) {
	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test resource and instance for mapping
	resource := model.TFResource{
		Mode:     "managed",
		Type:     "aws_instance",
		Name:     "test_instance",
		Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
		Values: map[string]interface{}{
			"resource_name": "test_instance",
		},
	}

	tfInstance := model.TFResourceInstance{
		Attributes: map[string]interface{}{
			"id":            "i-12345",
			"instance_type": "t2.micro",
			"ami":           "ami-12345",
			"tags": map[string]interface{}{
				"Name": "test-instance",
				"Env":  "test",
			},
			"vpc_security_group_ids": []interface{}{
				"sg-12345",
				"sg-67890",
			},
			"ebs_block_device": []interface{}{
				map[string]interface{}{
					"device_name": "/dev/sdf",
					"volume_size": float64(10),
					"volume_type": "gp2",
				},
			},
		},
	}

	// Map to instance
	instance, err := parser.mapToInstance(resource, tfInstance)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "i-12345", instance.ID)
	assert.Equal(t, "t2.micro", instance.InstanceType)
	assert.Equal(t, model.OriginTerraform, instance.Origin)
	assert.Equal(t, "ami-12345", instance.Attributes["ami"])
	assert.Equal(t, "test_instance", instance.Attributes["resource_name"])

	// Test instance without ID
	tfInstanceNoID := model.TFResourceInstance{
		Attributes: map[string]interface{}{
			"instance_type": "t2.micro",
			"ami":           "ami-12345",
		},
	}

	_, err = parser.mapToInstance(resource, tfInstanceNoID)
	assert.Error(t, err)
}

func TestStateParser_NormalizeAttributes(t *testing.T) {
	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test attributes for normalization
	attrs := map[string]interface{}{
		"instance_type": "t2.micro",
		"ami":           "ami-12345",
		"vpc_security_group_ids": []interface{}{
			"sg-12345",
			"sg-67890",
		},
		"tags": map[string]interface{}{
			"Name": "test-instance",
			"Env":  "test",
		},
		"ebs_block_device": []interface{}{
			map[string]interface{}{
				"device_name": "/dev/sdf",
				"volume_size": float64(10),
				"volume_type": "gp2",
			},
		},
	}

	// Normalize attributes
	normalized := parser.normalizeAttributes(attrs)

	// Check basic attributes
	assert.Equal(t, "t2.micro", normalized["instance_type"])
	assert.Equal(t, "ami-12345", normalized["ami"])

	// Check security groups
	secGroups, ok := normalized["vpc_security_group_ids"].([]string)
	assert.True(t, ok)
	assert.Len(t, secGroups, 2)
	assert.Equal(t, "sg-12345", secGroups[0])
	assert.Equal(t, "sg-67890", secGroups[1])

	// Check tags
	assert.Equal(t, attrs["tags"], normalized["tags"])

	// Check EBS block devices
	ebsDevices, ok := normalized["ebs_block_device"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, ebsDevices, 1)
	assert.Equal(t, "/dev/sdf", ebsDevices[0]["device_name"])
	assert.Equal(t, float64(10), ebsDevices[0]["volume_size"])
	assert.Equal(t, "gp2", ebsDevices[0]["volume_type"])
}

func TestStateParser_ProcessEBSBlockDevices(t *testing.T) {
	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test EBS block devices
	devices := []interface{}{
		map[string]interface{}{
			"device_name": "/dev/sdf",
			"volume_size": float64(10),
			"volume_type": "gp2",
		},
		map[string]interface{}{
			"device_name": "/dev/sdg",
			"volume_size": float64(20),
			"volume_type": "io1",
			"iops":        float64(1000),
		},
	}

	// Process EBS block devices
	processed := parser.processEBSBlockDevices(devices)

	// Check processed devices
	assert.Len(t, processed, 2)
	assert.Equal(t, "/dev/sdf", processed[0]["device_name"])
	assert.Equal(t, float64(10), processed[0]["volume_size"])
	assert.Equal(t, "gp2", processed[0]["volume_type"])
	assert.Equal(t, "/dev/sdg", processed[1]["device_name"])
	assert.Equal(t, float64(20), processed[1]["volume_size"])
	assert.Equal(t, "io1", processed[1]["volume_type"])
	assert.Equal(t, float64(1000), processed[1]["iops"])

	// Test with non-map device
	devicesWithNonMap := []interface{}{
		map[string]interface{}{
			"device_name": "/dev/sdf",
		},
		"invalid", // Not a map
	}

	processed = parser.processEBSBlockDevices(devicesWithNonMap)
	assert.Len(t, processed, 1) // Only the valid map should be processed
	assert.Equal(t, "/dev/sdf", processed[0]["device_name"])
}

func TestStateParser_GetInstancesFromStateFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "terraform-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample Terraform state file
	sampleState := model.TFState{
		Version:          4,
		TerraformVersion: "1.0.0",
		Resources: []model.TFResource{
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "test_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-12345",
							"instance_type": "t2.micro",
							"ami":           "ami-12345",
						},
					},
				},
			},
		},
	}

	// Write sample state to file
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	stateData, err := json.MarshalIndent(sampleState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	err = os.WriteFile(stateFile, stateData, 0644)
	if err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test getting instances from state file
	instances, err := parser.GetInstancesFromStateFile(context.Background(), stateFile)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "i-12345", instances[0].ID)
	assert.Equal(t, "t2.micro", instances[0].InstanceType)
	assert.Equal(t, model.OriginTerraform, instances[0].Origin)
	assert.Equal(t, "ami-12345", instances[0].Attributes["ami"])

	// Test with invalid file path
	_, err = parser.GetInstancesFromStateFile(context.Background(), "non-existent.tfstate")
	assert.Error(t, err)
}

func TestStateParser_GetInstanceByIDFromStateFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "terraform-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample Terraform state file
	sampleState := model.TFState{
		Version:          4,
		TerraformVersion: "1.0.0",
		Resources: []model.TFResource{
			{
				Mode:     "managed",
				Type:     "aws_instance",
				Name:     "test_instance",
				Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
				Instances: []model.TFResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id":            "i-12345",
							"instance_type": "t2.micro",
							"ami":           "ami-12345",
						},
					},
				},
			},
		},
	}

	// Write sample state to file
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	stateData, err := json.MarshalIndent(sampleState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	err = os.WriteFile(stateFile, stateData, 0644)
	if err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	// Create a new state parser
	parser := NewStateParser(logging.New())

	// Test getting instance by ID from state file
	instance, err := parser.GetInstanceByIDFromStateFile(context.Background(), stateFile, "i-12345")
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "i-12345", instance.ID)
	assert.Equal(t, "t2.micro", instance.InstanceType)
	assert.Equal(t, model.OriginTerraform, instance.Origin)
	assert.Equal(t, "ami-12345", instance.Attributes["ami"])

	// Test with non-existent instance ID
	_, err = parser.GetInstanceByIDFromStateFile(context.Background(), stateFile, "i-nonexistent")
	assert.Error(t, err)

	// Test with invalid file path
	_, err = parser.GetInstanceByIDFromStateFile(context.Background(), "non-existent.tfstate", "i-12345")
	assert.Error(t, err)
}
