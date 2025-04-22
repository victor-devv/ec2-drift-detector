package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/pkg/utils"
)

// Confirms no diffs on identical input
func TestCompareMaps_Identical(t *testing.T) {
	comp := utils.NewComparator()

	first := map[string]string{
		"Name": "test",
		"Env":  "dev",
	}

	second := map[string]string{
		"Name": "test",
		"Env":  "dev",
	}

	diff := comp.CompareMaps(first, second)
	require.Empty(t, diff, "expected no differences")
}

// Verifies detection of changed + missing fields
func TestCompareMaps_MissingAndChangedKeys(t *testing.T) {
	comp := utils.NewComparator()

	first := map[string]string{
		"Name": "prod-instance",
		"Env":  "prod",
		"Role": "web",
	}

	second := map[string]string{
		"Name": "prod-instance",
		"Env":  "staging",
		"Zone": "eu-north-1a",
	}

	diff := comp.CompareMaps(first, second)

	require.Len(t, diff, 3)

	require.Contains(t, diff, "Env")
	require.Equal(t, "prod", diff["Env"].FirstValue)
	require.Equal(t, "staging", diff["Env"].SecondValue)

	require.Contains(t, diff, "Zone")
	require.Equal(t, "(missing)", diff["Zone"].FirstValue)
	require.Equal(t, "eu-north-1a", diff["Zone"].SecondValue)

	require.Contains(t, diff, "Role")
	require.Equal(t, "web", diff["Role"].FirstValue)
	require.Equal(t, "(missing)", diff["Role"].SecondValue)
}

// Validates equality despite order
func TestCompareStringSlices_Identical(t *testing.T) {
	comp := utils.NewComparator()

	s1 := []string{"sg-123", "sg-456"}
	s2 := []string{"sg-456", "sg-123"} // same but unordered

	diff := comp.CompareStringSlices(s1, s2)
	require.Nil(t, diff, "expected identical slices to return nil")
}

// Confirms added elements detection
func TestCompareStringSlices_OneMissing(t *testing.T) {
	comp := utils.NewComparator()

	s1 := []string{"sg-123"}
	s2 := []string{"sg-123", "sg-789"}

	diff := comp.CompareStringSlices(s1, s2)

	require.ElementsMatch(t, []string{"+ sg-789"}, diff)
}

// Verifies two-way diff
func TestCompareStringSlices_SymmetricDifference(t *testing.T) {
	comp := utils.NewComparator()

	s1 := []string{"sg-111", "sg-222"}
	s2 := []string{"sg-333", "sg-222"}

	diff := comp.CompareStringSlices(s1, s2)

	require.ElementsMatch(t, []string{
		"- sg-111",
		"+ sg-333",
	}, diff)
}

// Confirms empty input is a match
func TestCompareMaps_Empty(t *testing.T) {
	comp := utils.NewComparator()

	first := map[string]string{}
	second := map[string]string{}

	diff := comp.CompareMaps(first, second)
	require.Empty(t, diff, "empty maps should produce no difference")
}

// Ensures nil is treated like empty
func TestCompareStringSlices_EmptyVsNil(t *testing.T) {
	comp := utils.NewComparator()

	var s1 []string = nil
	var s2 []string = nil

	diff := comp.CompareStringSlices(s1, s2)
	require.Nil(t, diff, "both nil should be equal")

	diff = comp.CompareStringSlices(s1, []string{})
	require.Nil(t, diff, "nil vs empty slice should be treated equally")
}
