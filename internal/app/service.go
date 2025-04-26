package app

import (
	"context"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
)

// DriftDetectorProvider defines the interface for a drift detector service
type DriftDetectorProvider interface {
	// DetectDrift detects drift between two instances for specified attributes
	DetectDrift(ctx context.Context, source, target *model.Instance, attributePaths []string) (*model.DriftResult, error)

	// DetectDriftByID detects drift for an instance by ID
	DetectDriftByID(ctx context.Context, instanceID string, attributePaths []string) (*model.DriftResult, error)

	// DetectDriftForAll detects drift for all instances
	DetectDriftForAll(ctx context.Context, attributePaths []string) ([]*model.DriftResult, error)

	// DetectAndReportDrift detects and reports drift for a single instance
	DetectAndReportDrift(ctx context.Context, instanceID string, attributePaths []string) error

	// DetectAndReportDriftForAll detects and reports drift for all instances
	DetectAndReportDriftForAll(ctx context.Context, attributePaths []string) error

	// RunScheduledDriftCheck runs a scheduled drift check
	RunScheduledDriftCheck(ctx context.Context) error

	// StartScheduler starts the scheduler
	StartScheduler(ctx context.Context) error

	// StopScheduler stops the scheduler
	StopScheduler()

	// Configuration setters
	SetSourceOfTruth(sourceOfTruth model.ResourceOrigin)
	SetAttributePaths(attributePaths []string)
	SetParallelChecks(parallelChecks int)
	SetTimeout(timeout time.Duration)
	SetScheduleExpression(expression string)
	SetReporters(reporters []service.Reporter)

	// Configuration getters
	GetAttributePaths() []string
	GetSourceOfTruth() model.ResourceOrigin
	GetParallelChecks() int
	GetTimeout() time.Duration
	GetScheduleExpression() string
}

// Application represents the application and its services
type Application struct {
	DriftDetector DriftDetectorProvider
}

// NewApplication creates a new application with the given services
func NewApplication(driftDetector DriftDetectorProvider) *Application {
	return &Application{
		DriftDetector: driftDetector,
	}
}
