package detector

import (
	"context"
	"sync"

	"github.com/victor-devv/ec2-drift-detector/internal/models"
)

// Detector interface defines the contract for all drift detectors
type Detector interface {
	// DetectDrift checks for differences between AWS resources and Terraform definitions
	DetectDrift(ctx context.Context, attributes []string) ([]models.DriftResult, error)

	// DetectDriftConcurrent checks for differences concurrently for multiple resources
	DetectDriftConcurrent(ctx context.Context, attributes []string) ([]models.DriftResult, error)
}

// BaseDetector provides common functionality for all detectors
type BaseDetector struct {
}

// runConcurrent executes the provided function concurrently for each item in the items slice
func runConcurrent[T any, R any](ctx context.Context, items []T, concurrency int, fn func(context.Context, T) (R, error)) ([]R, error) {
	if len(items) == 0 {
		return []R{}, nil
	}

	if concurrency <= 0 {
		concurrency = len(items)
	}

	var wg sync.WaitGroup
	jobs := make(chan T, len(items))
	results := make(chan R, len(items))
	errors := make(chan error, len(items))

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				select {
				case <-ctx.Done():
					errors <- ctx.Err()
					return
				default:
					result, err := fn(ctx, item)
					if err != nil {
						errors <- err
						return
					}
					results <- result
				}
			}
		}()
	}

	// Send jobs
	for _, item := range items {
		jobs <- item
	}
	close(jobs)

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	var resultList []R
	for r := range results {
		resultList = append(resultList, r)
	}

	// Check for errors
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return resultList, nil
}
