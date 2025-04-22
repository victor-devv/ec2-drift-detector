package utils

import (
	"fmt"
	"reflect"
	"sort"
)

// Comparator provides utility functions for comparing values
type Comparator struct{}

// NewComparator creates a new Comparator
func NewComparator() *Comparator {
	return &Comparator{}
}

// MapDifference represents a difference between two values
type MapDifference struct {
	FirstValue  string
	SecondValue string
}

// CompareMaps compares two maps and returns differences
func (c *Comparator) CompareMaps(first, second map[string]string) map[string]MapDifference {
	differences := make(map[string]MapDifference)

	// Check for keys in first but not in second, or with different values
	for k, v1 := range first {
		if v2, exists := second[k]; !exists {
			differences[k] = MapDifference{
				FirstValue:  v1,
				SecondValue: "(missing)",
			}
		} else if v1 != v2 {
			differences[k] = MapDifference{
				FirstValue:  v1,
				SecondValue: v2,
			}
		}
	}

	// Check for keys in second but not in first
	for k, v2 := range second {
		if _, exists := first[k]; !exists {
			differences[k] = MapDifference{
				FirstValue:  "(missing)",
				SecondValue: v2,
			}
		}
	}

	return differences
}

// CompareStringSlices compares two string slices and returns differences
func (c *Comparator) CompareStringSlices(first, second []string) []string {
	// Sort the slices for comparison
	sort.Strings(first)
	sort.Strings(second)

	if reflect.DeepEqual(first, second) {
		return nil
	}

	// Convert to maps for easier comparison
	firstMap := make(map[string]bool)
	for _, s := range first {
		firstMap[s] = true
	}

	secondMap := make(map[string]bool)
	for _, s := range second {
		secondMap[s] = true
	}

	// Find differences
	var differences []string

	// Items in first but not in second
	for s := range firstMap {
		if !secondMap[s] {
			differences = append(differences, fmt.Sprintf("- %s", s))
		}
	}

	// Items in second but not in first
	for s := range secondMap {
		if !firstMap[s] {
			differences = append(differences, fmt.Sprintf("+ %s", s))
		}
	}

	return differences
}
