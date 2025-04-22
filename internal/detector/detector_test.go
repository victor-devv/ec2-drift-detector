package detector_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/victor-devv/ec2-drift-detector/internal/detector"
)

func TestRunConcurrent_AllSuccess(t *testing.T) {
	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	var count int32
	results, err := detector.RunConcurrent(ctx, items, 2, func(ctx context.Context, i int) (string, error) {
		atomic.AddInt32(&count, 1)
		return fmt.Sprintf("done-%d", i), nil
	})

	require.NoError(t, err)
	require.Len(t, results, len(items))
	require.Equal(t, int32(len(items)), count)
}

func TestRunConcurrent_ErrorShortCircuits(t *testing.T) {
	ctx := context.Background()
	items := []int{1, 2, 3}

	results, err := detector.RunConcurrent(ctx, items, 2, func(ctx context.Context, i int) (string, error) {
		if i == 2 {
			return "", errors.New("boom")
		}
		time.Sleep(50 * time.Millisecond) // simulate work
		return fmt.Sprintf("ok-%d", i), nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
	require.Nil(t, results)
}

func TestRunConcurrent_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	items := []int{1, 2, 3}

	// cancel before work starts
	cancel()

	results, err := detector.RunConcurrent(ctx, items, 3, func(ctx context.Context, i int) (string, error) {
		return fmt.Sprintf("won't-run-%d", i), nil
	})

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, results)
}

func TestRunConcurrent_EmptyInput(t *testing.T) {
	ctx := context.Background()

	results, err := detector.RunConcurrent(ctx, []int{}, 5, func(ctx context.Context, i int) (string, error) {
		return "should not run", nil
	})

	require.NoError(t, err)
	require.Empty(t, results)
}

func TestRunConcurrent_ConcurrencyCap(t *testing.T) {
	ctx := context.Background()
	items := make([]int, 10)
	for i := range items {
		items[i] = i + 1
	}

	maxConcurrent := int32(0)
	currentRunning := int32(0)

	results, err := detector.RunConcurrent(ctx, items, 3, func(ctx context.Context, i int) (string, error) {
		n := atomic.AddInt32(&currentRunning, 1)
		if n > maxConcurrent {
			maxConcurrent = n
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&currentRunning, -1)
		return fmt.Sprintf("task-%d", i), nil
	})

	require.NoError(t, err)
	require.Len(t, results, len(items))
	require.LessOrEqual(t, maxConcurrent, int32(3))
}
