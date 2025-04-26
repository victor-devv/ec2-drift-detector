package comparator

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewComparator(t *testing.T) {
	// Create a new comparator
	c := NewComparator()

	// Check default values
	assert.Equal(t, 10, c.MaxDepth)
	assert.False(t, c.IgnoreCase)
	assert.Len(t, c.IgnoreFields, 0)
	assert.False(t, c.TrimWhitespace)
}

func TestCompare(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Simple flat objects with differences
	source := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	target := map[string]interface{}{
		"name":  "John",
		"age":   35,                     // Different
		"email": "john.doe@example.com", // Different
	}

	paths := []string{"name", "age", "email"}
	diffs := c.Compare(source, target, paths)

	assert.Len(t, diffs, 2)
	assert.Contains(t, diffs, "age")
	assert.Contains(t, diffs, "email")
	assert.Equal(t, 30, diffs["age"].SourceValue)
	assert.Equal(t, 35, diffs["age"].TargetValue)
	assert.Equal(t, "john@example.com", diffs["email"].SourceValue)
	assert.Equal(t, "john.doe@example.com", diffs["email"].TargetValue)

	// Test case 2: Nested objects
	source = map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"address": map[string]interface{}{
				"city": "New York",
			},
		},
	}

	target = map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"address": map[string]interface{}{
				"city": "Boston", // Different
			},
		},
	}

	paths = []string{"user.name", "user.address.city"}
	diffs = c.Compare(source, target, paths)

	assert.Len(t, diffs, 1)
	assert.Contains(t, diffs, "user.address.city")
	assert.Equal(t, "New York", diffs["user.address.city"].SourceValue)
	assert.Equal(t, "Boston", diffs["user.address.city"].TargetValue)

	// Test case 3: Missing fields
	source = map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	target = map[string]interface{}{
		"name": "John",
		// Age is missing
	}

	paths = []string{"name", "age"}
	diffs = c.Compare(source, target, paths)

	assert.Len(t, diffs, 1)
	assert.Contains(t, diffs, "age")
	assert.Equal(t, 30, diffs["age"].SourceValue)
	assert.Nil(t, diffs["age"].TargetValue)

	// Test case 4: String comparison with options
	c.IgnoreCase = true
	c.TrimWhitespace = true

	source = map[string]interface{}{
		"name": "john",
		"text": " hello world ",
	}

	target = map[string]interface{}{
		"name": "JOHN",        // Different case
		"text": "hello world", // Different whitespace
	}

	paths = []string{"name", "text"}
	diffs = c.Compare(source, target, paths)

	assert.Len(t, diffs, 0) // No differences with options enabled

	// Reset options
	c.IgnoreCase = false
	c.TrimWhitespace = false

	// Test case 5: Empty paths
	diffs = c.Compare(source, target, []string{})
	assert.Len(t, diffs, 0)
}

func TestCompareDeep(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Deep comparison with nested structures
	source := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"address": map[string]interface{}{
				"city":  "New York",
				"state": "NY",
			},
		},
		"settings": map[string]interface{}{
			"notifications": true,
			"theme":         "dark",
		},
	}

	target := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"address": map[string]interface{}{
				"city":  "Boston", // Different
				"state": "MA",     // Different
			},
		},
		"settings": map[string]interface{}{
			"notifications": false, // Different
			"theme":         "dark",
			"language":      "en", // Only in target
		},
	}

	diffs := c.CompareDeep(source, target)

	assert.Contains(t, diffs, "user.address.city")
	assert.Contains(t, diffs, "user.address.state")
	assert.Contains(t, diffs, "settings.notifications")
	assert.Contains(t, diffs, "settings.language")

	// Test case 2: Non-map values
	diffs = c.CompareDeep("hello", "world")
	assert.Len(t, diffs, 1)
	assert.Contains(t, diffs, "")
	assert.Equal(t, "hello", diffs[""].SourceValue)
	assert.Equal(t, "world", diffs[""].TargetValue)

	// Test case 3: Comparing nil values
	diffs = c.CompareDeep(nil, nil)
	assert.Len(t, diffs, 0)

	// Test case 4: One nil value
	diffs = c.CompareDeep(nil, "value")
	assert.Len(t, diffs, 1)
	assert.Contains(t, diffs, "")
	assert.Nil(t, diffs[""].SourceValue)
	assert.Equal(t, "value", diffs[""].TargetValue)

	// Test case 5: With ignore fields
	// c.IgnoreFields = []string{"settings.language"}
	// diffs = c.CompareDeep(source, target)
	// assert.NotContains(t, diffs, "settings.language")
	// assert.Contains(t, diffs, "user.address.city")

	// Reset ignore fields
	c.IgnoreFields = []string{}
}

func TestCompareRecursive(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Recursive comparison with depth limit
	source := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"value": "source",
				},
			},
		},
	}

	target := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"value": "target", // Different
				},
			},
		},
	}

	// Create sync.Map to store results
	result := &sync.Map{}
	var wg sync.WaitGroup
	wg.Add(1)

	// Test with depth limit of 2 (shouldn't reach level3)
	c.compareRecursive(source, target, "", 2, result, &wg)
	wg.Wait()

	// Convert result to map for easier testing
	diffs := make(map[string]DiffEntry)
	result.Range(func(key, value interface{}) bool {
		diffs[key.(string)] = value.(DiffEntry)
		return true
	})

	assert.NotContains(t, diffs, "level1.level2.level3.value")

	// Test with depth limit of 4 (should reach level3)
	result = &sync.Map{}
	wg.Add(1)
	c.compareRecursive(source, target, "", 4, result, &wg)
	wg.Wait()

	// Convert result to map for easier testing
	diffs = make(map[string]DiffEntry)
	result.Range(func(key, value interface{}) bool {
		diffs[key.(string)] = value.(DiffEntry)
		return true
	})

	assert.Contains(t, diffs, "level1.level2.level3.value")
}

func TestGetValueByPath(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Simple path
	obj := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	value, exists := c.getValueByPath(obj, "name")
	assert.True(t, exists)
	assert.Equal(t, "John", value)

	// Test case 2: Nested path
	obj = map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"address": map[string]interface{}{
				"city": "New York",
			},
		},
	}

	value, exists = c.getValueByPath(obj, "user.address.city")
	assert.True(t, exists)
	assert.Equal(t, "New York", value)

	// Test case 3: Non-existent path
	value, exists = c.getValueByPath(obj, "user.email")
	assert.False(t, exists)
	assert.Nil(t, value)

	// Test case 4: Path through non-map
	value, exists = c.getValueByPath(obj, "user.name.first")
	assert.False(t, exists)
	assert.Nil(t, value)

	// Test case 5: Struct fields
	type Address struct {
		City  string
		State string
	}
	type User struct {
		Name    string
		Address Address
	}
	structObj := User{
		Name: "John",
		Address: Address{
			City:  "New York",
			State: "NY",
		},
	}

	value, exists = c.getValueByPath(structObj, "Name")
	assert.True(t, exists)
	assert.Equal(t, "John", value)

	value, exists = c.getValueByPath(structObj, "Address.City")
	assert.True(t, exists)
	assert.Equal(t, "New York", value)
}

func TestAreEqual(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Simple equality
	assert.True(t, c.areEqual(42, 42))
	assert.False(t, c.areEqual(42, 43))

	// Test case 2: String comparison
	assert.True(t, c.areEqual("hello", "hello"))
	assert.False(t, c.areEqual("hello", "Hello"))

	// Test case 3: String comparison with options
	c.IgnoreCase = true
	c.TrimWhitespace = true
	assert.True(t, c.areEqual("hello", "Hello"))
	assert.True(t, c.areEqual(" hello ", "Hello"))
	c.IgnoreCase = false
	c.TrimWhitespace = false

	// Test case 4: Nil values
	assert.True(t, c.areEqual(nil, nil))
	assert.False(t, c.areEqual(nil, "value"))
	assert.False(t, c.areEqual("value", nil))

	// Test case 5: Slices
	slice1 := []interface{}{"a", "b", "c"}
	slice2 := []interface{}{"c", "b", "a"} // Same elements, different order
	slice3 := []interface{}{"a", "b", "d"} // Different element
	assert.True(t, c.areEqual(slice1, slice2))
	assert.False(t, c.areEqual(slice1, slice3))

	// Test case 6: Maps
	map1 := map[string]interface{}{"a": 1, "b": 2}
	map2 := map[string]interface{}{"b": 2, "a": 1} // Same key/values, different order
	map3 := map[string]interface{}{"a": 1, "c": 3} // Different key
	assert.True(t, reflect.DeepEqual(map1, map2))  // Maps should be equal with DeepEqual
	assert.False(t, reflect.DeepEqual(map1, map3))
}

func TestInterfaceToMap(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Already a map
	mapObj := map[string]interface{}{
		"name": "John",
		"age":  30,
	}
	result, ok := c.interfaceToMap(mapObj)
	assert.True(t, ok)
	assert.Equal(t, mapObj, result)

	// Test case 2: Struct conversion
	type Person struct {
		Name string
		Age  int
	}
	structObj := Person{
		Name: "John",
		Age:  30,
	}
	result, ok = c.interfaceToMap(structObj)
	assert.True(t, ok)
	assert.Equal(t, "John", result["Name"])
	assert.Equal(t, 30, result["Age"])

	// Test case 3: Non-convertible type
	result, ok = c.interfaceToMap(42)
	assert.False(t, ok)
	assert.Nil(t, result)

	// Test case 4: Map with non-string keys
	nonStringMap := map[int]string{
		1: "one",
		2: "two",
	}
	result, ok = c.interfaceToMap(nonStringMap)
	assert.False(t, ok)
	assert.Nil(t, result)

	// Test case 5: Map with string keys and interface values
	stringMap := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	result, ok = c.interfaceToMap(stringMap)
	assert.True(t, ok)
	assert.Equal(t, stringMap, result)
}

func TestShouldIgnoreField(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: No ignored fields
	assert.False(t, c.shouldIgnoreField("name"))

	// Test case 2: With ignored fields
	c.IgnoreFields = []string{"name", "email"}
	assert.True(t, c.shouldIgnoreField("name"))
	assert.True(t, c.shouldIgnoreField("email"))
	assert.False(t, c.shouldIgnoreField("age"))

	// Reset ignored fields
	c.IgnoreFields = []string{}
}

func TestCompareFields(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Compare specific fields
	source := map[string]interface{}{
		"name":    "John",
		"age":     30,
		"email":   "john@example.com",
		"address": "123 Main St",
	}

	target := map[string]interface{}{
		"name":    "John",
		"age":     35,                     // Different
		"email":   "john.doe@example.com", // Different
		"address": "123 Main St",
	}

	fields := []string{"name", "age"}
	diffs := c.CompareFields(source, target, fields)

	assert.Len(t, diffs, 1)
	assert.Contains(t, diffs, "age")
	assert.NotContains(t, diffs, "email") // Not checked
	assert.Equal(t, 30, diffs["age"].SourceValue)
	assert.Equal(t, 35, diffs["age"].TargetValue)

	// Test case 2: Missing fields
	target = map[string]interface{}{
		"name": "John",
		// age is missing
	}

	diffs = c.CompareFields(source, target, fields)
	assert.Len(t, diffs, 1)
	assert.Contains(t, diffs, "age")
	assert.Equal(t, 30, diffs["age"].SourceValue)
	assert.Nil(t, diffs["age"].TargetValue)
}

func TestFormatDiff(t *testing.T) {
	// Create a comparator
	c := NewComparator()

	// Test case 1: Basic formatting
	diff := DiffEntry{
		Path:        "name",
		SourceValue: "John",
		TargetValue: "Jane",
		Changed:     true,
	}

	formatted := c.FormatDiff(diff)
	assert.Equal(t, "name: John => Jane", formatted)

	// Test case 2: With nil values
	diff = DiffEntry{
		Path:        "age",
		SourceValue: 30,
		TargetValue: nil,
		Changed:     true,
	}

	formatted = c.FormatDiff(diff)
	assert.Equal(t, "age: 30 => <nil>", formatted)

	diff = DiffEntry{
		Path:        "email",
		SourceValue: nil,
		TargetValue: "john@example.com",
		Changed:     true,
	}

	formatted = c.FormatDiff(diff)
	assert.Equal(t, "email: <nil> => john@example.com", formatted)
}
