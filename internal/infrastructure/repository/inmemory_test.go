package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/common/logging"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/model"
)

func TestInMemoryDriftRepository(t *testing.T) {
	// Create a repository
	repo := NewInMemoryDriftRepository(logging.New())
	ctx := context.Background()

	// Create test drift results
	result1 := model.NewDriftResult("i-12345", model.OriginTerraform)
	result1.AddDriftedAttribute("instance_type", "t2.micro", "t2.small")

	result2 := model.NewDriftResult("i-12345", model.OriginTerraform)
	result2.AddDriftedAttribute("ami", "ami-12345", "ami-67890")

	result3 := model.NewDriftResult("i-67890", model.OriginTerraform)
	// No drift detected

	// Test SaveDriftResult
	err := repo.SaveDriftResult(ctx, result1)
	require.NoError(t, err)

	err = repo.SaveDriftResult(ctx, result2)
	require.NoError(t, err)

	err = repo.SaveDriftResult(ctx, result3)
	require.NoError(t, err)

	// Test GetDriftResult
	retrievedResult, err := repo.GetDriftResult(ctx, result1.ID)
	require.NoError(t, err)
	require.Equal(t, result1.ID, retrievedResult.ID)
	require.Equal(t, "i-12345", retrievedResult.ResourceID)
	require.True(t, retrievedResult.HasDrift)
	require.Len(t, retrievedResult.DriftedAttributes, 1)
	require.Contains(t, retrievedResult.DriftedAttributes, "instance_type")

	// Test GetDriftResult with non-existent ID
	_, err = repo.GetDriftResult(ctx, "non-existent")
	require.Error(t, err)

	// Test GetDriftResultsByInstanceID
	results, err := repo.GetDriftResultsByInstanceID(ctx, "i-12345")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Test GetDriftResultsByInstanceID with non-existent instance ID
	_, err = repo.GetDriftResultsByInstanceID(ctx, "non-existent")
	require.Error(t, err)

	// Test ListDriftResults
	allResults, err := repo.ListDriftResults(ctx)
	require.NoError(t, err)
	require.Len(t, allResults, 3)

	// Test Count
	require.Equal(t, 3, repo.Count())

	// Test ClearResults
	repo.ClearResults()
	require.Equal(t, 0, repo.Count())

	// Verify all results are cleared
	allResults, err = repo.ListDriftResults(ctx)
	require.NoError(t, err)
	require.Len(t, allResults, 0)
}

func TestInMemoryDriftRepository_ConcurrentAccess(t *testing.T) {
	// Create a repository
	repo := NewInMemoryDriftRepository(logging.New())
	ctx := context.Background()

	// Create a large number of results
	const numResults = 100
	resultIDs := make([]string, numResults)

	// Save results concurrently
	done := make(chan bool)
	for i := 0; i < numResults; i++ {
		go func(index int) {
			instanceID := "i-12345"
			if index%2 == 0 {
				instanceID = "i-67890"
			}

			result := model.NewDriftResult(instanceID, model.OriginTerraform)
			err := repo.SaveDriftResult(ctx, result)
			require.NoError(t, err)
			resultIDs[index] = result.ID
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numResults; i++ {
		<-done
	}

	// Verify all results were saved
	require.Equal(t, numResults, repo.Count())

	// Read results concurrently
	readDone := make(chan bool)
	for i := 0; i < numResults; i++ {
		go func(index int) {
			// Randomly choose between different read operations
			switch index % 3 {
			case 0:
				// Get individual result
				result, err := repo.GetDriftResult(ctx, resultIDs[index])
				require.NoError(t, err)
				require.NotNil(t, result)
			case 1:
				// Get results by instance ID
				results, err := repo.GetDriftResultsByInstanceID(ctx, "i-12345")
				require.NoError(t, err)
				require.NotEmpty(t, results)
			case 2:
				// List all results
				results, err := repo.ListDriftResults(ctx)
				require.NoError(t, err)
				require.Len(t, results, numResults)
			}
			readDone <- true
		}(i)
	}

	// Wait for all read operations to complete
	for i := 0; i < numResults; i++ {
		<-readDone
	}
}

// func TestInMemoryDriftRepository_EdgeCases(t *testing.T) {
// 	// Create a repository
// 	repo := NewInMemoryDriftRepository(logging.New())
// 	ctx := context.Background()

// 	// Test with nil result
// 	err := repo.SaveDriftResult(ctx, nil)
// 	require.NoError(t, err) // No error but nothing saved

// 	// Verify nothing was saved
// 	require.Equal(t, 0, repo.Count())

// 	// Create an empty result
// 	emptyResult := model.NewDriftResult("", model.OriginTerraform)
// 	err = repo.SaveDriftResult(ctx, emptyResult)
// 	require.NoError(t, err)

// 	// Verify it was saved despite empty instance ID
// 	require.Equal(t, 1, repo.Count())

// 	// Test edge case: GetDriftResultsByInstanceID with empty instance ID
// 	results, err := repo.GetDriftResultsByInstanceID(ctx, "")
// 	require.NoError(t, err)
// 	require.Len(t, results, 1)

// 	// Test edge case: ListDriftResults with no results after clearing
// 	repo.ClearResults()
// 	results, err = repo.ListDriftResults(ctx)
// 	require.NoError(t, err)
// 	require.Len(t, results, 0)
// }

func TestInMemoryDriftRepository_NoResultsForInstance(t *testing.T) {
	// Create a repository
	repo := NewInMemoryDriftRepository(logging.New())
	ctx := context.Background()

	// Save a result for one instance ID
	result := model.NewDriftResult("i-12345", model.OriginTerraform)
	err := repo.SaveDriftResult(ctx, result)
	require.NoError(t, err)

	// Check that a different instance ID has no results
	_, err = repo.GetDriftResultsByInstanceID(ctx, "i-67890")
	require.Error(t, err)

	// Add a result and then remove it
	result2 := model.NewDriftResult("i-67890", model.OriginTerraform)
	err = repo.SaveDriftResult(ctx, result2)
	require.NoError(t, err)

	// Now we should have results for both instances
	results1, err := repo.GetDriftResultsByInstanceID(ctx, "i-12345")
	require.NoError(t, err)
	require.Len(t, results1, 1)

	results2, err := repo.GetDriftResultsByInstanceID(ctx, "i-67890")
	require.NoError(t, err)
	require.Len(t, results2, 1)

	// Clear and verify both are gone
	repo.ClearResults()
	_, err = repo.GetDriftResultsByInstanceID(ctx, "i-12345")
	require.Error(t, err)
	_, err = repo.GetDriftResultsByInstanceID(ctx, "i-67890")
	require.Error(t, err)
}

func TestInMemoryDriftRepository_MultipleResultsPerInstance(t *testing.T) {
	// Create a repository
	repo := NewInMemoryDriftRepository(logging.New())
	ctx := context.Background()

	// Create multiple results for the same instance ID
	const numResults = 5
	for i := 0; i < numResults; i++ {
		result := model.NewDriftResult("i-12345", model.OriginTerraform)
		result.AddDriftedAttribute(fmt.Sprintf("attr%d", i), "old", "new")
		err := repo.SaveDriftResult(ctx, result)
		require.NoError(t, err)
	}

	// Get results for the instance
	results, err := repo.GetDriftResultsByInstanceID(ctx, "i-12345")
	require.NoError(t, err)
	require.Len(t, results, numResults)

	// Verify each result has a different ID but the same instance ID
	ids := make(map[string]bool)
	for _, result := range results {
		require.Equal(t, "i-12345", result.ResourceID)
		require.False(t, ids[result.ID], "Duplicate result ID found")
		ids[result.ID] = true
	}
}
