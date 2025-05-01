package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victor-devv/ec2-drift-detector/internal/app"
	"github.com/victor-devv/ec2-drift-detector/internal/domain/service"
)

// mockDriftDetector is a dummy implementation of the DriftDetectorProvider interface
type mockDriftDetector struct {
	service.DriftDetectorProvider
}

func TestNewApplication(t *testing.T) {
	mock := &mockDriftDetector{}

	appInstance := app.NewApplication(mock)

	assert.NotNil(t, appInstance)
	assert.Equal(t, mock, appInstance.DriftDetector)
}
