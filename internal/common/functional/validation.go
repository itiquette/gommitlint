// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package functional provides functional programming utilities for validation.
package functional

import (
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// ValidationCheck represents a validation function that returns optional errors.
type ValidationCheck func() []appErrors.ValidationError

// CombineErrors combines multiple error slices into one, filtering out nil/empty slices.
func CombineErrors(errorSlices ...[]appErrors.ValidationError) []appErrors.ValidationError {
	var totalLength int
	for _, errors := range errorSlices {
		totalLength += len(errors)
	}

	if totalLength == 0 {
		return nil
	}

	result := make([]appErrors.ValidationError, 0, totalLength)
	for _, errors := range errorSlices {
		result = append(result, errors...)
	}

	return result
}

// CheckIf performs a validation check only if the condition is true.
func CheckIf(condition bool, check ValidationCheck) []appErrors.ValidationError {
	if !condition {
		return nil
	}

	return check()
}

// CheckUnless performs a validation check unless the condition is true.
func CheckUnless(condition bool, check ValidationCheck) []appErrors.ValidationError {
	return CheckIf(!condition, check)
}

// FirstError returns the first non-empty error slice from the checks.
func FirstError(checks ...ValidationCheck) []appErrors.ValidationError {
	for _, check := range checks {
		if errors := check(); len(errors) > 0 {
			return errors
		}
	}

	return nil
}

// AllErrors runs all checks and combines their errors.
func AllErrors(checks ...ValidationCheck) []appErrors.ValidationError {
	errorSlices := make([][]appErrors.ValidationError, 0, len(checks))

	for _, check := range checks {
		if errors := check(); len(errors) > 0 {
			errorSlices = append(errorSlices, errors)
		}
	}

	return CombineErrors(errorSlices...)
}

// SingleError creates a slice with a single error.
func SingleError(err appErrors.ValidationError) []appErrors.ValidationError {
	return []appErrors.ValidationError{err}
}

// NoErrors returns an empty error slice.
func NoErrors() []appErrors.ValidationError {
	return nil
}
