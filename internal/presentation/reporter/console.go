package reporter

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

// ConsoleReporter is an implementation of the Reporter interface that reports to the console
type ConsoleReporter struct {
	logger  *logging.Logger
	colored bool
}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter(logger *logging.Logger) *ConsoleReporter {
	return &ConsoleReporter{
		logger:  logger.WithField("component", "console-reporter"),
		colored: true,
	}
}

// ReportDrift reports a single drift detection result
func (r *ConsoleReporter) ReportDrift(result *model.DriftResult) error {
	r.logger.Info(fmt.Sprintf("Reporting drift for instance %s", result.ResourceID))

	fmt.Println(r.formatHeader("Drift Detection Report"))
	fmt.Println()
	fmt.Printf("Instance ID: %s\n", result.ResourceID)
	fmt.Printf("Source Type: %s\n", result.SourceType)
	fmt.Printf("Timestamp: %s\n", result.Timestamp.Format(time.RFC3339))
	fmt.Printf("Has Drift: %s\n", r.formatBool(result.HasDrift))
	fmt.Println()

	if !result.HasDrift {
		fmt.Println(r.formatSuccess("No drift detected."))
		return nil
	}

	fmt.Println(r.formatHeader("Drifted Attributes"))
	fmt.Println()

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Attribute\tSource Value\tTarget Value")
	fmt.Fprintln(w, "---------\t------------\t------------")

	for path, drift := range result.DriftedAttributes {
		fmt.Fprintf(w, "%s\t%v\t%v\n", path, drift.SourceValue, drift.TargetValue)
	}
	w.Flush()
	fmt.Println()

	return nil
}

// ReportMultipleDrifts reports multiple drift detection results
func (r *ConsoleReporter) ReportMultipleDrifts(results []*model.DriftResult) error {
	r.logger.Info(fmt.Sprintf("Reporting drift for %d instances", len(results)))

	fmt.Println(r.formatHeader("Drift Detection Summary"))
	fmt.Println()
	fmt.Printf("Number of Instances: %d\n", len(results))

	// Count instances with drift
	var driftCount int
	for _, result := range results {
		if result.HasDrift {
			driftCount++
		}
	}

	fmt.Printf("Instances with Drift: %s (%d/%d)\n", r.formatBool(driftCount > 0), driftCount, len(results))
	fmt.Println()

	if driftCount == 0 {
		fmt.Println(r.formatSuccess("No drift detected in any instance."))
		return nil
	}

	fmt.Println(r.formatHeader("Instances with Drift"))
	fmt.Println()

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Instance ID\tDrifted Attributes\tTimestamp")
	fmt.Fprintln(w, "-----------\t------------------\t---------")

	for _, result := range results {
		if result.HasDrift {
			attrs := make([]string, 0, len(result.DriftedAttributes))
			for path := range result.DriftedAttributes {
				attrs = append(attrs, path)
			}
			attrsStr := strings.Join(attrs, ", ")
			fmt.Fprintf(w, "%s\t%s\t%s\n", result.ResourceID, attrsStr, result.Timestamp.Format(time.RFC3339))
		}
	}
	w.Flush()
	fmt.Println()

	// Prompt to show details
	fmt.Println("Use 'drift-detector show <instance-id>' to see detailed drift information for a specific instance.")
	fmt.Println()

	return nil
}

// formatHeader formats a header string
func (r *ConsoleReporter) formatHeader(text string) string {
	if r.colored {
		return fmt.Sprintf("\033[1;36m=== %s ===\033[0m", text)
	}
	return fmt.Sprintf("=== %s ===", text)
}

// formatBool formats a boolean value
func (r *ConsoleReporter) formatBool(value bool) string {
	if !r.colored {
		if value {
			return "Yes"
		}
		return "No"
	}

	if value {
		return "\033[1;31mYes\033[0m" // Red for drift
	}
	return "\033[1;32mNo\033[0m" // Green for no drift
}

// formatSuccess formats a success message
func (r *ConsoleReporter) formatSuccess(text string) string {
	if r.colored {
		return fmt.Sprintf("\033[1;32m%s\033[0m", text)
	}
	return text
}

// formatError formats an error message
func (r *ConsoleReporter) formatError(text string) string {
	if r.colored {
		return fmt.Sprintf("\033[1;31m%s\033[0m", text)
	}
	return text
}

// formatWarning formats a warning message
func (r *ConsoleReporter) formatWarning(text string) string {
	if r.colored {
		return fmt.Sprintf("\033[1;33m%s\033[0m", text)
	}
	return text
}

// IsColorEnabled returns whether color is enabled
func (r *ConsoleReporter) IsColorEnabled() bool {
	return r.colored
}

// SetColorEnabled sets whether color is enabled
func (r *ConsoleReporter) SetColorEnabled(enabled bool) {
	r.colored = enabled
}
