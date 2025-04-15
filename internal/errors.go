// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package internal

import "errors"

// Exit codes used in the application:
// 0 - success
// 1 - general error
// 2 - validation error
// 3 - git error
// 4 - configuration error
// 5 - input error

// ApplicationError represents an error with an associated exit code.
type ApplicationError struct {
	Cause    error
	Code     int
	Category string
	Context  map[string]string
}

// Error returns the error message.
func (e *ApplicationError) Error() string {
	if e.Cause == nil {
		return e.Category + " error"
	}

	return e.Cause.Error()
}

// ExitCode returns the appropriate exit code for the error.
func (e *ApplicationError) ExitCode() int {
	return e.Code
}

// WithContext adds context information to the error.
func (e *ApplicationError) WithContext(key, value string) *ApplicationError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}

	e.Context[key] = value

	return e
}

// Unwrap returns the underlying cause of the error.
func (e *ApplicationError) Unwrap() error {
	return e.Cause
}

// NewApplicationError creates a new application error.
func NewApplicationError(cause error, code int, category string) *ApplicationError {
	return &ApplicationError{
		Cause:    cause,
		Code:     code,
		Category: category,
		Context:  make(map[string]string),
	}
}

// NewGitError creates a new error related to Git operations.
func NewGitError(cause error, context ...map[string]string) *ApplicationError {
	err := NewApplicationError(cause, 3, "Git")

	if len(context) > 0 {
		for k, v := range context[0] {
			_ = err.WithContext(k, v)
		}
	}

	return err
}

// NewConfigError creates a new error related to configuration issues.
func NewConfigError(cause error, context ...map[string]string) *ApplicationError {
	err := NewApplicationError(cause, 4, "Configuration")

	if len(context) > 0 {
		for k, v := range context[0] {
			_ = err.WithContext(k, v)
		}
	}

	return err
}

// NewInputError creates a new error related to user input issues.
func NewInputError(cause error, context ...map[string]string) *ApplicationError {
	err := NewApplicationError(cause, 5, "Input")

	if len(context) > 0 {
		for k, v := range context[0] {
			_ = err.WithContext(k, v)
		}
	}

	return err
}

// NewValidationError creates a specialized application error for validation failures.
func NewValidationError(cause error, context ...map[string]string) *ApplicationError {
	err := NewApplicationError(cause, 2, "Validation")

	if len(context) > 0 {
		for k, v := range context[0] {
			_ = err.WithContext(k, v)
		}
	}

	return err
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == 2
	}

	return false
}

// IsGitError checks if an error is a Git error.
func IsGitError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == 3
	}

	return false
}

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == 4
	}

	return false
}

// IsInputError checks if an error is an input error.
func IsInputError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == 5
	}

	return false
}
