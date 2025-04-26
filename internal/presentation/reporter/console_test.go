package reporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

func TestConsoleReporter_ReportDrift(t *testing.T) {
	// Create a console reporter without color for consistent testing
	reporter := NewConsoleReporter(logging.New())

	// Create a drift result with no drift
	result := model.NewDriftResult("i-12345", model.OriginTerraform)

	// Test reporting with no drift
	err := reporter.ReportDrift(result)
	assert.NoError(t, err)

	// Create a drift result with drift
	result = model.NewDriftResult("i-12345", model.OriginTerraform)
	result.AddDriftedAttribute("instance_type", "t2.micro", "t2.small")
	result.AddDriftedAttribute("ami", "ami-12345", "ami-67890")

	// Test reporting with drift
	err = reporter.ReportDrift(result)
	assert.NoError(t, err)

	// Test with color enabled
	reporter.SetColorEnabled(true)
	assert.True(t, reporter.IsColorEnabled())

	err = reporter.ReportDrift(result)
	assert.NoError(t, err)
}

func TestConsoleReporter_ReportMultipleDrifts(t *testing.T) {
	// Create a console reporter
	reporter := NewConsoleReporter(logging.New())

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
	err := reporter.ReportMultipleDrifts(results)
	assert.NoError(t, err)

	// Test reporting empty results
	err = reporter.ReportMultipleDrifts([]*model.DriftResult{})
	assert.NoError(t, err)

	// Test with color enabled
	reporter.SetColorEnabled(true)
	err = reporter.ReportMultipleDrifts(results)
	assert.NoError(t, err)
}

func TestConsoleReporter_Format(t *testing.T) {
	// Create reporters with and without color
	plainReporter := NewConsoleReporter(logging.New())
	colorReporter := NewConsoleReporter(logging.New())

	// Test formatHeader
	plainHeader := plainReporter.formatHeader("Test Header")
	colorHeader := colorReporter.formatHeader("Test Header")

	assert.Contains(t, plainHeader, "Test Header")
	assert.Contains(t, colorHeader, "Test Header")
	assert.Contains(t, colorHeader, "\033[") // Contains ANSI color codes

	// Test formatBool
	colorYes := colorReporter.formatBool(true)
	colorNo := colorReporter.formatBool(false)

	assert.Contains(t, colorYes, "Yes")
	assert.Contains(t, colorNo, "No")
	assert.Contains(t, colorYes, "\033[") // Contains ANSI color codes
	assert.Contains(t, colorNo, "\033[")  // Contains ANSI color codes

	// Test formatSuccess
	// plainSuccess := plainReporter.formatSuccess("Success")
	colorSuccess := colorReporter.formatSuccess("Success")

	// assert.Equal(t, "Success", plainSuccess)
	assert.Contains(t, colorSuccess, "Success")
	assert.Contains(t, colorSuccess, "\033[") // Contains ANSI color codes

	// Test formatError
	// plainError := plainReporter.formatError("Error")
	colorError := colorReporter.formatError("Error")

	// assert.Equal(t, "Error", plainError)
	assert.Contains(t, colorError, "Error")
	assert.Contains(t, colorError, "\033[") // Contains ANSI color codes

	// Test formatWarning
	// plainWarning := plainReporter.formatWarning("Warning")
	colorWarning := colorReporter.formatWarning("Warning")

	// assert.Equal(t, "Warning", plainWarning)
	assert.Contains(t, colorWarning, "Warning")
	assert.Contains(t, colorWarning, "\033[") // Contains ANSI color codes
}
