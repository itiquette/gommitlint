// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
// Package rules provides validation rules for Git commits.
package rules

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SubjectLengthRule validates that the commit subject is not too long.
type SubjectLengthRule struct {
	maxLength int
	errors    []appErrors.ValidationError
}

// NewSubjectLengthRule creates a new SubjectLengthRule.
func NewSubjectLengthRule(maxLength int) SubjectLengthRule {
	// Use default if not specified
	if maxLength <= 0 {
		maxLength = 100 // Default max length
	}

	return SubjectLengthRule{
		maxLength: maxLength,
		errors:    make([]appErrors.ValidationError, 0),
	}
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return "SubjectLength"
}

// SetErrors sets the validation errors after validation.
// This is needed to support value receivers while maintaining state.
func (r SubjectLengthRule) SetErrors(errors []appErrors.ValidationError) SubjectLengthRule {
	r.errors = errors

	return r
}

// Validate validates the commit subject length.
func (r SubjectLengthRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Create a new error slice instead of modifying r.errors
	errors := make([]appErrors.ValidationError, 0)

	// Calculate subject length
	subjectLength := utf8.RuneCountInString(commit.Subject)

	// Check if subject is too long
	if subjectLength > r.maxLength {
		// Create error with context in one step
		context := map[string]string{
			"actual_length": strconv.Itoa(subjectLength),
			"max_length":    strconv.Itoa(r.maxLength),
		}

		// Create error with app errors package
		err := appErrors.New(
			r.Name(),
			appErrors.ErrSubjectTooLong,
			fmt.Sprintf("Subject length (%d) exceeds maximum allowed (%d)", subjectLength, r.maxLength),
			appErrors.WithContextMap(context),
		)

		// Add to errors
		errors = append(errors, err)
	}

	return errors
}

// Result returns a concise result message.
func (r SubjectLengthRule) Result() string {
	if len(r.errors) > 0 {
		return "Subject too long"
	}

	return "Subject length OK"
}

// VerboseResult returns a detailed result message.
func (r SubjectLengthRule) VerboseResult() string {
	if len(r.errors) > 0 {
		// Get context values
		actualLength := "unknown"
		maxLength := strconv.Itoa(r.maxLength)

		if ctx := r.errors[0].Context; ctx != nil {
			if val, ok := ctx["actual_length"]; ok {
				actualLength = val
			}

			if val, ok := ctx["max_length"]; ok {
				maxLength = val
			}
		}

		return fmt.Sprintf("Subject length (%s characters) exceeds maximum length (%s characters)", actualLength, maxLength)
	}

	return fmt.Sprintf("Subject length is within the limit of %d characters", r.maxLength)
}

// Help returns guidance on how to fix rule violations.
func (r SubjectLengthRule) Help() string {
	if len(r.errors) == 0 {
		return "No help needed"
	}

	// Get max length from context
	maxLength := r.maxLength

	if ctx := r.errors[0].Context; ctx != nil {
		if val, ok := ctx["max_length"]; ok {
			if parsedVal, err := strconv.Atoi(val); err == nil {
				maxLength = parsedVal
			}
		}
	}

	// Use the template-based help message
	return appErrors.FormatHelp(appErrors.ErrSubjectTooLong, maxLength)
}

// Errors returns all validation errors.
func (r SubjectLengthRule) Errors() []appErrors.ValidationError {
	return r.errors
}
