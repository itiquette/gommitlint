// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"fmt"
)

// ValidationErrorCode represents standardized error codes for validation errors.
// These codes provide a stable interface for programmatic error handling across
// the application. They are organized by category (format, subject, body, etc.)
// and follow a consistent naming pattern using snake_case.
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
	ErrInvalidBody  ValidationErrorCode = "invalid_body"
	ErrMissingBody  ValidationErrorCode = "missing_body"
	ErrBodyTooShort ValidationErrorCode = "body_too_short"

	// Conventional commit errors.
	ErrInvalidType               ValidationErrorCode = "invalid_type"
	ErrInvalidScope              ValidationErrorCode = "invalid_scope"
	ErrEmptyDescription          ValidationErrorCode = "empty_description"
	ErrDescriptionTooLong        ValidationErrorCode = "description_too_long"
	ErrInvalidConventionalFormat ValidationErrorCode = "invalid_conventional_format"
	ErrInvalidConventionalType   ValidationErrorCode = "invalid_conventional_type"
	ErrMissingConventionalScope  ValidationErrorCode = "missing_conventional_scope"
	ErrInvalidConventionalScope  ValidationErrorCode = "invalid_conventional_scope"
	ErrConventionalDescTooLong   ValidationErrorCode = "conventional_desc_too_long"

	// Jira errors.
	ErrMissingJira    ValidationErrorCode = "missing_jira"
	ErrMisplacedJira  ValidationErrorCode = "misplaced_jira"
	ErrInvalidProject ValidationErrorCode = "invalid_project"

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
	ErrSpelling         ValidationErrorCode = "spelling_error"
	ErrMisspelledWord   ValidationErrorCode = "misspelled_word"
	ErrSpellCheckFailed ValidationErrorCode = "spell_check_failed"

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

// ValidationError represents an error detected during validation.
// This is the standard error type for the entire application.
type ValidationError struct {
	// Rule is the name of the rule that produced this error.
	Rule string

	// Code is the error code, which can be used for programmatic handling.
	Code string

	// Message is a human-readable error message.
	Message string

	// Help is an optional help text that provides guidance on how to fix the error.
	Help string

	// Context contains additional information about the error.
	Context map[string]string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.Message
}

// WithContextMap adds context values to a ValidationError.
func (e ValidationError) WithContextMap(ctx map[string]string) ValidationError {
	result := e

	// Copy existing context plus new values
	newContext := make(map[string]string, len(result.Context)+len(ctx))
	for k, v := range result.Context {
		newContext[k] = v
	}

	for k, v := range ctx {
		newContext[k] = v
	}

	result.Context = newContext

	return result
}

// WithHelp adds help text to a ValidationError.
func (e ValidationError) WithHelp(help string) ValidationError {
	result := e
	result.Help = help

	return result
}

// WithUserMessage updates the error message with a user-friendly version.
// This allows providing clearer, more actionable messages while preserving the original technical message.
func (e ValidationError) WithUserMessage(format string, args ...interface{}) ValidationError {
	result := e
	result.Message = fmt.Sprintf(format, args...)

	return result
}

// New creates a new ValidationError.
func New(rule string, code ValidationErrorCode, message string) ValidationError {
	return ValidationError{
		Rule:    rule,
		Code:    string(code),
		Message: message,
		Help:    "",
		Context: make(map[string]string),
	}
}
