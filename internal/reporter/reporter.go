/*
Outputs drift results to:

Console (human-readable format)

JSON (for CI/CD, dashboards, or file storage)

Supports file output via --output-file
*/
package reporter

import (
	"context"

	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// Reporter interface defines the contract for all drift reporters
type Reporter interface {
	// Report generates a report for the drift detection results
	Report(ctx context.Context, results []models.DriftResult) error
}

// BaseReporter provides common functionality for all reporters
type BaseReporter struct {
}
