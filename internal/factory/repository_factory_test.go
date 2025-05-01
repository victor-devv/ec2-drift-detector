package factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/factory"
)

func TestNewRepositoryFactory(t *testing.T) {
	logger := logging.New()
	f := factory.NewRepositoryFactory(logger)
	assert.NotNil(t, f)
}

func TestCreateDriftRepository(t *testing.T) {
	logger := logging.New()
	f := factory.NewRepositoryFactory(logger)
	repo := f.CreateDriftRepository()
	assert.NotNil(t, repo)
}

func TestCreateDriftRepositoryWithConfig(t *testing.T) {
	logger := logging.New()
	f := factory.NewRepositoryFactory(logger)
	cfg := &config.Config{}

	repo, err := f.CreateDriftRepositoryWithConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestCreateHistoricalDriftRepository(t *testing.T) {
	logger := logging.New()
	f := factory.NewRepositoryFactory(logger)
	cfg := &config.Config{}

	repo, err := f.CreateHistoricalDriftRepository(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestGetRepositoryStats(t *testing.T) {
	logger := logging.New()
	f := factory.NewRepositoryFactory(logger)
	repo := f.CreateDriftRepository()
	stats := f.GetRepositoryStats(repo)

	assert.Equal(t, "in-memory", stats["type"])
	assert.Equal(t, false, stats["persistent"])
	assert.Contains(t, stats, "count")
}

func TestClearRepository_Success(t *testing.T) {
	logger := logging.New()
	f := factory.NewRepositoryFactory(logger)
	repo := f.CreateDriftRepository()
	err := f.ClearRepository(repo)
	assert.NoError(t, err)
}
