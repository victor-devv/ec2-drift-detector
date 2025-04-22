package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
	"github.com/victor-devv/ec2-drift-detector/pkg/utils"
)

// JSONReporter reports drift detection results in JSON format
type JSONReporter struct {
	BaseReporter
	writer io.Writer
	log    *logrus.Logger
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(log *logrus.Logger) *JSONReporter {
	return &JSONReporter{
		writer: os.Stdout,
		log:    log,
	}
}

// WithWriter sets the writer for the JSON reporter
func (r *JSONReporter) WithWriter(writer io.Writer) *JSONReporter {
	r.writer = writer
	return r
}

func (r *JSONReporter) Writer() io.Writer {
	return r.writer
}

// enable storing output to file
func (r *JSONReporter) WithOutputFile(path string) *JSONReporter {
	path = utils.AppendUniqueSuffix(path)
	f, err := os.Create(path)
	if err != nil {
		r.log.Errorf("Failed to create output file: %v", err)
		r.writer = os.Stdout
	} else {
		r.writer = f
	}
	return r
}

// JSONReport represents the structure of a JSON drift report
type JSONReport struct {
	Summary struct {
		TotalResources      int `json:"totalResources"`
		DriftedResources    int `json:"driftedResources"`
		NonDriftedResources int `json:"nonDriftedResources"`
	} `json:"summary"`
	Results []models.DriftResult `json:"results"`
}

// Report generates a JSON report for the drift detection results
func (r *JSONReporter) Report(ctx context.Context, results []models.DriftResult) error {
	if len(results) == 0 {
		emptyReport := JSONReport{
			Summary: struct {
				TotalResources      int `json:"totalResources"`
				DriftedResources    int `json:"driftedResources"`
				NonDriftedResources int `json:"nonDriftedResources"`
			}{
				TotalResources:      0,
				DriftedResources:    0,
				NonDriftedResources: 0,
			},
			Results: []models.DriftResult{},
		}

		return json.NewEncoder(r.writer).Encode(emptyReport)
	}

	// Create report
	report := JSONReport{
		Results: results,
	}

	// Count drifted resources
	driftedCount := 0
	for _, res := range results {
		if res.Drifted {
			driftedCount++
		}
	}

	// Update summary
	report.Summary.TotalResources = len(results)
	report.Summary.DriftedResources = driftedCount
	report.Summary.NonDriftedResources = len(results) - driftedCount

	// Marshal to JSON
	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode JSON report: %w", err)
	}

	return nil
}
