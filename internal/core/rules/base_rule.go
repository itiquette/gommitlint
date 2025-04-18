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
func NewBaseRule(name string) *BaseRule {
	return &BaseRule{
		name:   name,
		errors: make([]errors.ValidationError, 0),
	}
}

// Name returns the rule's name.
func (r *BaseRule) Name() string {
	return r.name
}

// AddError adds a validation error to the rule's error collection.
func (r *BaseRule) AddError(err errors.ValidationError) {
	r.errors = append(r.errors, err)
}

// AddErrorWithCode creates and adds a new error with the given code and message.
func (r *BaseRule) AddErrorWithCode(code errors.ValidationErrorCode, message string) {
	err := errors.New(r.name, code, message)
	r.errors = append(r.errors, err)
}

// AddErrorWithFormatf creates and adds an error with formatted message.
func (r *BaseRule) AddErrorWithFormatf(code errors.ValidationErrorCode, format string, args ...interface{}) {
	err := errors.NewWithFormat(r.name, code, format, args...)
	r.errors = append(r.errors, err)
}

// AddErrorWithContext creates and adds an error with the given context.
func (r *BaseRule) AddErrorWithContext(code errors.ValidationErrorCode, message string, context map[string]string) {
	err := errors.New(r.name, code, message, errors.WithContextMap(context))
	r.errors = append(r.errors, err)
}

// AddErrorWithFormatAndContextf creates an error with formatted message and context.
func (r *BaseRule) AddErrorWithFormatAndContextf(code errors.ValidationErrorCode, context map[string]string, format string, args ...interface{}) {
	err := errors.New(
		r.name,
		code,
		fmt.Sprintf(format, args...),
		errors.WithContextMap(context),
	)
	r.errors = append(r.errors, err)
}

// AddErrorWithHelp creates an error with a help message.
func (r *BaseRule) AddErrorWithHelp(code errors.ValidationErrorCode, message string, help string) {
	err := errors.New(r.name, code, message, errors.WithHelp(help))
	r.errors = append(r.errors, err)
}

// AddAppError creates and adds a new error with just code and message (no context).
// This is a convenience method for the common case of simple errors.
func (r *BaseRule) AddAppError(code errors.ValidationErrorCode, message string) {
	err := errors.New(r.name, code, message)
	r.errors = append(r.errors, err)
}

// Errors returns all validation errors found by this rule.
func (r *BaseRule) Errors() []errors.ValidationError {
	return r.errors
}

// ClearErrors clears all recorded errors.
func (r *BaseRule) ClearErrors() {
	r.errors = make([]errors.ValidationError, 0)
}

// HasErrors returns true if the rule has found any errors.
func (r *BaseRule) HasErrors() bool {
	return len(r.errors) > 0
}

// ErrorCount returns the number of errors found by this rule.
func (r *BaseRule) ErrorCount() int {
	return len(r.errors)
}

// SetVerbosity sets additional information for verbose output.
func (r *BaseRule) SetVerbosity(verbosity string) {
	r.verbosity = verbosity
}

// Result returns a concise result message.
func (r *BaseRule) Result() string {
	if !r.hasRun {
		return "Rule has not been run"
	}

	if r.HasErrors() {
		return fmt.Sprintf("%s: Failed with %d error(s)", r.name, r.ErrorCount())
	}

	return r.name + ": Passed"
}

// VerboseResult returns a detailed result message with all errors and context.
func (r *BaseRule) VerboseResult() string {
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
func (r *BaseRule) Help() string {
	if !r.HasErrors() {
		return r.name + ": No errors to fix"
	}

	result := r.name + ": How to fix errors:\n"
	for i, err := range r.errors {
		result += fmt.Sprintf("  %d. %s (code: %s)\n", i+1, err.Message, err.Code)
	}

	return result
}

// MarkAsRun marks the rule as having been run.
func (r *BaseRule) MarkAsRun() {
	r.hasRun = true
}
