// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errorx

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// CreateValidationError creates a standard validation error with context.
func CreateValidationError(
	rule string,
	code domain.ValidationErrorCode,
	message string,
	context map[string]string,
) *domain.ValidationError {
	// Create a standard error
	err := domain.NewStandardValidationError(rule, code, message)

	// Add any context values
	for key, value := range context {
		err = err.WithContext(key, value)
	}

	return err
}

// AddTemplatedError adds a templated error to an existing validation error collection.
func AddTemplatedError(
	errors []*domain.ValidationError,
	rule string,
	template ErrorTemplate,
	args ...interface{},
) []*domain.ValidationError {
	err := NewValidationError(rule, template, args...)

	return append(errors, err)
}

// AddError is a helper that creates a validation error and adds it to an error collection.
func AddError(
	errors []*domain.ValidationError,
	rule string,
	code domain.ValidationErrorCode,
	message string,
	context map[string]string,
) []*domain.ValidationError {
	err := CreateValidationError(rule, code, message, context)

	return append(errors, err)
}
