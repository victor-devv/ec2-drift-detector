package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/pkg/utils"
)

// JSONReporter is an implementation of the Reporter interface that reports to JSON files
type JSONReporter struct {
	logger      *logging.Logger
	outputFile  string
	prettyPrint bool
}

// JSONReport represents the structure of a JSON report
type JSONReport struct {
	Timestamp      time.Time            `json:"timestamp"`
	TotalInstances int                  `json:"total_instances"`
	DriftedCount   int                  `json:"drifted_count"`
	Results        []*model.DriftResult `json:"results"`
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(logger *logging.Logger, outputFile string) *JSONReporter {
	return &JSONReporter{
		logger:      logger.WithField("component", "json-reporter"),
		outputFile:  utils.AppendUniqueSuffix(outputFile),
		prettyPrint: true,
	}
}

// ReportDrift reports a single drift detection result
func (r *JSONReporter) ReportDrift(result *model.DriftResult) error {
	r.logger.Info(fmt.Sprintf("Reporting drift for instance %s to JSON file", result.ResourceID))

	// Create a report with a single result
	report := &JSONReport{
		Timestamp:      time.Now(),
		TotalInstances: 1,
		DriftedCount:   boolToInt(result.HasDrift),
		Results:        []*model.DriftResult{result},
	}

	// Write the report to the output file
	return r.writeReport(report)
}

// ReportMultipleDrifts reports multiple drift detection results
func (r *JSONReporter) ReportMultipleDrifts(results []*model.DriftResult) error {
	r.logger.Info(fmt.Sprintf("Reporting drift for %d instances to JSON file", len(results)))

	// Count instances with drift
	var driftCount int
	for _, result := range results {
		if result.HasDrift {
			driftCount++
		}
	}

	// Create a report with multiple results
	report := &JSONReport{
		Timestamp:      time.Now(),
		TotalInstances: len(results),
		DriftedCount:   driftCount,
		Results:        results,
	}

	// Write the report to the output file
	return r.writeReport(report)
}

// writeReport writes a report to the output file
func (r *JSONReporter) writeReport(report *JSONReport) error {
	// Create the output directory if it doesn't exist
	dir := filepath.Dir(r.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewOperationalError(fmt.Sprintf("Failed to create output directory %s", dir), err)
	}

	// Encode the report to JSON
	var data []byte
	var err error
	if r.prettyPrint {
		data, err = json.MarshalIndent(report, "", "  ")
	} else {
		data, err = json.Marshal(report)
	}
	if err != nil {
		return errors.NewOperationalError("Failed to marshal report to JSON", err)
	}

	// Write the report to the output file
	if err := os.WriteFile(r.outputFile, data, 0644); err != nil {
		return errors.NewOperationalError(fmt.Sprintf("Failed to write report to %s", r.outputFile), err)
	}

	r.logger.Info(fmt.Sprintf("Successfully written report to %s", r.outputFile))
	return nil
}

// GetOutputFile returns the output file path
func (r *JSONReporter) GetOutputFile() string {
	return r.outputFile
}

// SetOutputFile sets the output file path
func (r *JSONReporter) SetOutputFile(outputFile string) {
	r.outputFile = outputFile
}

// IsPrettyPrint returns whether to use pretty printing
func (r *JSONReporter) IsPrettyPrint() bool {
	return r.prettyPrint
}

// SetPrettyPrint sets whether to use pretty printing
func (r *JSONReporter) SetPrettyPrint(prettyPrint bool) {
	r.prettyPrint = prettyPrint
}

// boolToInt converts a boolean to an integer (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
