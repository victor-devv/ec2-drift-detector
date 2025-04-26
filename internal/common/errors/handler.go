package errors

import (
	"fmt"
	"os"
	"runtime/debug"
)

// ErrorHandler defines how to handle different types of errors
type ErrorHandler struct {
	logger Logger
}

// Logger defines the minimal logging interface required by ErrorHandler
type Logger interface {
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// NewErrorHandler creates a new error handler with the provided logger
func NewErrorHandler(logger Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle handles an error based on its type
func (h *ErrorHandler) Handle(err error) {
	if err == nil {
		return
	}

	var appErr *AppError
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		// Wrap unknown errors as operational errors
		appErr = NewOperationalError("Unknown error occurred", err)
	}

	switch appErr.Type {
	case SystemError:
		h.handleSystemError(appErr)
	case OperationalError:
		h.handleOperationalError(appErr)
	case ValidationError:
		h.handleValidationError(appErr)
	case NotFoundError:
		h.handleNotFoundError(appErr)
	default:
		h.handleOperationalError(appErr)
	}
}

// handleSystemError handles system errors by logging and panicking
func (h *ErrorHandler) handleSystemError(err *AppError) {
	stackTrace := string(debug.Stack())
	h.logger.Error("SYSTEM ERROR: %s (cause: %v)", err.Message, err.Cause)
	h.logger.Error("Stack trace: %s", stackTrace)

	// System errors should cause application to panic
	panic(fmt.Sprintf("System error: %s", err.Error()))
}

// handleOperationalError handles operational errors by logging
func (h *ErrorHandler) handleOperationalError(err *AppError) {
	h.logger.Error("OPERATIONAL ERROR: %s (cause: %v)", err.Message, err.Cause)

	// Log additional context if available
	if len(err.Context) > 0 {
		h.logger.Debug("Error context: %v", err.Context)
	}
}

// handleValidationError handles validation errors
func (h *ErrorHandler) handleValidationError(err *AppError) {
	h.logger.Warn("VALIDATION ERROR: %s", err.Message)

	// Log additional context if available
	if len(err.Context) > 0 {
		h.logger.Debug("Validation context: %v", err.Context)
	}
}

// handleNotFoundError handles not found errors
func (h *ErrorHandler) handleNotFoundError(err *AppError) {
	h.logger.Info("NOT FOUND: %s", err.Message)

	// Log additional context if available
	if len(err.Context) > 0 {
		h.logger.Debug("Not found context: %v", err.Context)
	}
}

// HandleWithExit handles an error and exits the program with the appropriate exit code if needed
func (h *ErrorHandler) HandleWithExit(err error) {
	if err == nil {
		return
	}

	h.Handle(err)

	// Exit with non-zero status code for system errors
	if IsSystemError(err) {
		os.Exit(1)
	}
}

// MustHandle handles an error and panics if it's a system error
func (h *ErrorHandler) MustHandle(err error) {
	if err == nil {
		return
	}

	if IsSystemError(err) {
		h.handleSystemError(err.(*AppError))
	} else {
		h.Handle(err)
	}
}
