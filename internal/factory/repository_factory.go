package factory

import (
	"fmt"

	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/config"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
	"github.com/victor-devv/ec2-drift-detector/internal/infrastructure/repository"
)

// RepositoryFactory creates repositories for data storage
type RepositoryFactory struct {
	logger *logging.Logger
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(logger *logging.Logger) *RepositoryFactory {
	return &RepositoryFactory{
		logger: logger,
	}
}

// CreateDriftRepository creates a repository for storing drift detection results
// Currently only supports in-memory repository, but could be extended to support
// persistent storage options in the future
func (f *RepositoryFactory) CreateDriftRepository() service.DriftRepository {
	f.logger.Info("Creating in-memory drift repository")
	return repository.NewInMemoryDriftRepository(f.logger)
}

// CreateDriftRepositoryWithConfig creates a repository based on configuration
// This is a placeholder for future extension to support different repository types
func (f *RepositoryFactory) CreateDriftRepositoryWithConfig(cfg *config.Config) (service.DriftRepository, error) {
	// This could be expanded in the future to support different repository types
	// based on configuration, such as file-based, database, etc.

	// For now, we always create an in-memory repository
	f.logger.Info("Creating in-memory drift repository from configuration")
	repo := repository.NewInMemoryDriftRepository(f.logger)

	// Log repository creation
	f.logger.Debug("Repository created: in-memory")
	f.logger.Debug("No persistence across restarts")

	return repo, nil
}

// CreateHistoricalDriftRepository is a placeholder for a potential future
// implementation that could store historical drift data
func (f *RepositoryFactory) CreateHistoricalDriftRepository(cfg *config.Config) (service.DriftRepository, error) {
	// This method is a placeholder for future implementation
	// It would create a repository that stores historical data
	// For now, return the same in-memory repository
	f.logger.Warn("Historical drift repository requested but not implemented - using in-memory")
	return f.CreateDriftRepositoryWithConfig(cfg)
}

// GetRepositoryStats returns statistics about the repository
// Useful for monitoring and debugging
func (f *RepositoryFactory) GetRepositoryStats(repo service.DriftRepository) map[string]interface{} {
	f.logger.Debug("Getting repository statistics")

	stats := make(map[string]interface{})

	// Add basic stats if available
	if countable, ok := repo.(interface{ Count() int }); ok {
		stats["count"] = countable.Count()
	}

	// Add repository type
	stats["type"] = "in-memory"
	stats["persistent"] = false

	return stats
}

// ClearRepository is a helper method to clear the repository
// Useful for testing and development
func (f *RepositoryFactory) ClearRepository(repo service.DriftRepository) error {
	f.logger.Info("Clearing repository")

	// Check if the repository supports clearing
	if clearable, ok := repo.(interface{ ClearResults() }); ok {
		clearable.ClearResults()
		return nil
	}

	// If not clearable, return an error
	return fmt.Errorf("repository does not support clearing")
}
