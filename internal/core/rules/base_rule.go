// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/errors"
)

// BaseRule provides common functionality for rule implementations.
// It handles error collection and standard result formatting to reduce
// duplication across rule implementations.
type BaseRule struct {
	name      string
	errors    []errors.ValidationError
	hasRun    bool
	verbosity string
}

// NewBaseRule creates a new BaseRule with the given name.
func NewBaseRule(name string) BaseRule {
	return BaseRule{
		name:   name,
		errors: make([]errors.ValidationError, 0),
	}
}

// Name returns the rule's name.
func (r BaseRule) Name() string {
	return r.name
}

// WithError adds a validation error and returns a new BaseRule.
func (r BaseRule) WithError(err errors.ValidationError) BaseRule {
	newErrors := make([]errors.ValidationError, len(r.errors), len(r.errors)+1)
	copy(newErrors, r.errors)

	result := r

	result.errors = append(newErrors, err)

	return result
}

// WithErrorWithCode creates and adds a new error with the given code and message,
// returning a new BaseRule.
func (r BaseRule) WithErrorWithCode(code errors.ValidationErrorCode, message string) BaseRule {
	err := errors.New(r.name, code, message)

	return r.WithError(err)
}

// WithErrorWithFormatf creates and adds an error with formatted message,
// returning a new BaseRule.
func (r BaseRule) WithErrorWithFormatf(code errors.ValidationErrorCode, format string, args ...interface{}) BaseRule {
	err := errors.NewWithFormat(r.name, code, format, args...)

	return r.WithError(err)
}

// WithErrorWithContext creates and adds an error with the given context,
// returning a new BaseRule.
func (r BaseRule) WithErrorWithContext(code errors.ValidationErrorCode, message string, context map[string]string) BaseRule {
	err := errors.New(r.name, code, message, errors.WithContextMap(context))

	return r.WithError(err)
}

// WithErrorWithFormatAndContextf creates an error with formatted message and context,
// returning a new BaseRule.
func (r BaseRule) WithErrorWithFormatAndContextf(code errors.ValidationErrorCode, context map[string]string, format string, args ...interface{}) BaseRule {
	err := errors.New(
		r.name,
		code,
		fmt.Sprintf(format, args...),
		errors.WithContextMap(context),
	)

	return r.WithError(err)
}

// WithErrorWithHelp creates an error with a help message,
// returning a new BaseRule.
func (r BaseRule) WithErrorWithHelp(code errors.ValidationErrorCode, message string, help string) BaseRule {
	err := errors.New(r.name, code, message, errors.WithHelp(help))

	return r.WithError(err)
}

// WithAppError creates and adds a new error with just code and message (no context),
// returning a new BaseRule. This is a convenience method for the common case of simple errors.
func (r BaseRule) WithAppError(code errors.ValidationErrorCode, message string) BaseRule {
	err := errors.New(r.name, code, message)

	return r.WithError(err)
}

// Errors returns all validation errors found by this rule.
func (r BaseRule) Errors() []errors.ValidationError {
	return r.errors
}

// WithClearedErrors returns a new BaseRule with all errors cleared.
func (r BaseRule) WithClearedErrors() BaseRule {
	result := r
	result.errors = make([]errors.ValidationError, 0)

	return result
}

// HasErrors returns true if the rule has found any errors.
func (r BaseRule) HasErrors() bool {
	return len(r.errors) > 0
}

// ErrorCount returns the number of errors found by this rule.
func (r BaseRule) ErrorCount() int {
	return len(r.errors)
}

// WithVerbosity returns a new BaseRule with the specified verbosity.
func (r BaseRule) WithVerbosity(verbosity string) BaseRule {
	result := r
	result.verbosity = verbosity

	return result
}

// Result returns a concise result message.
func (r BaseRule) Result() string {
	if !r.hasRun {
		return "Rule has not been run"
	}

	if r.HasErrors() {
		return fmt.Sprintf("%s: Failed with %d error(s)", r.name, r.ErrorCount())
	}

	return r.name + ": Passed"
}

// VerboseResult returns a detailed result message with all errors and context.
func (r BaseRule) VerboseResult() string {
	if !r.hasRun {
		return r.name + ": Rule has not been run"
	}

	if !r.HasErrors() {
		if r.verbosity != "" {
			return fmt.Sprintf("%s: Passed - %s", r.name, r.verbosity)
		}

		return r.name + ": Passed"
	}

	result := fmt.Sprintf("%s: Failed with %d error(s):\n", r.name, r.ErrorCount())
	for i, err := range r.errors {
		result += fmt.Sprintf("  %d. %s", i+1, err.Message)
		if len(err.Context) > 0 {
			result += " ["

			contextInfo := ""
			for k, v := range err.Context {
				if contextInfo != "" {
					contextInfo += ", "
				}

				contextInfo += fmt.Sprintf("%s: %s", k, v)
			}

			result += contextInfo + "]"
		}

		result += "\n"
	}

	return result
}

// Help returns guidance on how to fix rule violations.
func (r BaseRule) Help() string {
	if !r.HasErrors() {
		return r.name + ": No errors to fix"
	}

	result := r.name + ": How to fix errors:\n"
	for i, err := range r.errors {
		result += fmt.Sprintf("  %d. %s (code: %s)\n", i+1, err.Message, err.Code)
	}

	return result
}

// WithRun returns a new BaseRule marked as having been run.
func (r BaseRule) WithRun() BaseRule {
	result := r
	result.hasRun = true

	return result
}
