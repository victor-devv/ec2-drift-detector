package service

import (
	"context"
	"time"

	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

// InstanceProvider defines the interface for retrieving instance configurations
type InstanceProvider interface {
	// GetInstance retrieves instance configuration by ID
	GetInstance(ctx context.Context, instanceID string) (*model.Instance, error)

	// ListInstances retrieves all available instances
	ListInstances(ctx context.Context) ([]*model.Instance, error)
}

// DriftDetector defines the interface for detecting drift between instances
type DriftDetector interface {
	// DetectDrift detects drift between two instances for specified attributes
	DetectDrift(ctx context.Context, source, target *model.Instance, attributePaths []string) (*model.DriftResult, error)

	// DetectDriftByID detects drift for an instance by ID
	DetectDriftByID(ctx context.Context, instanceID string, attributePaths []string) (*model.DriftResult, error)

	// DetectDriftForAll detects drift for all instances
	DetectDriftForAll(ctx context.Context, attributePaths []string) ([]*model.DriftResult, error)
}

// DriftRepository defines the interface for storing and retrieving drift results
type DriftRepository interface {
	// SaveDriftResult saves a drift detection result
	SaveDriftResult(ctx context.Context, result *model.DriftResult) error

	// GetDriftResult retrieves a drift detection result by ID
	GetDriftResult(ctx context.Context, id string) (*model.DriftResult, error)

	// GetDriftResultsByInstanceID retrieves drift detection results by instance ID
	GetDriftResultsByInstanceID(ctx context.Context, instanceID string) ([]*model.DriftResult, error)

	// ListDriftResults retrieves all drift detection results
	ListDriftResults(ctx context.Context) ([]*model.DriftResult, error)
}

// Reporter defines the interface for reporting drift detection results
type Reporter interface {
	// ReportDrift reports a single drift detection result
	ReportDrift(result *model.DriftResult) error

	// ReportMultipleDrifts reports multiple drift detection results
	ReportMultipleDrifts(results []*model.DriftResult) error
}

// DriftService defines the high-level interface for drift detection operations
type DriftService interface {
	// DetectAndReportDrift detects and reports drift for a single instance
	DetectAndReportDrift(ctx context.Context, instanceID string, attributePaths []string) error

	// DetectAndReportDriftForAll detects and reports drift for all instances
	DetectAndReportDriftForAll(ctx context.Context, attributePaths []string) error

	// RunScheduledDriftCheck runs a scheduled drift check
	RunScheduledDriftCheck(ctx context.Context) error
}

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
	SetReporters(reporters []Reporter)

	// Configuration getters
	GetAttributePaths() []string
	GetSourceOfTruth() model.ResourceOrigin
	GetParallelChecks() int
	GetTimeout() time.Duration
	GetScheduleExpression() string
}

// DriftDetectorConfig holds the configuration for drift detector services
type DriftDetectorConfig struct {
	SourceOfTruth      model.ResourceOrigin
	AttributePaths     []string
	ParallelChecks     int
	Timeout            time.Duration
	ScheduleExpression string
}
