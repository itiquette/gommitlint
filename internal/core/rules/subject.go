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
	"github.com/itiquette/gommitlint/internal/errorx"
)

// SubjectLengthRule validates that the commit subject is not too long.
type SubjectLengthRule struct {
	maxLength int
	errors    []*domain.ValidationError
}

// NewSubjectLengthRule creates a new SubjectLengthRule.
func NewSubjectLengthRule(maxLength int) *SubjectLengthRule {
	// Use default if not specified
	if maxLength <= 0 {
		maxLength = 100 // Default max length
	}

	return &SubjectLengthRule{
		maxLength: maxLength,
		errors:    make([]*domain.ValidationError, 0),
	}
}

// Name returns the rule name.
func (r *SubjectLengthRule) Name() string {
	return "SubjectLength"
}

// Validate validates the commit subject length.
func (r *SubjectLengthRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors
	r.errors = make([]*domain.ValidationError, 0)

	// Calculate subject length
	subjectLength := utf8.RuneCountInString(commit.Subject)

	// Check if subject is too long
	if subjectLength > r.maxLength {
		// Create error with context in one step
		context := map[string]string{
			"actual_length": strconv.Itoa(subjectLength),
			"max_length":    strconv.Itoa(r.maxLength),
		}

		err := errorx.NewErrorWithContext(r.Name(), errorx.ErrSubjectTooLong, context, subjectLength, r.maxLength)

		// Add to errors
		r.errors = append(r.errors, err)
	}

	return r.errors
}

// Result returns a concise result message.
func (r *SubjectLengthRule) Result() string {
	if len(r.errors) > 0 {
		return "Subject too long"
	}

	return "Subject length OK"
}

// VerboseResult returns a detailed result message.
func (r *SubjectLengthRule) VerboseResult() string {
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
func (r *SubjectLengthRule) Help() string {
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
	return errorx.FormatHelp(errorx.ErrSubjectTooLong, maxLength)
}

// Errors returns all validation errors.
func (r *SubjectLengthRule) Errors() []*domain.ValidationError {
	return r.errors
}
