package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	// Test case 1: Create error with message only
	err := &AppError{
		Type:    ValidationError,
		Message: "Invalid input",
	}

	assert.Equal(t, ValidationError, err.Type)
	assert.Equal(t, "Invalid input", err.Message)
	assert.Nil(t, err.Cause)
	assert.Equal(t, "VALIDATION_ERROR: Invalid input", err.Error())

	// Test case 2: Create error with cause
	cause := errors.New("database connection failed")
	err = &AppError{
		Type:    SystemError,
		Message: "Database error",
		Cause:   cause,
	}

	assert.Equal(t, SystemError, err.Type)
	assert.Equal(t, "Database error", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.Equal(t, "SYSTEM_ERROR: Database error (cause: database connection failed)", err.Error())
	assert.Equal(t, cause, err.Unwrap())

	// Test case 3: Error with context
	err = &AppError{
		Type:    OperationalError,
		Message: "API call failed",
		Context: map[string]interface{}{
			"status_code": 500,
			"endpoint":    "/api/resource",
		},
	}

	assert.Equal(t, OperationalError, err.Type)
	assert.Contains(t, err.Context, "status_code")
	assert.Contains(t, err.Context, "endpoint")
	assert.Equal(t, 500, err.Context["status_code"])
}

func TestWithContext(t *testing.T) {
	// Setup
	err := NewOperationalError("API call failed", nil)

	// Test adding context
	result := err.WithContext("status_code", 500)
	assert.Equal(t, err, result) // Should return the same error
	assert.Contains(t, err.Context, "status_code")
	assert.Equal(t, 500, err.Context["status_code"])

	// Test adding another context value
	err.WithContext("endpoint", "/api/resource")
	assert.Contains(t, err.Context, "endpoint")
	assert.Equal(t, "/api/resource", err.Context["endpoint"])

	// Test overwriting existing context
	err.WithContext("status_code", 404)
	assert.Equal(t, 404, err.Context["status_code"])
}

func TestNewSystemError(t *testing.T) {
	// Test case 1: Without cause
	err := NewSystemError("Critical system failure", nil)

	assert.Equal(t, SystemError, err.Type)
	assert.Equal(t, "Critical system failure", err.Message)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Context)

	// Test case 2: With cause
	cause := errors.New("memory allocation failed")
	err = NewSystemError("Memory error", cause)

	assert.Equal(t, SystemError, err.Type)
	assert.Equal(t, "Memory error", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewOperationalError(t *testing.T) {
	// Test case 1: Without cause
	err := NewOperationalError("Operation timeout", nil)

	assert.Equal(t, OperationalError, err.Type)
	assert.Equal(t, "Operation timeout", err.Message)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Context)

	// Test case 2: With cause
	cause := errors.New("connection reset")
	err = NewOperationalError("Network error", cause)

	assert.Equal(t, OperationalError, err.Type)
	assert.Equal(t, "Network error", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("Invalid parameter")

	assert.Equal(t, ValidationError, err.Type)
	assert.Equal(t, "Invalid parameter", err.Message)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Context)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("User", "123")

	assert.Equal(t, NotFoundError, err.Type)
	assert.Equal(t, "User with ID '123' not found", err.Message)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Context)
	assert.Equal(t, "User", err.Context["resourceType"])
	assert.Equal(t, "123", err.Context["identifier"])
}

func TestErrorTypeChecks(t *testing.T) {
	// Setup
	sysErr := NewSystemError("System error", nil)
	opErr := NewOperationalError("Operational error", nil)
	valErr := NewValidationError("Validation error")
	notFoundErr := NewNotFoundError("Resource", "123")
	stdErr := errors.New("Standard error")

	// Test IsSystemError
	assert.True(t, IsSystemError(sysErr))
	assert.False(t, IsSystemError(opErr))
	assert.False(t, IsSystemError(valErr))
	assert.False(t, IsSystemError(notFoundErr))
	assert.False(t, IsSystemError(stdErr))
	assert.False(t, IsSystemError(nil))

	// Test IsOperationalError
	assert.False(t, IsOperationalError(sysErr))
	assert.True(t, IsOperationalError(opErr))
	assert.False(t, IsOperationalError(valErr))
	assert.False(t, IsOperationalError(notFoundErr))
	assert.False(t, IsOperationalError(stdErr))
	assert.False(t, IsOperationalError(nil))

	// Test IsValidationError
	assert.False(t, IsValidationError(sysErr))
	assert.False(t, IsValidationError(opErr))
	assert.True(t, IsValidationError(valErr))
	assert.False(t, IsValidationError(notFoundErr))
	assert.False(t, IsValidationError(stdErr))
	assert.False(t, IsValidationError(nil))

	// Test IsNotFoundError
	assert.False(t, IsNotFoundError(sysErr))
	assert.False(t, IsNotFoundError(opErr))
	assert.False(t, IsNotFoundError(valErr))
	assert.True(t, IsNotFoundError(notFoundErr))
	assert.False(t, IsNotFoundError(stdErr))
	assert.False(t, IsNotFoundError(nil))
}
