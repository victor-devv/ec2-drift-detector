package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestNewErrorHandler(t *testing.T) {
	// Setup
	logger := new(MockLogger)

	// Test creation
	handler := NewErrorHandler(logger)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
}

func TestHandleNilError(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	handler := NewErrorHandler(logger)

	// Test handling nil error (should do nothing)
	handler.Handle(nil)

	// Assert logger was not called
	logger.AssertNotCalled(t, "Error")
	logger.AssertNotCalled(t, "Warn")
	logger.AssertNotCalled(t, "Info")
	logger.AssertNotCalled(t, "Debug")
}

func TestHandleSystemError(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	logger.On("Error", mock.Anything, mock.Anything).Return()
	handler := NewErrorHandler(logger)

	// Create a system error
	sysErr := NewSystemError("Critical failure", nil)

	// Test handling system error with panic recovery
	defer func() {
		if r := recover(); r != nil {
			// Verify that panic message contains the error
			assert.Contains(t, r.(string), "System error: SYSTEM_ERROR: Critical failure")
		} else {
			t.Error("Expected panic but none occurred")
		}
	}()

	// This should panic
	handler.Handle(sysErr)

	// Assertions shouldn't reach here due to panic
	t.Error("Code execution continued after expected panic")
}

func TestHandleOperationalError(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	handler := NewErrorHandler(logger)

	// Create an operational error with context
	opErr := NewOperationalError("Operation failed", errors.New("connection timeout"))
	opErr.WithContext("operation", "api_call")

	// Test handling operational error
	handler.Handle(opErr)

	// Verify logger was called with error message
	logger.AssertCalled(t, "Error", mock.Anything, mock.Anything)
	logger.AssertCalled(t, "Debug", mock.Anything, mock.Anything)
}

func TestHandleValidationError(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	handler := NewErrorHandler(logger)

	// Create a validation error
	valErr := NewValidationError("Invalid input")
	valErr.WithContext("field", "email")

	// Test handling validation error
	handler.Handle(valErr)

	// Verify logger was called with warn message
	logger.AssertCalled(t, "Warn", mock.Anything, mock.Anything)
	logger.AssertCalled(t, "Debug", mock.Anything, mock.Anything)
}

func TestHandleNotFoundError(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	handler := NewErrorHandler(logger)

	// Create a not found error
	notFoundErr := NewNotFoundError("User", "123")

	// Test handling not found error
	handler.Handle(notFoundErr)

	// Verify logger was called with info message
	logger.AssertCalled(t, "Info", mock.Anything, mock.Anything)
	logger.AssertCalled(t, "Debug", mock.Anything, mock.Anything)
}

func TestHandleUnknownError(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	logger.On("Error", mock.Anything, mock.Anything).Return()
	handler := NewErrorHandler(logger)

	// Create a standard error
	stdErr := errors.New("Unknown error")

	// Test handling unknown error
	handler.Handle(stdErr)

	// Verify logger was called with error message (should be wrapped as operational)
	logger.AssertCalled(t, "Error", mock.Anything, mock.Anything)
}

func TestMustHandle(t *testing.T) {
	// Setup
	logger := new(MockLogger)
	logger.On("Error", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()
	handler := NewErrorHandler(logger)

	// Test case 1: Handling nil error
	handler.MustHandle(nil)

	// Test case 2: Handling operational error
	opErr := NewOperationalError("Operation failed", nil)
	handler.MustHandle(opErr)
	logger.AssertCalled(t, "Error", mock.Anything, mock.Anything)

	// Test case 3: Handling system error with panic recovery
	sysErr := NewSystemError("Critical failure", nil)

	defer func() {
		if r := recover(); r != nil {
			// Verify panic message
			assert.Contains(t, r.(string), "System error")
		} else {
			t.Error("Expected panic but none occurred")
		}
	}()

	// This should panic
	handler.MustHandle(sysErr)
}
