package reporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// ConsoleReporter reports drift detection results to the console
type ConsoleReporter struct {
	BaseReporter
	writer io.Writer
	log    *logrus.Logger
}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter(log *logrus.Logger) *ConsoleReporter {
	return &ConsoleReporter{
		writer: os.Stdout,
		log:    log,
	}
}

// WithWriter sets the writer for the console reporter
func (r *ConsoleReporter) WithWriter(writer io.Writer) *ConsoleReporter {
	r.writer = writer
	return r
}

// Report generates a console report for the drift detection results
func (r *ConsoleReporter) Report(ctx context.Context, results []models.DriftResult) error {
	if len(results) == 0 {
		fmt.Fprintln(r.writer, "No resources found for drift detection.")
		return nil
	}

	// Count drifted resources
	driftedCount := 0
	for _, res := range results {
		if res.Drifted {
			driftedCount++
		}
	}

	// Print summary
	if driftedCount == 0 {
		color.New(color.FgGreen).Fprintf(r.writer, "✓ No drift detected across %d resource(s)\n\n", len(results))
	} else {
		color.New(color.FgRed).Fprintf(r.writer, "✗ Drift detected in %d out of %d resource(s)\n\n", driftedCount, len(results))
	}

	// Create a new tabwriter
	w := tabwriter.NewWriter(r.writer, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintf(w, "RESOURCE ID\tTYPE\tSTATUS\tDETAILS\n")
	fmt.Fprintf(w, "----------\t----\t------\t-------\n")

	// Print results
	for _, res := range results {
		status := "OK"
		statusColor := color.FgGreen
		details := "-"

		if res.Drifted {
			status = "DRIFTED"
			statusColor = color.FgRed

			if !res.InTerraform {
				details = "Resource exists in AWS but not in Terraform"
			} else if !res.InAWS {
				details = "Resource exists in Terraform but not in AWS"
			} else if len(res.DriftDetails) > 0 {
				var driftDetails []string
				for _, drift := range res.DriftDetails {
					driftDetails = append(driftDetails, fmt.Sprintf("%s: AWS='%s', TF='%s'",
						drift.Attribute, drift.AWSValue, drift.TerraformValue))
				}
				details = strings.Join(driftDetails, "; ")
			}
		}

		fmt.Fprintf(w, "%s\t%s\t", res.ResourceID, res.ResourceType)
		color.New(statusColor).Fprintf(w, "%s\t", status)
		fmt.Fprintf(w, "%s\n", details)
	}

	return w.Flush()
}
