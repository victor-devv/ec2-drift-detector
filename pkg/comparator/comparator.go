package comparator

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Comparator provides methods for comparing complex structures
type Comparator struct {
	// MaxDepth is the maximum depth for recursive comparisons
	MaxDepth int
	
	// IgnoreCase indicates whether string comparisons should be case-insensitive
	IgnoreCase bool
	
	// IgnoreFields is a list of field names to ignore during comparison
	IgnoreFields []string
	
	// TrimWhitespace indicates whether to trim whitespace in string comparisons
	TrimWhitespace bool
}

// DiffEntry represents a difference between two values
type DiffEntry struct {
	// Path is the dot-notation path to the differing attribute
	Path string
	
	// SourceValue is the value from the source object
	SourceValue interface{}
	
	// TargetValue is the value from the target object
	TargetValue interface{}
	
	// Changed indicates whether the values are different
	Changed bool
}

// NewComparator creates a new comparator with default settings
func NewComparator() *Comparator {
	return &Comparator{
		MaxDepth:       10,
		IgnoreCase:     false,
		IgnoreFields:   []string{},
		TrimWhitespace: false,
	}
}

// Compare compares two objects and returns a map of differences
func (c *Comparator) Compare(source, target interface{}, paths []string) map[string]DiffEntry {
	result := make(map[string]DiffEntry)
	var wg sync.WaitGroup
	resultMutex := sync.Mutex{}

	// Compare specified paths
	for _, path := range paths {
		wg.Add(1)
		go func(attrPath string) {
			defer wg.Done()

			sourceVal, sourceExists := c.getValueByPath(source, attrPath)
			targetVal, targetExists := c.getValueByPath(target, attrPath)

			// Check for existence in both sources
			if !sourceExists && !targetExists {
				return
			}

			if !sourceExists || !targetExists {
				resultMutex.Lock()
				result[attrPath] = DiffEntry{
					Path:        attrPath,
					SourceValue: sourceVal,
					TargetValue: targetVal,
					Changed:     true,
				}
				resultMutex.Unlock()
				return
			}

			// If both values exist, compare them
			if !c.areEqual(sourceVal, targetVal) {
				resultMutex.Lock()
				result[attrPath] = DiffEntry{
					Path:        attrPath,
					SourceValue: sourceVal,
					TargetValue: targetVal,
					Changed:     true,
				}
				resultMutex.Unlock()
			}
		}(path)
	}

	wg.Wait()
	return result
}

// CompareDeep performs a deep comparison of two objects
func (c *Comparator) CompareDeep(source, target interface{}) map[string]DiffEntry {
	result := make(map[string]DiffEntry)
	
	// Convert interfaces to maps for comparison
	sourceMap, sourceIsMap := c.interfaceToMap(source)
	targetMap, targetIsMap := c.interfaceToMap(target)
	
	if !sourceIsMap || !targetIsMap {
		// If either is not a map, compare directly
		if !c.areEqual(source, target) {
			result[""] = DiffEntry{
				Path:        "",
				SourceValue: source,
				TargetValue: target,
				Changed:     true,
			}
		}
		return result
	}
	
	// Do a deep comparison of the maps
	resultMap := sync.Map{}
	var wg sync.WaitGroup
	
	wg.Add(1)
	go c.compareRecursive(sourceMap, targetMap, "", c.MaxDepth, &resultMap, &wg)
	
	wg.Wait()
	
	// Convert resultMap to result
	resultMap.Range(func(key, value interface{}) bool {
		if path, ok := key.(string); ok {
			if entry, ok := value.(DiffEntry); ok {
				result[path] = entry
			}
		}
		return true
	})
	
	return result
}

// compareRecursive recursively compares two maps
func (c *Comparator) compareRecursive(source, target map[string]interface{}, basePath string, depth int, result *sync.Map, wg *sync.WaitGroup) {
	defer wg.Done()
	
	if depth <= 0 {
		return
	}
	
	// Compare keys in source
	for key, sourceVal := range source {
		// Skip ignored fields
		if c.shouldIgnoreField(key) {
			continue
		}
		
		path := key
		if basePath != "" {
			path = basePath + "." + key
		}
		
		targetVal, exists := target[key]
		if !exists {
			result.Store(path, DiffEntry{
				Path:        path,
				SourceValue: sourceVal,
				TargetValue: nil,
				Changed:     true,
			})
			continue
		}
		
		// Check if both values are maps
		sourceMapVal, sourceIsMap := c.interfaceToMap(sourceVal)
		targetMapVal, targetIsMap := c.interfaceToMap(targetVal)
		
		if sourceIsMap && targetIsMap {
			// Recursively compare maps
			wg.Add(1)
			go c.compareRecursive(sourceMapVal, targetMapVal, path, depth-1, result, wg)
		} else if !c.areEqual(sourceVal, targetVal) {
			// Compare non-map values
			result.Store(path, DiffEntry{
				Path:        path,
				SourceValue: sourceVal,
				TargetValue: targetVal,
				Changed:     true,
			})
		}
	}
	
	// Check for keys in target that aren't in source
	for key, targetVal := range target {
		// Skip ignored fields
		if c.shouldIgnoreField(key) {
			continue
		}
		
		path := key
		if basePath != "" {
			path = basePath + "." + key
		}
		
		if _, exists := source[key]; !exists {
			result.Store(path, DiffEntry{
				Path:        path,
				SourceValue: nil,
				TargetValue: targetVal,
				Changed:     true,
			})
		}
	}
}

// getValueByPath retrieves a value from an object by dot-notation path
func (c *Comparator) getValueByPath(obj interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	
	var current interface{} = obj
	
	for _, part := range parts {
		// Handle array indexing if needed with future implementation
		
		// Try to access as a map
		if m, ok := current.(map[string]interface{}); ok {
			current, ok = m[part]
			if !ok {
				return nil, false
			}
			continue
		}
		
		// Try to access as a struct
		v := reflect.ValueOf(current)
		if v.Kind() == reflect.Struct {
			field := v.FieldByName(part)
			if !field.IsValid() {
				return nil, false
			}
			current = field.Interface()
			continue
		}
		
		// If not a map or struct, can't navigate further
		return nil, false
	}
	
	return current, true
}

// areEqual compares two values for equality with options
func (c *Comparator) areEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	
	if a == nil || b == nil {
		return false
	}
	
	// Special handling for strings with options
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)
	
	if aIsStr && bIsStr {
		if c.TrimWhitespace {
			aStr = strings.TrimSpace(aStr)
			bStr = strings.TrimSpace(bStr)
		}
		
		if c.IgnoreCase {
			return strings.EqualFold(aStr, bStr)
		}
		
		return aStr == bStr
	}
	
	// Special handling for slices
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)
	
	if aVal.Kind() == reflect.Slice && bVal.Kind() == reflect.Slice {
		if aVal.Len() != bVal.Len() {
			return false
		}
		
		// Check if all elements are equal
		for i := 0; i < aVal.Len(); i++ {
			aElem := aVal.Index(i).Interface()
			
			// Find a matching element in b
			found := false
			for j := 0; j < bVal.Len(); j++ {
				bElem := bVal.Index(j).Interface()
				if c.areEqual(aElem, bElem) {
					found = true
					break
				}
			}
			
			if !found {
				return false
			}
		}
		
		return true
	}
	
	// Default to reflect.DeepEqual for other types
	return reflect.DeepEqual(a, b)
}

// interfaceToMap converts an interface to a map
func (c *Comparator) interfaceToMap(obj interface{}) (map[string]interface{}, bool) {
	// If it's already a map, return it
	if m, ok := obj.(map[string]interface{}); ok {
		return m, true
	}
	
	// Try to convert from a struct
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Struct {
		m := make(map[string]interface{})
		t := v.Type()
		
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath == "" { // Exported field
				m[field.Name] = v.Field(i).Interface()
			}
		}
		
		return m, true
	}
	
	// Try to convert from a map with string keys but interface{} values
	if v.Kind() == reflect.Map && v.Type().Key().Kind() == reflect.String {
		m := make(map[string]interface{})
		
		for _, key := range v.MapKeys() {
			m[key.String()] = v.MapIndex(key).Interface()
		}
		
		return m, true
	}
	
	return nil, false
}

// shouldIgnoreField checks if a field should be ignored
func (c *Comparator) shouldIgnoreField(field string) bool {
	for _, ignore := range c.IgnoreFields {
		if field == ignore {
			return true
		}
	}
	return false
}

// CompareFields compares specific fields between two objects
func (c *Comparator) CompareFields(source, target interface{}, fields []string) map[string]DiffEntry {
	result := make(map[string]DiffEntry)
	
	for _, field := range fields {
		sourceVal, sourceExists := c.getValueByPath(source, field)
		targetVal, targetExists := c.getValueByPath(target, field)
		
		// If either doesn't exist, mark as changed
		if !sourceExists || !targetExists {
			result[field] = DiffEntry{
				Path:        field,
				SourceValue: sourceVal,
				TargetValue: targetVal,
				Changed:     true,
			}
			continue
		}
		
		// Compare the values
		if !c.areEqual(sourceVal, targetVal) {
			result[field] = DiffEntry{
				Path:        field,
				SourceValue: sourceVal,
				TargetValue: targetVal,
				Changed:     true,
			}
		}
	}
	
	return result
}

// FormatDiff formats a diff entry as a string
func (c *Comparator) FormatDiff(entry DiffEntry) string {
	sourceStr := fmt.Sprintf("%v", entry.SourceValue)
	targetStr := fmt.Sprintf("%v", entry.TargetValue)
	
	if entry.SourceValue == nil {
		sourceStr = "<nil>"
	}
	
	if entry.TargetValue == nil {
		targetStr = "<nil>"
	}
	
	return fmt.Sprintf("%s: %s => %s", entry.Path, sourceStr, targetStr)
}
