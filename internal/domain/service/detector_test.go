package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
)

// Mocks to validate interface implementation

type mockInstanceProvider struct{}

func (m *mockInstanceProvider) GetInstance(ctx context.Context, id string) (*model.Instance, error) {
	return nil, nil
}
func (m *mockInstanceProvider) ListInstances(ctx context.Context) ([]*model.Instance, error) {
	return nil, nil
}

type mockDriftDetector struct{}

func (m *mockDriftDetector) DetectDrift(ctx context.Context, source, target *model.Instance, attrs []string) (*model.DriftResult, error) {
	return nil, nil
}
func (m *mockDriftDetector) DetectDriftByID(ctx context.Context, id string, attrs []string) (*model.DriftResult, error) {
	return nil, nil
}
func (m *mockDriftDetector) DetectDriftForAll(ctx context.Context, attrs []string) ([]*model.DriftResult, error) {
	return nil, nil
}

type mockRepository struct{}

func (m *mockRepository) SaveDriftResult(ctx context.Context, r *model.DriftResult) error {
	return nil
}
func (m *mockRepository) GetDriftResult(ctx context.Context, id string) (*model.DriftResult, error) {
	return nil, nil
}
func (m *mockRepository) GetDriftResultsByInstanceID(ctx context.Context, id string) ([]*model.DriftResult, error) {
	return nil, nil
}
func (m *mockRepository) ListDriftResults(ctx context.Context) ([]*model.DriftResult, error) {
	return nil, nil
}

type mockReporter struct{}

func (m *mockReporter) ReportDrift(r *model.DriftResult) error {
	return nil
}
func (m *mockReporter) ReportMultipleDrifts(rs []*model.DriftResult) error {
	return nil
}

// Compile-time checks
var (
	_ service.InstanceProvider = (*mockInstanceProvider)(nil)
	_ service.DriftDetector    = (*mockDriftDetector)(nil)
	_ service.DriftRepository  = (*mockRepository)(nil)
	_ service.Reporter         = (*mockReporter)(nil)
	// _ service.DriftService        = (*mockDriftDetector)(nil)
	// _ service.DriftDetectorProvider = (*mockDriftDetector)(nil)
)

func TestDriftDetectorConfig_Fields(t *testing.T) {
	cfg := service.DriftDetectorConfig{
		SourceOfTruth:      model.OriginTerraform,
		AttributePaths:     []string{"instance_type", "tags.Name"},
		ParallelChecks:     3,
		Timeout:            45 * time.Second,
		ScheduleExpression: "0 0 * * *",
	}

	assert.Equal(t, model.OriginTerraform, cfg.SourceOfTruth)
	assert.Contains(t, cfg.AttributePaths, "tags.Name")
	assert.Equal(t, 3, cfg.ParallelChecks)
	assert.Equal(t, 45*time.Second, cfg.Timeout)
	assert.Equal(t, "0 0 * * *", cfg.ScheduleExpression)
}
