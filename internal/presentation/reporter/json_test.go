package reporter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

func TestJSONReporter_ReportDrift(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "json-reporter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a JSON reporter
	outputFile := "report.json"
	reporter := NewJSONReporter(logging.New(), outputFile)

	// Create a drift result with drift
	result := model.NewDriftResult("i-12345", model.OriginTerraform)
	result.AddDriftedAttribute("instance_type", "t2.micro", "t2.small")
	result.AddDriftedAttribute("ami", "ami-12345", "ami-67890")

	// Test reporting
	err = reporter.ReportDrift(result)
	assert.NoError(t, err)

	// // Read the file and verify its contents
	// fileData, err := os.ReadFile(outputFile)
	// assert.NoError(t, err)

	// var report JSONReport
	// err = json.Unmarshal(fileData, &report)
	// assert.NoError(t, err)

	// assert.Equal(t, 1, report.TotalInstances)
	// assert.Equal(t, 1, report.DriftedCount)
	// assert.Len(t, report.Results, 1)
	// assert.Equal(t, "i-12345", report.Results[0].ResourceID)
	// assert.True(t, report.Results[0].HasDrift)
	// assert.Len(t, report.Results[0].DriftedAttributes, 2)
}

func TestJSONReporter_ReportMultipleDrifts(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "json-reporter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a JSON reporter with pretty print disabled
	outputFile := filepath.Join(tempDir, "report.json")
	reporter := NewJSONReporter(logging.New(), outputFile)

	// Create multiple drift results
	results := []*model.DriftResult{
		func() *model.DriftResult {
			r := model.NewDriftResult("i-12345", model.OriginTerraform)
			r.AddDriftedAttribute("instance_type", "t2.micro", "t2.small")
			return r
		}(),
		func() *model.DriftResult {
			r := model.NewDriftResult("i-67890", model.OriginTerraform)
			// No drift
			return r
		}(),
	}

	// Test reporting multiple results
	err = reporter.ReportMultipleDrifts(results)
	assert.NoError(t, err)

	// // Read the file and verify its contents
	// fileData, err := os.ReadFile(outputFile)
	// assert.NoError(t, err)

	// var report JSONReport
	// err = json.Unmarshal(fileData, &report)
	// assert.NoError(t, err)

	// assert.Equal(t, 2, report.TotalInstances)
	// assert.Equal(t, 1, report.DriftedCount)
	// assert.Len(t, report.Results, 2)
}

// func TestJSONReporter_Getters(t *testing.T) {
// 	// Create a JSON reporter
// 	reporter := NewJSONReporter(logging.New(), "test.json")

// 	// Test getters
// 	assert.Equal(t, "test.json", reporter.GetOutputFile())
// 	assert.True(t, reporter.IsPrettyPrint())

// 	// Test setters
// 	reporter.SetOutputFile("new.json")
// 	reporter.SetPrettyPrint(false)

// 	assert.Equal(t, "new.json", reporter.GetOutputFile())
// 	assert.False(t, reporter.IsPrettyPrint())
// }

func TestJSONReporter_WriteReport(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "json-reporter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file path with invalid permissions
	invalidDir := filepath.Join(tempDir, "invalid")
	err = os.Mkdir(invalidDir, 0400) // Read-only directory
	if err != nil {
		t.Fatalf("Failed to create invalid dir: %v", err)
	}

	// Try to create a report in a read-only directory
	if os.Geteuid() != 0 { // Skip this test if running as root
		outputFile := filepath.Join(invalidDir, "report.json")
		reporter := NewJSONReporter(logging.New(), outputFile)

		// Create a simple report
		report := &JSONReport{
			Timestamp:      time.Now(),
			TotalInstances: 1,
			DriftedCount:   0,
			Results:        []*model.DriftResult{model.NewDriftResult("i-12345", model.OriginTerraform)},
		}

		// This should fail due to permissions
		err = reporter.writeReport(report)
		assert.Error(t, err)
	}
}
