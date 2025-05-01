package app

import (
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
)

// Application represents the application and its services
type Application struct {
	DriftDetector service.DriftDetectorProvider
}

// NewApplication creates a new application with the given services
func NewApplication(driftDetector service.DriftDetectorProvider) *Application {
	return &Application{
		DriftDetector: driftDetector,
	}
}
