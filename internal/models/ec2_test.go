package models_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

func TestEC2Instance_JSONMarshalling(t *testing.T) {
	instance := models.EC2Instance{
		ID:               "i-123456",
		InstanceType:     "t2.micro",
		AMI:              "ami-abc123",
		SubnetID:         "subnet-xyz",
		SecurityGroupIDs: []string{"sg-001", "sg-002"},
		Tags: map[string]string{
			"Name": "test-instance",
		},
		MonitoringEnabled: true,
	}

	data, err := json.Marshal(instance)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "\"id\":\"i-123456\"")

	var decoded models.EC2Instance
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "i-123456", decoded.ID)
	assert.Equal(t, "t2.micro", decoded.InstanceType)
	assert.True(t, decoded.MonitoringEnabled)
}

func TestDriftResult_JSONMarshalling(t *testing.T) {
	drift := models.DriftResult{
		ResourceID:   "i-abc",
		ResourceType: "aws_instance",
		InTerraform:  true,
		InAWS:        true,
		Drifted:      true,
		DriftDetails: []models.AttributeDiff{
			{
				Attribute:      "instance_type",
				AWSValue:       "t2.micro",
				TerraformValue: "t3.micro",
			},
			{
				Attribute:      "tags.Name",
				AWSValue:       "prod",
				TerraformValue: "dev",
				IsComplex:      false,
			},
		},
	}

	data, err := json.Marshal(drift)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "\"resourceId\":\"i-abc\"")

	var decoded models.DriftResult
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "aws_instance", decoded.ResourceType)
	assert.Equal(t, 2, len(decoded.DriftDetails))
	assert.Equal(t, "instance_type", decoded.DriftDetails[0].Attribute)
}

func TestAttributeDiff_FieldIntegrity(t *testing.T) {
	diff := models.AttributeDiff{
		Attribute:      "ebs_optimized",
		AWSValue:       true,
		TerraformValue: false,
		IsComplex:      false,
	}

	assert.Equal(t, "ebs_optimized", diff.Attribute)
	assert.Equal(t, true, diff.AWSValue)
	assert.Equal(t, false, diff.TerraformValue)
	assert.False(t, diff.IsComplex)
}
