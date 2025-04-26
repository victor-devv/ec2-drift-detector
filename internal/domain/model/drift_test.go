package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDriftResult(t *testing.T) {
	// Test creation of a new drift result
	result := NewDriftResult("i-12345", OriginAWS)

	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "i-12345", result.ResourceID)
	assert.Equal(t, OriginAWS, result.SourceType)
	assert.False(t, result.HasDrift)
	assert.NotNil(t, result.DriftedAttributes)
	assert.Empty(t, result.DriftedAttributes)
	assert.WithinDuration(t, time.Now(), result.Timestamp, 2*time.Second)
}

func TestAddDriftedAttribute(t *testing.T) {
	// Setup
	result := NewDriftResult("i-12345", OriginTerraform)

	// Test case 1: Add a drifted attribute
	result.AddDriftedAttribute("instance_type", "t2.micro", "t2.small")

	assert.True(t, result.HasDrift)
	assert.Equal(t, 1, len(result.DriftedAttributes))
	assert.Contains(t, result.DriftedAttributes, "instance_type")
	assert.Equal(t, "t2.micro", result.DriftedAttributes["instance_type"].SourceValue)
	assert.Equal(t, "t2.small", result.DriftedAttributes["instance_type"].TargetValue)
	assert.True(t, result.DriftedAttributes["instance_type"].Changed)

	// Test case 2: Add another drifted attribute
	result.AddDriftedAttribute("ami", "ami-12345", "ami-67890")

	assert.True(t, result.HasDrift)
	assert.Equal(t, 2, len(result.DriftedAttributes))
	assert.Contains(t, result.DriftedAttributes, "ami")
	assert.Equal(t, "ami-12345", result.DriftedAttributes["ami"].SourceValue)
	assert.Equal(t, "ami-67890", result.DriftedAttributes["ami"].TargetValue)
	assert.True(t, result.DriftedAttributes["ami"].Changed)
}

func TestSetDriftedAttributes(t *testing.T) {
	// Setup
	result := NewDriftResult("i-12345", OriginTerraform)
	drifts := map[string]AttributeDrift{
		"instance_type": {
			Path:        "instance_type",
			SourceValue: "t2.micro",
			TargetValue: "t2.small",
			Changed:     true,
		},
		"ami": {
			Path:        "ami",
			SourceValue: "ami-12345",
			TargetValue: "ami-67890",
			Changed:     true,
		},
	}

	// Test case 1: Set drifted attributes on an empty result
	result.SetDriftedAttributes(drifts)

	assert.True(t, result.HasDrift)
	assert.Equal(t, 2, len(result.DriftedAttributes))
	assert.Equal(t, drifts, result.DriftedAttributes)

	// Test case 2: Set empty drifted attributes
	result = NewDriftResult("i-12345", OriginTerraform)
	result.SetDriftedAttributes(map[string]AttributeDrift{})

	assert.False(t, result.HasDrift)
	assert.Empty(t, result.DriftedAttributes)
}

func TestGenerateUUID(t *testing.T) {
	// Test the UUID generation function
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)
	assert.NotEqual(t, uuid1, uuid2)
}
