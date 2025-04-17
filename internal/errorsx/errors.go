// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package errorsx provides enhanced error handling for the application.
package errorsx

import (
	"errors"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// ErrorCode defines the type of error that occurred.
type ErrorCode string

const (
	// ErrUnknown is used when the error type cannot be determined.
	ErrUnknown ErrorCode = "unknown"

	// ErrValidation indicates a validation error.
	ErrValidation ErrorCode = "validation"

	// ErrGit indicates an error with Git operations.
	ErrGit ErrorCode = "git"

	// ErrConfiguration indicates a configuration error.
	ErrConfiguration ErrorCode = "configuration"

	// ErrInput indicates an error with user input.
	ErrInput ErrorCode = "input"

	// ErrApplication indicates a general application error.
	ErrApplication ErrorCode = "application"

	// Rule-specific error codes.
	ErrSubjectLength      ErrorCode = "subject_too_long"
	ErrSubjectCase        ErrorCode = "invalid_subject_case"
	ErrSubjectSuffix      ErrorCode = "invalid_subject_suffix"
	ErrConventionalFormat ErrorCode = "invalid_conventional_format"
	ErrInvalidType        ErrorCode = "invalid_type"
	ErrInvalidScope       ErrorCode = "invalid_scope"
	ErrEmptyDescription   ErrorCode = "empty_description"
	ErrSpacingError       ErrorCode = "spacing_error"
	ErrSignatureError     ErrorCode = "signature_error"
	ErrSignoffMissing     ErrorCode = "signoff_missing"
	ErrImperativeVerb     ErrorCode = "non_imperative_verb"
	ErrJiraReference      ErrorCode = "missing_jira_reference"
	ErrSpellCheck         ErrorCode = "spelling_error"
	ErrCommitBody         ErrorCode = "invalid_commit_body"
	ErrCommitsAhead       ErrorCode = "commits_ahead_violation"
)

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
	Code     ErrorCode
	Message  string
	ExitCode int
	Context  map[string]string
}

// Error returns the error message.
func (e *ApplicationError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
}

// Unwrap implements error wrapping.
func (e *ApplicationError) Unwrap() error {
	return e.Cause
}

// ExitCodeValue returns the appropriate exit code for the error.
func (e *ApplicationError) ExitCodeValue() int {
	return e.ExitCode
}

// WithContext adds context information to the error.
func (e *ApplicationError) WithContext(key, value string) *ApplicationError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}

	e.Context[key] = value

	return e
}

// Is enables error comparison.
func (e *ApplicationError) Is(target error) bool {
	var appErr *ApplicationError
	if !errors.As(target, &appErr) {
		return false
	}

	return e.Code == appErr.Code
}

// NewError creates a new application error.
func NewError(code ErrorCode, message string, exitCode int, cause error) *ApplicationError {
	return &ApplicationError{
		Cause:    cause,
		Code:     code,
		Message:  message,
		ExitCode: exitCode,
		Context:  make(map[string]string),
	}
}

// NewApplicationError creates a general application error.
func NewApplicationError(message string, cause error) *ApplicationError {
	return NewError(ErrApplication, message, 1, cause)
}

// NewValidationError creates a validation error.
func NewValidationError(message string, cause error) *ApplicationError {
	return NewError(ErrValidation, message, 2, cause)
}

// NewGitError creates a Git error.
func NewGitError(message string, cause error) *ApplicationError {
	return NewError(ErrGit, message, 3, cause)
}

// NewConfigError creates a configuration error.
func NewConfigError(message string, cause error) *ApplicationError {
	return NewError(ErrConfiguration, message, 4, cause)
}

// NewInputError creates an input error.
func NewInputError(message string, cause error) *ApplicationError {
	return NewError(ErrInput, message, 5, cause)
}

// NewRuleValidationError creates a validation error with a specific rule error code.
func NewRuleValidationError(ruleCode ErrorCode, ruleName, message string, cause error) *ApplicationError {
	return NewError(ruleCode, message, 2, cause).
		WithContext("rule", ruleName)
}

// ConvertDomainValidationError converts a domain.ValidationError to an ApplicationError.
func ConvertDomainValidationError(err *domain.ValidationError) *ApplicationError {
	// Map domain error code to application error code
	errorCode := ErrValidation

	switch err.Code {
	case "invalid_format", "spacing_error":
		errorCode = ErrConventionalFormat
	case "invalid_type":
		errorCode = ErrInvalidType
	case "invalid_scope":
		errorCode = ErrInvalidScope
	case "empty_description":
		errorCode = ErrEmptyDescription
	case "description_too_long":
		errorCode = ErrSubjectLength
	case "invalid_case":
		errorCode = ErrSubjectCase
	case "invalid_suffix":
		errorCode = ErrSubjectSuffix
	case "non_imperative":
		errorCode = ErrImperativeVerb
	case "missing_jira":
		errorCode = ErrJiraReference
	case "invalid_signature":
		errorCode = ErrSignatureError
	case "missing_signoff":
		errorCode = ErrSignoffMissing
	case "spelling_error":
		errorCode = ErrSpellCheck
	case "invalid_body":
		errorCode = ErrCommitBody
	case "too_many_commits":
		errorCode = ErrCommitsAhead
	}

	// Create application error
	appErr := NewError(errorCode, err.Message, 2, nil).
		WithContext("rule", err.Rule)

	// Copy context from domain error
	for k, v := range err.Context {
		_ = appErr.WithContext(k, v)
	}

	return appErr
}

// NewSubjectLengthError creates a subject length validation error.
func NewSubjectLengthError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrSubjectLength, ruleName, message, cause)
}

// NewSubjectCaseError creates a subject case validation error.
func NewSubjectCaseError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrSubjectCase, ruleName, message, cause)
}

// NewSubjectSuffixError creates a subject suffix validation error.
func NewSubjectSuffixError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrSubjectSuffix, ruleName, message, cause)
}

// NewConventionalFormatError creates a conventional format validation error.
func NewConventionalFormatError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrConventionalFormat, ruleName, message, cause)
}

// NewImperativeVerbError creates an imperative verb validation error.
func NewImperativeVerbError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrImperativeVerb, ruleName, message, cause)
}

// NewJiraReferenceError creates a JIRA reference validation error.
func NewJiraReferenceError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrJiraReference, ruleName, message, cause)
}

// NewSignatureError creates a signature validation error.
func NewSignatureError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrSignatureError, ruleName, message, cause)
}

// NewSignoffError creates a signoff validation error.
func NewSignoffError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrSignoffMissing, ruleName, message, cause)
}

// NewSpellCheckError creates a spell check validation error.
func NewSpellCheckError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrSpellCheck, ruleName, message, cause)
}

// NewCommitBodyError creates a commit body validation error.
func NewCommitBodyError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrCommitBody, ruleName, message, cause)
}

// NewCommitsAheadError creates a commits ahead validation error.
func NewCommitsAheadError(ruleName, message string, cause error) *ApplicationError {
	return NewRuleValidationError(ErrCommitsAhead, ruleName, message, cause)
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrValidation
	}

	return false
}

// IsGitError checks if an error is a Git error.
func IsGitError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrGit
	}

	return false
}

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrConfiguration
	}

	return false
}

// IsInputError checks if an error is an input error.
func IsInputError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrInput
	}

	return false
}
