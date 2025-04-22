package detector

import (
	"context"

	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// Detector interface defines the contract for all drift detectors
type Detector interface {
	// DetectDrift checks for differences between AWS resources and Terraform definitions
	DetectDrift(ctx context.Context, attributes []string) ([]models.DriftResult, error)

	// DetectDriftConcurrent checks for differences concurrently for multiple resources
	DetectDriftConcurrent(ctx context.Context, attributes []string) ([]models.DriftResult, error)
}

// BaseDetector provides common functionality for all detectors
type BaseDetector struct {
}
