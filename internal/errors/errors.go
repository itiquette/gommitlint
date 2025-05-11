// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package errors provides a consolidated error handling system for gommitlint.
// It defines a structured error type and a factory function approach for creating errors.
package errors

import (
	"strconv"
)

// ValidationErrorCode represents standardized error codes for validation errors.
type ValidationErrorCode string

func (v ValidationErrorCode) String() string {
	return string(v)
}

// Define all error codes here in a grouped structure for better organization.
const (
	// Format errors.
	ErrInvalidFormat ValidationErrorCode = "invalid_format"
	ErrSpacing       ValidationErrorCode = "spacing_error"

	// Subject errors.
	ErrSubjectTooLong ValidationErrorCode = "subject_too_long"
	ErrSubjectLength  ValidationErrorCode = "subject_length"
	ErrSubjectCase    ValidationErrorCode = "invalid_case"
	ErrSubjectSuffix  ValidationErrorCode = "invalid_suffix"
	ErrMissingSubject ValidationErrorCode = "missing_subject"
	ErrNoFirstWord    ValidationErrorCode = "no_first_word"
	ErrEmptyMessage   ValidationErrorCode = "empty_message"

	// Body errors.
	ErrInvalidBody ValidationErrorCode = "invalid_body"

	// Conventional commit errors.
	ErrInvalidType        ValidationErrorCode = "invalid_type"
	ErrInvalidScope       ValidationErrorCode = "invalid_scope"
	ErrEmptyDescription   ValidationErrorCode = "empty_description"
	ErrDescriptionTooLong ValidationErrorCode = "description_too_long"

	// Jira errors.
	ErrMissingJira   ValidationErrorCode = "missing_jira"
	ErrMisplacedJira ValidationErrorCode = "misplaced_jira"

	// Imperative mood errors.
	ErrNonImperative ValidationErrorCode = "non_imperative"
	ErrNonVerb       ValidationErrorCode = "non_verb"
	ErrPastTense     ValidationErrorCode = "past_tense"
	ErrGerund        ValidationErrorCode = "gerund"
	ErrThirdPerson   ValidationErrorCode = "third_person"

	// Signature errors.
	ErrCommitNil              ValidationErrorCode = "commit_nil"
	ErrNoKeyDir               ValidationErrorCode = "no_key_dir"
	ErrInvalidKeyDir          ValidationErrorCode = "invalid_key_dir"
	ErrMissingSignature       ValidationErrorCode = "missing_signature"
	ErrInvalidSignature       ValidationErrorCode = "invalid_signature"
	ErrInvalidSignatureFormat ValidationErrorCode = "invalid_signature_format"
	ErrUnknownSigFormat       ValidationErrorCode = "unknown_signature_format"
	ErrKeyNotTrusted          ValidationErrorCode = "key_not_trusted"
	ErrWeakKey                ValidationErrorCode = "weak_key"
	ErrVerificationFailed     ValidationErrorCode = "verification_failed"
	ErrDisallowedSigType      ValidationErrorCode = "disallowed_signature_type"
	ErrIncompleteGPGSig       ValidationErrorCode = "incomplete_gpg_signature"
	ErrIncompleteSSHSig       ValidationErrorCode = "incomplete_ssh_signature"
	ErrInvalidGPGFormat       ValidationErrorCode = "invalid_gpg_format"
	ErrInvalidSSHFormat       ValidationErrorCode = "invalid_ssh_format"
	ErrInvalidCommit          ValidationErrorCode = "invalid_commit"

	// Signoff errors.
	ErrMissingSignoff ValidationErrorCode = "missing_signoff"

	// Spelling errors.
	ErrSpelling       ValidationErrorCode = "spelling_error"
	ErrMisspelledWord ValidationErrorCode = "misspelled_word"

	// Commits ahead errors.
	ErrTooManyCommits ValidationErrorCode = "too_many_commits"

	// Git operation errors.
	ErrInvalidRepo        ValidationErrorCode = "invalid_repo"
	ErrInvalidConfig      ValidationErrorCode = "invalid_config"
	ErrCancelled          ValidationErrorCode = "operation_cancelled"
	ErrGitOperationFailed ValidationErrorCode = "git_operation_failed"
	ErrContextCancelled   ValidationErrorCode = "context_cancelled"
	ErrCommitNotFound     ValidationErrorCode = "commit_not_found"
	ErrRangeNotFound      ValidationErrorCode = "range_not_found"

	// Reference errors.
	ErrInvalidReference ValidationErrorCode = "invalid_reference"
	ErrMissingReference ValidationErrorCode = "missing_reference"

	// Max length errors.
	ErrMaxLengthExceeded ValidationErrorCode = "max_length_exceeded"

	// Generic errors.
	ErrUnknown ValidationErrorCode = "unknown_error"
)

// Note: Error templates have been removed as they are now handled by specialized error helpers
// in enhanced_errors.go like FormatError, LengthError, etc. These provide rich context
// and standardized formatting.

// ValidationError represents an error detected during validation.
// This is the standard error type for the entire application.
type ValidationError struct {
	// Rule is the name of the rule that produced this error.
	Rule string

	// Code is the error code, which can be used for programmatic handling.
	Code string

	// Message is a human-readable error message.
	Message string

	// Context contains additional information about the error.
	Context map[string]string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.Message
}

// WithContext adds context information to a ValidationError.
// Note: This implementation was updated for immutability with functional programming principles.
func (e ValidationError) WithContext(key, value string) ValidationError {
	result := e

	// Create a new context map if needed
	if result.Context == nil {
		result.Context = make(map[string]string)
	} else {
		// Copy the existing map
		newContext := make(map[string]string, len(result.Context)+1)
		for k, v := range result.Context {
			newContext[k] = v
		}

		result.Context = newContext
	}

	result.Context[key] = value

	return result
}

// WithIntContext adds an integer context value.
func (e ValidationError) WithIntContext(key string, value int) ValidationError {
	return e.WithContext(key, strconv.Itoa(value))
}

// ErrorOption defines a function that can modify a ValidationError.
// Note: This is maintained for backward compatibility but new code should use
// the functional chaining approach with methods like WithContext().
type ErrorOption func(*ValidationError)

// CreateBasicError creates a new ValidationError.
func CreateBasicError(rule string, code ValidationErrorCode, message string) ValidationError {
	return ValidationError{
		Rule:    rule,
		Code:    string(code),
		Message: message,
		Context: make(map[string]string),
	}
}
