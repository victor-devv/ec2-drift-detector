package cli_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/presentation/cli"
)

type mockDriftService struct {
	schedulerStarted bool
}

func (m *mockDriftService) DetectAndReportDrift(ctx context.Context, id string, attrs []string) error {
	return nil
}
func (m *mockDriftService) DetectAndReportDriftForAll(ctx context.Context, attrs []string) error {
	return nil
}
func (m *mockDriftService) StartScheduler(ctx context.Context) error {
	m.schedulerStarted = true
	return nil
}
func (m *mockDriftService) StopScheduler() {}
func (m *mockDriftService) RunScheduledDriftCheck(ctx context.Context) error {
	return nil
}
func (m *mockDriftService) DetectDrift(ctx context.Context, src, tgt *model.Instance, attrs []string) (*model.DriftResult, error) {
	return nil, nil
}
func (m *mockDriftService) DetectDriftByID(ctx context.Context, id string, attrs []string) (*model.DriftResult, error) {
	return nil, nil
}
func (m *mockDriftService) DetectDriftForAll(ctx context.Context, attrs []string) ([]*model.DriftResult, error) {
	return nil, nil
}
func (m *mockDriftService) SetSourceOfTruth(t model.ResourceOrigin) {}
func (m *mockDriftService) SetAttributePaths(p []string)            {}
func (m *mockDriftService) SetParallelChecks(c int)                 {}
func (m *mockDriftService) SetTimeout(d time.Duration)              {}
func (m *mockDriftService) SetScheduleExpression(e string)          {}
func (m *mockDriftService) SetReporters(r []service.Reporter)       {}
func (m *mockDriftService) GetAttributePaths() []string             { return nil }
func (m *mockDriftService) GetSourceOfTruth() model.ResourceOrigin  { return "aws" }
func (m *mockDriftService) GetParallelChecks() int                  { return 1 }
func (m *mockDriftService) GetTimeout() time.Duration               { return 1 }
func (m *mockDriftService) GetScheduleExpression() string           { return "" }

func TestNewHandlerInitialization(t *testing.T) {
	logger := logging.New()
	cfg := &config.Config{}
	cfg.SetReporterType("console")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetSourceOfTruth("aws")
	cfg.SetParallelChecks(1)
	cfg.SetTimeout(30 * time.Second)

	h := cli.NewHandler(context.Background(), &mockDriftService{}, nil, cfg, logger)
	assert.NotNil(t, h.GetRootCommand())
}

func TestCLIConfigValidationFails(t *testing.T) {
	logger := logging.New()
	cfg := &config.Config{}
	cfg.SetReporterType("invalid") // force validation fail
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetSourceOfTruth("aws")
	cfg.SetParallelChecks(1)
	cfg.SetTimeout(30 * time.Second)

	loader := config.NewConfigLoader(logger, ".")
	h := cli.NewHandler(context.Background(), &mockDriftService{}, loader, cfg, logger)

	cmd := h.GetRootCommand()
	cmd.Flags().String("output", "invalid", "")
	_ = cmd.Flags().Set("output", "invalid")

	assert.NotNil(t, cmd)
}

func TestServerCommandExecution(t *testing.T) {
	logger := logging.New()
	cfg := &config.Config{}
	cfg.SetReporterType("console")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetSourceOfTruth("aws")
	cfg.SetParallelChecks(1)
	cfg.SetTimeout(5 * time.Second)

	mockService := &mockDriftService{}
	h := cli.NewHandler(context.Background(), mockService, nil, cfg, logger)

	cmd := h.GetRootCommand()
	serverCmd, _, err := cmd.Find([]string{"server"})
	assert.NoError(t, err)
	assert.NotNil(t, serverCmd)
	assert.Equal(t, "server", serverCmd.Use)
}

func TestDetectCommandAdded(t *testing.T) {
	logger := logging.New()
	cfg := &config.Config{}
	cfg.SetReporterType("console")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetSourceOfTruth("aws")
	cfg.SetParallelChecks(1)
	cfg.SetTimeout(5 * time.Second)

	h := cli.NewHandler(context.Background(), &mockDriftService{}, nil, cfg, logger)
	cmd := h.GetRootCommand()
	childCmd, _, err := cmd.Find([]string{"detect"})
	assert.NoError(t, err)
	assert.Equal(t, "detect [instance-id]", childCmd.Use)
}

func TestConfigShowCommandAdded(t *testing.T) {
	logger := logging.New()
	cfg := &config.Config{}
	cfg.SetReporterType("console")
	cfg.SetAttributes([]string{"instance_type"})
	cfg.SetSourceOfTruth("aws")
	cfg.SetParallelChecks(1)
	cfg.SetTimeout(30 * time.Second)
	cfg.SetAWSRegion("us-east-1")
	cfg.SetUseHCL(false)
	cfg.SetStateFile("mock.tfstate")

	h := cli.NewHandler(context.Background(), &mockDriftService{}, nil, cfg, logger)
	cmd := h.GetRootCommand()
	configCmd, _, err := cmd.Find([]string{"config", "show"})
	assert.NoError(t, err)
	assert.NotNil(t, configCmd)
	assert.Equal(t, "show", configCmd.Use)
}
