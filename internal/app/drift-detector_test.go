package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/app"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
)

type mockInstanceProvider struct {
	instances []*model.Instance
	err       error
}

func (m *mockInstanceProvider) GetInstance(ctx context.Context, id string) (*model.Instance, error) {
	if len(m.instances) > 0 {
		return m.instances[0], nil
	}
	return nil, m.err
}

func (m *mockInstanceProvider) ListInstances(ctx context.Context) ([]*model.Instance, error) {
	return m.instances, m.err
}

type mockRepository struct {
	saved []*model.DriftResult
}

func (m *mockRepository) SaveDriftResult(ctx context.Context, result *model.DriftResult) error {
	m.saved = append(m.saved, result)
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

type mockReporter struct {
	reported []*model.DriftResult
}

func (m *mockReporter) ReportDrift(result *model.DriftResult) error {
	m.reported = append(m.reported, result)
	return nil
}
func (m *mockReporter) ReportMultipleDrifts(results []*model.DriftResult) error {
	m.reported = append(m.reported, results...)
	return nil
}

func TestDetectAndReportDrift(t *testing.T) {
	awsInst := model.NewInstance("i-123", map[string]interface{}{"instance_type": "t2.micro"}, model.OriginAWS)
	tfInst := model.NewInstance("i-123", map[string]interface{}{"instance_type": "t2.small"}, model.OriginTerraform)
	repo := &mockRepository{}
	reporter := &mockReporter{}

	detector := app.NewDriftDetectorService(
		&mockInstanceProvider{instances: []*model.Instance{awsInst}},
		&mockInstanceProvider{instances: []*model.Instance{tfInst}},
		repo,
		[]service.Reporter{reporter},
		service.DriftDetectorConfig{
			SourceOfTruth:  model.OriginAWS,
			AttributePaths: []string{"instance_type"},
			Timeout:        2 * time.Second,
			ParallelChecks: 1,
		},
		logging.New(),
	)

	err := detector.DetectAndReportDrift(context.Background(), "i-123", nil)
	assert.NoError(t, err)
	assert.Len(t, reporter.reported, 1)
	assert.Len(t, repo.saved, 1)
	assert.True(t, repo.saved[0].HasDrift)
}

func TestDetectDriftByID_HandlesErrors(t *testing.T) {
	detector := app.NewDriftDetectorService(
		&mockInstanceProvider{err: errors.New("aws error")},
		&mockInstanceProvider{err: errors.New("tf error")},
		&mockRepository{},
		[]service.Reporter{},
		service.DriftDetectorConfig{
			SourceOfTruth: model.OriginAWS,
			Timeout:       2 * time.Second,
		},
		logging.New(),
	)

	_, err := detector.DetectDriftByID(context.Background(), "i-123", []string{"instance_type"})
	assert.Error(t, err)
}

func TestSettersAndGetters(t *testing.T) {
	detector := app.NewDriftDetectorService(nil, nil, nil, nil, service.DriftDetectorConfig{}, logging.New())

	detector.SetAttributePaths([]string{"tags.Name"})
	detector.SetParallelChecks(3)
	detector.SetTimeout(5 * time.Second)
	detector.SetScheduleExpression("@every 10s")
	detector.SetSourceOfTruth(model.OriginTerraform)

	assert.Equal(t, []string{"tags.Name"}, detector.GetAttributePaths())
	assert.Equal(t, 3, detector.GetParallelChecks())
	assert.Equal(t, 5*time.Second, detector.GetTimeout())
	assert.Equal(t, "@every 10s", detector.GetScheduleExpression())
	assert.Equal(t, model.OriginTerraform, detector.GetSourceOfTruth())
}

func TestStartAndStopScheduler(t *testing.T) {
	detector := app.NewDriftDetectorService(nil, nil, nil, nil, service.DriftDetectorConfig{
		ScheduleExpression: "@every 1m",
		Timeout:            1 * time.Second,
	}, logging.New())

	err := detector.StartScheduler(context.Background())
	assert.NoError(t, err)
	detector.StopScheduler()
}
