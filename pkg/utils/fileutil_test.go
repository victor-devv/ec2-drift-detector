package utils_test

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/pkg/utils"
)

func TestAppendUniqueSuffix_AppendsTimestampAndKeepsExtension(t *testing.T) {
	filename := "report.json"
	result := utils.AppendUniqueSuffix(filename)

	// Validate structure: report_YYYYMMDD_HHMMSS.json
	re := regexp.MustCompile(`^report_\d{8}_\d{6}\.json$`)
	require.True(t, re.MatchString(result), "filename did not match expected pattern")
}

func TestAppendUniqueSuffix_HandlesNoExtension(t *testing.T) {
	filename := "report"
	result := utils.AppendUniqueSuffix(filename)

	// Expected pattern: report_YYYYMMDD_HHMMSS
	re := regexp.MustCompile(`^report_\d{8}_\d{6}$`)
	require.True(t, re.MatchString(result), "filename without extension did not match expected pattern")
}

func TestAppendUniqueSuffix_PreservesCustomExtension(t *testing.T) {
	filename := "drift.report.txt"
	result := utils.AppendUniqueSuffix(filename)

	ext := filepath.Ext(result)
	require.Equal(t, ".txt", ext, "file extension not preserved")

	// Should contain original base name + suffix
	require.True(t, strings.HasPrefix(result, "drift.report_"), "prefix missing")
}
