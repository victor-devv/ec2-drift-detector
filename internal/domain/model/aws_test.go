package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewInstance(t *testing.T) {
	// Test case 1: Basic instance creation
	attrs := map[string]interface{}{
		"instance_type": "t2.micro",
		"ami":           "ami-12345",
		"tags": map[string]string{
			"Name": "test",
			"Env":  "dev",
		},
	}

	instance := NewInstance("i-12345", attrs, OriginAWS)

	require.NotNil(t, instance)
	require.Equal(t, "i-12345", instance.ID)
	require.Equal(t, "t2.micro", instance.InstanceType)
	require.Equal(t, OriginAWS, instance.Origin)
	require.Equal(t, attrs["ami"], instance.Attributes["ami"])
	require.Equal(t, attrs["tags"], instance.Attributes["tags"])

	// Test case 2: Instance creation without instance type
	attrs = map[string]interface{}{
		"ami": "ami-12345",
	}

	instance = NewInstance("i-12345", attrs, OriginTerraform)

	require.NotNil(t, instance)
	require.Equal(t, "i-12345", instance.ID)
	require.Empty(t, instance.InstanceType)
	require.Equal(t, OriginTerraform, instance.Origin)
	require.Equal(t, attrs["ami"], instance.Attributes["ami"])
}

func TestGetAttribute(t *testing.T) {
	// Setup test instance
	attrs := map[string]interface{}{
		"instance_type": "t2.micro",
		"ami":           "ami-12345",
		"tags": map[string]string{
			"Name": "test",
			"Env":  "dev",
		},
		"placement": map[string]interface{}{
			"availability_zone": "us-west-2a",
			"tenancy":           "default",
		},
		"root_block_device": map[string]interface{}{
			"volume_size": 8,
			"volume_type": "gp2",
		},
	}

	instance := NewInstance("i-12345", attrs, OriginAWS)

	// Test case 1: Get instance_type (direct attribute)
	val, exists := instance.GetAttribute("instance_type")
	require.True(t, exists)
	require.Equal(t, "t2.micro", val)

	// Test case 2: Get a nested attribute using dot notation
	val, exists = instance.GetAttribute("tags.Name")
	require.True(t, exists)
	require.Equal(t, "test", val)

	// Test case 3: Get a nested attribute two levels deep
	val, exists = instance.GetAttribute("placement.availability_zone")
	require.True(t, exists)
	require.Equal(t, "us-west-2a", val)

	// Test case 4: Get a non-existent attribute
	val, exists = instance.GetAttribute("non_existent")
	require.False(t, exists)
	require.Nil(t, val)

	// Test case 5: Get a non-existent nested attribute
	val, exists = instance.GetAttribute("tags.non_existent")
	require.False(t, exists)
	require.Nil(t, val)
}

func TestGetNestedValue(t *testing.T) {
	// Setup test data
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "value",
			},
		},
		"array": []interface{}{
			map[string]interface{}{
				"item": "value",
			},
		},
	}

	// Test case 1: Get a deeply nested value
	val, exists := GetNestedValue(data, "level1.level2.level3")
	require.True(t, exists)
	require.Equal(t, "value", val)

	// Test case 2: Get from non-existent path
	val, exists = GetNestedValue(data, "level1.non_existent.level3")
	require.False(t, exists)
	require.Nil(t, val)

	// Test case 3: Get from path where intermediate is not a map
	val, exists = GetNestedValue(data, "array.0.item")
	require.True(t, exists)
	require.Equal(t, "value", val)
}

func TestCompareAttributes(t *testing.T) {
	// Setup test instances
	sourceAttrs := map[string]interface{}{
		"instance_type": "t2.micro",
		"ami":           "ami-12345",
		"tags": map[string]string{
			"Name": "test",
			"Env":  "dev",
		},
	}

	targetAttrs := map[string]interface{}{
		"instance_type": "t2.small",  // Different
		"ami":           "ami-12345", // Same
		"tags": map[string]string{
			"Name": "test",
			"Env":  "prod", // Different
		},
	}

	source := NewInstance("i-12345", sourceAttrs, OriginTerraform)
	target := NewInstance("i-12345", targetAttrs, OriginAWS)

	// Test case 1: Compare attributes with differences
	attributePaths := []string{"instance_type", "ami", "tags"}
	drifts := CompareAttributes(source, target, attributePaths)

	require.Equal(t, 2, len(drifts))
	require.Contains(t, drifts, "instance_type")
	require.Contains(t, drifts, "tags")
	require.Equal(t, "t2.micro", drifts["instance_type"].SourceValue)
	require.Equal(t, "t2.small", drifts["instance_type"].TargetValue)

	// Test case 2: Compare attributes without differences
	attributePaths = []string{"ami"}
	drifts = CompareAttributes(source, target, attributePaths)

	require.Equal(t, 0, len(drifts))

	// Test case 3: Compare non-existent attributes
	attributePaths = []string{"non_existent"}
	drifts = CompareAttributes(source, target, attributePaths)

	require.Equal(t, 0, len(drifts))

	// Test case 4: One attribute exists in source but not in target
	delete(targetAttrs, "ami")
	target = NewInstance("i-12345", targetAttrs, OriginAWS)

	attributePaths = []string{"ami"}
	drifts = CompareAttributes(source, target, attributePaths)

	require.Equal(t, 1, len(drifts))
	require.Contains(t, drifts, "ami")

	// Test case 5: Empty attribute paths
	drifts = CompareAttributes(source, target, []string{})
	require.Equal(t, 0, len(drifts))
}

func TestNestedCompare(t *testing.T) {
	// Setup test data
	source := map[string]interface{}{
		"level1": map[string]interface{}{
			"a": "value1",
			"b": map[string]interface{}{
				"c": "value2",
			},
		},
		"level2": "unchanged",
	}

	target := map[string]interface{}{
		"level1": map[string]interface{}{
			"a": "value1_changed", // Different
			"b": map[string]interface{}{
				"c": "value2", // Same
			},
		},
		"level2": "unchanged", // Same
		"level3": "new_value", // Only in target
	}

	// Run nested comparison
	result := &sync.Map{}
	var wg sync.WaitGroup
	wg.Add(1)
	NestedCompare(source, target, "", 3, result, &wg)
	wg.Wait()

	// Convert result to a regular map for easier testing
	drifts := make(map[string]AttributeDrift)
	result.Range(func(key, value interface{}) bool {
		drifts[key.(string)] = value.(AttributeDrift)
		return true
	})

	// Test case 1: Should detect change in level1.a
	require.Contains(t, drifts, "level1.a")
	require.Equal(t, "value1", drifts["level1.a"].SourceValue)
	require.Equal(t, "value1_changed", drifts["level1.a"].TargetValue)

	// Test case 2: Should detect level3 only in target
	require.Contains(t, drifts, "level3")
	require.Nil(t, drifts["level3"].SourceValue)
	require.Equal(t, "new_value", drifts["level3"].TargetValue)

	// Test case 3: Should not detect drift in unchanged values
	require.NotContains(t, drifts, "level2")
	require.NotContains(t, drifts, "level1.b.c")

	// Test depth limitation
	result = &sync.Map{}
	wg.Add(1)
	NestedCompare(source, target, "", 1, result, &wg) // Only 1 level deep
	wg.Wait()

	drifts = make(map[string]AttributeDrift)
	result.Range(func(key, value interface{}) bool {
		drifts[key.(string)] = value.(AttributeDrift)
		return true
	})

	// Should detect level1 as different but not go deeper to level1.a
	require.NotContains(t, drifts, "level1.a")
	require.Contains(t, drifts, "level3")
}
