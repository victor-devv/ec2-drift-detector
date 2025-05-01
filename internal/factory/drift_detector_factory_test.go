package factory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/factory"
)

// Mock implementations
type mockInstanceProvider struct {
	mock.Mock
}

func (m *mockInstanceProvider) GetInstance(ctx context.Context, id string) (*model.Instance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Instance), args.Error(1)
}

func (m *mockInstanceProvider) ListInstances(ctx context.Context) ([]*model.Instance, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Instance), args.Error(1)
}

type mockDriftRepository struct {
	mock.Mock
}

func (m *mockDriftRepository) SaveDriftResult(ctx context.Context, result *model.DriftResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *mockDriftRepository) GetDriftResult(ctx context.Context, resourceID string) (*model.DriftResult, error) {
	args := m.Called(ctx, resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriftResult), args.Error(1)
}

func (m *mockDriftRepository) ListDriftResults(ctx context.Context) ([]*model.DriftResult, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.DriftResult), args.Error(1)
}

func (m *mockDriftRepository) GetDriftResultsByInstanceID(ctx context.Context, instanceID string) ([]*model.DriftResult, error) {
	args := m.Called(ctx, instanceID)
	return args.Get(0).([]*model.DriftResult), args.Error(1)
}

type mockReporter struct {
	mock.Mock
}

func (m *mockReporter) ReportDrift(result *model.DriftResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *mockReporter) ReportMultipleDrifts(results []*model.DriftResult) error {
	args := m.Called(results)
	return args.Error(0)
}

type mockDriftDetector struct {
	mock.Mock
}

func (m *mockDriftDetector) DetectDrift(ctx context.Context, source, target *model.Instance, attributePaths []string) (*model.DriftResult, error) {
	args := m.Called(ctx, source, target, attributePaths)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriftResult), args.Error(1)
}

func (m *mockDriftDetector) DetectDriftByID(ctx context.Context, instanceID string, attributePaths []string) (*model.DriftResult, error) {
	args := m.Called(ctx, instanceID, attributePaths)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DriftResult), args.Error(1)
}

func (m *mockDriftDetector) DetectDriftForAll(ctx context.Context, attributePaths []string) ([]*model.DriftResult, error) {
	args := m.Called(ctx, attributePaths)
	return args.Get(0).([]*model.DriftResult), args.Error(1)
}

func (m *mockDriftDetector) DetectAndReportDrift(ctx context.Context, instanceID string, attributePaths []string) error {
	args := m.Called(ctx, instanceID, attributePaths)
	return args.Error(0)
}

func (m *mockDriftDetector) DetectAndReportDriftForAll(ctx context.Context, attributePaths []string) error {
	args := m.Called(ctx, attributePaths)
	return args.Error(0)
}

func (m *mockDriftDetector) RunScheduledDriftCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDriftDetector) StartScheduler(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDriftDetector) StopScheduler() {
	m.Called()
}

func (m *mockDriftDetector) SetSourceOfTruth(sourceOfTruth model.ResourceOrigin) {
	m.Called(sourceOfTruth)
}

func (m *mockDriftDetector) SetAttributePaths(attributePaths []string) {
	m.Called(attributePaths)
}

func (m *mockDriftDetector) SetParallelChecks(parallelChecks int) {
	m.Called(parallelChecks)
}

func (m *mockDriftDetector) SetTimeout(timeout time.Duration) {
	m.Called(timeout)
}

func (m *mockDriftDetector) SetScheduleExpression(expression string) {
	m.Called(expression)
}

func (m *mockDriftDetector) GetAttributePaths() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockDriftDetector) GetSourceOfTruth() model.ResourceOrigin {
	args := m.Called()
	return args.Get(0).(model.ResourceOrigin)
}

func (m *mockDriftDetector) GetParallelChecks() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockDriftDetector) GetTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockDriftDetector) GetScheduleExpression() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockDriftDetector) SetReporters(reporters []service.Reporter) {
	m.Called(reporters)
}

func TestNewDriftDetectorFactory(t *testing.T) {
	logger := logging.New()

	factory := factory.NewDriftDetectorFactory(logger)

	assert.NotNil(t, factory, "Factory should not be nil")
}

func TestCreateDriftDetector_Success(t *testing.T) {
	logger := logging.New()
	awsProvider := new(mockInstanceProvider)
	terraformProvider := new(mockInstanceProvider)
	repository := new(mockDriftRepository)
	reporters := []service.Reporter{new(mockReporter)}

	cfg := &config.Config{}
	cfg.SetSourceOfTruth("aws")
	cfg.SetAttributes([]string{"instance_type", "security_groups"})
	cfg.SetParallelChecks(5)
	cfg.SetTimeout(30 * time.Second)
	cfg.SetScheduleExpression("*/30 * * * *")

	mockDetector := new(mockDriftDetector)

	serviceFactory := func(
		awsProvider service.InstanceProvider,
		terraformProvider service.InstanceProvider,
		repository service.DriftRepository,
		reporters []service.Reporter,
		config service.DriftDetectorConfig,
		logger *logging.Logger,
	) service.DriftDetectorProvider {
		// Verify the configuration was correctly passed
		assert.Equal(t, model.OriginAWS, config.SourceOfTruth)
		assert.Equal(t, []string{"instance_type", "security_groups"}, config.AttributePaths)
		assert.Equal(t, 5, config.ParallelChecks)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, "*/30 * * * *", config.ScheduleExpression)

		return mockDetector
	}

	factory := factory.NewDriftDetectorFactory(logger)

	detector, err := factory.CreateDriftDetector(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		cfg,
		serviceFactory,
	)

	assert.NoError(t, err)
	assert.Equal(t, mockDetector, detector)
}

func TestCreateDriftDetector_NilFactory(t *testing.T) {
	logger := logging.New()
	awsProvider := new(mockInstanceProvider)
	terraformProvider := new(mockInstanceProvider)
	repository := new(mockDriftRepository)
	reporters := []service.Reporter{new(mockReporter)}

	cfg := &config.Config{}
	cfg.SetSourceOfTruth("aws")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetParallelChecks(5)
	cfg.SetTimeout(30 * time.Second)
	cfg.SetScheduleExpression("*/30 * * * *")

	factory := factory.NewDriftDetectorFactory(logger)

	detector, err := factory.CreateDriftDetector(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		cfg,
		nil, // Nil factory function
	)

	assert.Error(t, err)
	assert.Nil(t, detector)
	assert.Contains(t, err.Error(), "drift detector service factory is nil")
}

func TestCreateDriftDetectorWithCustomConfig_Success(t *testing.T) {
	logger := logging.New()
	awsProvider := new(mockInstanceProvider)
	terraformProvider := new(mockInstanceProvider)
	repository := new(mockDriftRepository)
	reporters := []service.Reporter{new(mockReporter)}

	customConfig := service.DriftDetectorConfig{
		SourceOfTruth:      model.OriginTerraform,
		AttributePaths:     []string{"tags", "volume"},
		ParallelChecks:     10,
		Timeout:            60 * time.Second,
		ScheduleExpression: "0 * * * *",
	}

	mockDetector := new(mockDriftDetector)

	serviceFactory := func(
		awsProvider service.InstanceProvider,
		terraformProvider service.InstanceProvider,
		repository service.DriftRepository,
		reporters []service.Reporter,
		config service.DriftDetectorConfig,
		logger *logging.Logger,
	) service.DriftDetectorProvider {
		// Verify the configuration was correctly passed
		assert.Equal(t, model.OriginTerraform, config.SourceOfTruth)
		assert.Equal(t, []string{"tags", "volume"}, config.AttributePaths)
		assert.Equal(t, 10, config.ParallelChecks)
		assert.Equal(t, 60*time.Second, config.Timeout)
		assert.Equal(t, "0 * * * *", config.ScheduleExpression)

		return mockDetector
	}

	factory := factory.NewDriftDetectorFactory(logger)

	detector, err := factory.CreateDriftDetectorWithCustomConfig(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		customConfig,
		serviceFactory,
	)

	assert.NoError(t, err)
	assert.Equal(t, mockDetector, detector)
}

func TestCreateDriftDetectorWithCustomConfig_NilFactory(t *testing.T) {
	logger := logging.New()
	awsProvider := new(mockInstanceProvider)
	terraformProvider := new(mockInstanceProvider)
	repository := new(mockDriftRepository)
	reporters := []service.Reporter{new(mockReporter)}

	customConfig := service.DriftDetectorConfig{
		SourceOfTruth:      model.OriginTerraform,
		AttributePaths:     []string{"tags", "volume"},
		ParallelChecks:     10,
		Timeout:            60 * time.Second,
		ScheduleExpression: "0 * * * *",
	}

	factory := factory.NewDriftDetectorFactory(logger)

	detector, err := factory.CreateDriftDetectorWithCustomConfig(
		awsProvider,
		terraformProvider,
		repository,
		reporters,
		customConfig,
		nil, // Nil factory function
	)

	assert.Error(t, err)
	assert.Nil(t, detector)
	assert.Contains(t, err.Error(), "drift detector service factory is nil")
}
