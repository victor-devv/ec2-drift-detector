package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/victor-devv/ec2-drift-detector/internal/common/errors"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

// InMemoryDriftRepository is an in-memory implementation of the DriftRepository interface
type InMemoryDriftRepository struct {
	// results is a map of result ID to result
	results map[string]*model.DriftResult

	// instanceResults is a map of instance ID to result IDs
	instanceResults map[string][]string

	// mutex for thread safety
	mu sync.RWMutex

	// logger
	logger *logging.Logger
}

// NewInMemoryDriftRepository creates a new in-memory drift repository
func NewInMemoryDriftRepository(logger *logging.Logger) *InMemoryDriftRepository {
	return &InMemoryDriftRepository{
		results:         make(map[string]*model.DriftResult),
		instanceResults: make(map[string][]string),
		logger:          logger.WithField("component", "inmemory-drift-repo"),
	}
}

// SaveDriftResult saves a drift detection result
func (r *InMemoryDriftRepository) SaveDriftResult(ctx context.Context, result *model.DriftResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store the result
	r.results[result.ID] = result

	// Add the result ID to the instance's results
	r.instanceResults[result.ResourceID] = append(r.instanceResults[result.ResourceID], result.ID)

	r.logger.Debug(fmt.Sprintf("Saved drift result %s for instance %s", result.ID, result.ResourceID))
	return nil
}

// GetDriftResult retrieves a drift detection result by ID
func (r *InMemoryDriftRepository) GetDriftResult(ctx context.Context, id string) (*model.DriftResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get the result
	result, ok := r.results[id]
	if !ok {
		return nil, errors.NewNotFoundError("DriftResult", id)
	}

	return result, nil
}

// GetDriftResultsByInstanceID retrieves drift detection results by instance ID
func (r *InMemoryDriftRepository) GetDriftResultsByInstanceID(ctx context.Context, instanceID string) ([]*model.DriftResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get the result IDs for the instance
	resultIDs, ok := r.instanceResults[instanceID]
	if !ok {
		return nil, errors.NewNotFoundError("DriftResults for Instance", instanceID)
	}

	// Get the results
	results := make([]*model.DriftResult, 0, len(resultIDs))
	for _, id := range resultIDs {
		result, ok := r.results[id]
		if ok {
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return nil, errors.NewNotFoundError("DriftResults for Instance", instanceID)
	}

	return results, nil
}

// ListDriftResults retrieves all drift detection results
func (r *InMemoryDriftRepository) ListDriftResults(ctx context.Context) ([]*model.DriftResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get all results
	results := make([]*model.DriftResult, 0, len(r.results))
	for _, result := range r.results {
		results = append(results, result)
	}

	return results, nil
}

// ClearResults clears all results
func (r *InMemoryDriftRepository) ClearResults() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results = make(map[string]*model.DriftResult)
	r.instanceResults = make(map[string][]string)
}

// Count returns the number of results
func (r *InMemoryDriftRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.results)
}
