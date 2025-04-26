package errors

import (
	"fmt"
)

// ErrorType defines the type of an error
type ErrorType string

const (
	// SystemError represents critical errors that should cause the application to panic
	SystemError ErrorType = "SYSTEM_ERROR"

	// OperationalError represents errors that can be handled gracefully
	OperationalError ErrorType = "OPERATIONAL_ERROR"

	// ValidationError represents invalid input or configuration
	ValidationError ErrorType = "VALIDATION_ERROR"

	// NotFoundError represents a resource not found error
	NotFoundError ErrorType = "NOT_FOUND_ERROR"
)

// AppError represents an application-specific error with contextual information
type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause of the error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds contextual information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewSystemError creates a new system error
func NewSystemError(message string, cause error) *AppError {
	return &AppError{
		Type:    SystemError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewOperationalError creates a new operational error
func NewOperationalError(message string, cause error) *AppError {
	return &AppError{
		Type:    OperationalError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:    ValidationError,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resourceType string, identifier string) *AppError {
	return &AppError{
		Type:    NotFoundError,
		Message: fmt.Sprintf("%s with ID '%s' not found", resourceType, identifier),
		Context: map[string]interface{}{
			"resourceType": resourceType,
			"identifier":   identifier,
		},
	}
}

// IsSystemError checks if an error is a system error
func IsSystemError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == SystemError
	}
	return false
}

// IsOperationalError checks if an error is an operational error
func IsOperationalError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == OperationalError
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ValidationError
	}
	return false
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == NotFoundError
	}
	return false
}
