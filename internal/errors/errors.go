// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package errors provides a consolidated error handling system for gommitlint.
// It defines a unified error type and a factory function approach for creating errors.
package errors

import (
	"errors"
	"fmt"
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
	ErrMissingJira ValidationErrorCode = "missing_jira"

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
	ErrSpelling ValidationErrorCode = "spelling_error"

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

	// Generic errors.
	ErrUnknown ValidationErrorCode = "unknown_error"
)

// Error templates for common error messages.
var errorTemplates = map[ValidationErrorCode]struct {
	Format  string
	HelpFmt string
}{
	// Format errors
	ErrInvalidFormat: {
		Format:  "Invalid format: %s does not follow the required pattern",
		HelpFmt: "Ensure %s follows the pattern: %s",
	},

	// Subject errors
	ErrSubjectTooLong: {
		Format:  "Subject exceeds maximum length of %d characters (found %d)",
		HelpFmt: "Keep the subject line to %d characters or less",
	},
	ErrSubjectCase: {
		Format:  "Subject has invalid case: %s (expected: %s)",
		HelpFmt: "Use %s case for the subject line",
	},
	ErrSubjectSuffix: {
		Format:  "Subject ends with invalid suffix: %q",
		HelpFmt: "Avoid ending the subject with %q",
	},
	ErrMissingSubject: {
		Format:  "Missing commit subject",
		HelpFmt: "Provide a subject line for the commit message",
	},

	// Body errors
	ErrInvalidBody: {
		Format:  "Invalid commit body: %s",
		HelpFmt: "Format the commit body correctly: %s",
	},

	// Conventional commit errors
	ErrInvalidType: {
		Format:  "Invalid commit type: %q (allowed types: %s)",
		HelpFmt: "Use one of the allowed commit types: %s",
	},
	ErrInvalidScope: {
		Format:  "Invalid commit scope: %q (allowed scopes: %s)",
		HelpFmt: "Use one of the allowed commit scopes: %s",
	},
	ErrEmptyDescription: {
		Format:  "Missing commit description",
		HelpFmt: "Add a description after the commit type",
	},

	// Imperative errors
	ErrNonImperative: {
		Format:  "Subject does not use imperative mood: %q",
		HelpFmt: "Use imperative mood in the subject line (e.g., 'Add' instead of 'Added')",
	},
	ErrNonVerb: {
		Format:  "Subject does not start with a verb: %q",
		HelpFmt: "Start the subject with a verb in imperative form",
	},
	ErrPastTense: {
		Format:  "Subject uses past tense verb: %q",
		HelpFmt: "Use present tense verbs in commit messages (e.g., 'Add' not 'Added')",
	},
	ErrGerund: {
		Format:  "Subject uses gerund verb form: %q",
		HelpFmt: "Avoid -ing verb forms in commit messages (e.g., 'Add' not 'Adding')",
	},
	ErrThirdPerson: {
		Format:  "Subject uses third person verb form: %q",
		HelpFmt: "Avoid third-person singular verbs (e.g., 'Add' not 'Adds')",
	},

	// Jira errors
	ErrMissingJira: {
		Format:  "No Jira issue key found in the commit message",
		HelpFmt: "Include a JIRA issue reference (e.g., PROJ-123) in the commit message",
	},

	// Signature errors
	ErrMissingSignature: {
		Format:  "Missing commit signature",
		HelpFmt: "Sign your commit using --gpg-sign or --sign",
	},
	ErrInvalidSignature: {
		Format:  "Invalid signature: %s",
		HelpFmt: "Fix the signature issue: %s",
	},
	ErrDisallowedSigType: {
		Format:  "Disallowed signature type: %s (allowed: %s)",
		HelpFmt: "Use one of the allowed signature types: %s",
	},

	// Signoff errors
	ErrMissingSignoff: {
		Format:  "Missing Signed-off-by line",
		HelpFmt: "Add a Signed-off-by line using git commit --signoff",
	},

	// Spelling errors
	ErrSpelling: {
		Format:  "Spelling error detected: %q in %s",
		HelpFmt: "Correct the spelling of %q in the commit %s",
	},

	// Commits ahead errors
	ErrTooManyCommits: {
		Format:  "Too many commits ahead of base branch: %d (maximum: %d)",
		HelpFmt: "Reduce the number of commits to %d or less (consider squashing or rebasing)",
	},
}

// ValidationError represents an error detected during validation.
// This is the single, unified error type for the entire application.
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
type ErrorOption func(*ValidationError)

// WithContextMap adds an entire context map to the error.
func WithContextMap(ctx map[string]string) ErrorOption {
	return func(err *ValidationError) {
		if err.Context == nil {
			err.Context = make(map[string]string)
		}

		for k, v := range ctx {
			err.Context[k] = v
		}
	}
}

// WithHelp adds a help message to the error context.
func WithHelp(help string) ErrorOption {
	return func(err *ValidationError) {
		_ = err.WithContext("help", help)
	}
}

// WithFormattedMessage formats the error message with arguments.
func WithFormattedMessage(format string, args ...interface{}) ErrorOption {
	return func(err *ValidationError) {
		err.Message = fmt.Sprintf(format, args...)
	}
}

// New creates a new ValidationError with options.
// This is the main factory function for creating errors throughout the application.
func New(rule string, code ValidationErrorCode, message string, opts ...ErrorOption) ValidationError {
	err := ValidationError{
		Rule:    rule,
		Code:    string(code),
		Message: message,
		Context: make(map[string]string),
	}

	// Apply all options
	for _, opt := range opts {
		opt(&err)
	}

	return err
}

// NewWithFormat creates an error with a formatted message.
func NewWithFormat(rule string, code ValidationErrorCode, format string, args ...interface{}) ValidationError {
	return New(rule, code, fmt.Sprintf(format, args...))
}

// FormatError creates a consistently formatted error message using templates.
func FormatError(code ValidationErrorCode, args ...interface{}) string {
	template, exists := errorTemplates[code]
	if !exists {
		return fmt.Sprintf("Error: %s", code)
	}

	return fmt.Sprintf(template.Format, args...)
}

// FormatHelp creates a consistently formatted help message using templates.
func FormatHelp(code ValidationErrorCode, args ...interface{}) string {
	template, exists := errorTemplates[code]
	if !exists {
		return "Fix the error and try again"
	}

	return fmt.Sprintf(template.HelpFmt, args...)
}

// NewErrorWithContext creates a validation error with context in one step.
func NewErrorWithContext(rule string, code ValidationErrorCode, contextMap map[string]string, args ...interface{}) ValidationError {
	err := New(rule, code, FormatError(code, args...))

	// Add context if any
	for key, value := range contextMap {
		err = err.WithContext(key, value)
	}

	return err
}

// ==========================================
// Application error handling (from errorsx)
// ==========================================

// ApplicationErrorCode defines the type of application error that occurred.
type ApplicationErrorCode string

const (
	// AppErrUnknown is used when the error type cannot be determined.
	AppErrUnknown ApplicationErrorCode = "unknown"

	// AppErrValidation indicates a validation error.
	AppErrValidation ApplicationErrorCode = "validation"

	// AppErrGit indicates an error with Git operations.
	AppErrGit ApplicationErrorCode = "git"

	// AppErrConfiguration indicates a configuration error.
	AppErrConfiguration ApplicationErrorCode = "configuration"

	// AppErrInput indicates an error with user input.
	AppErrInput ApplicationErrorCode = "input"

	// AppErrApplication indicates a general application error.
	AppErrApplication ApplicationErrorCode = "application"
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
	Code     ApplicationErrorCode
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

// WithAppContext adds context information to the application error.
func (e *ApplicationError) WithAppContext(key, value string) *ApplicationError {
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

// NewAppError creates a new application error.
func NewAppError(code ApplicationErrorCode, message string, exitCode int, cause error) *ApplicationError {
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
	return NewAppError(AppErrApplication, message, 1, cause)
}

// NewValidationAppError creates a validation error.
func NewValidationAppError(message string, cause error) *ApplicationError {
	return NewAppError(AppErrValidation, message, 2, cause)
}

// NewGitError creates a Git error.
func NewGitError(message string, cause error) *ApplicationError {
	return NewAppError(AppErrGit, message, 3, cause)
}

// NewConfigError creates a configuration error.
func NewConfigError(message string, cause error) *ApplicationError {
	return NewAppError(AppErrConfiguration, message, 4, cause)
}

// NewInputError creates an input error.
func NewInputError(message string, cause error) *ApplicationError {
	return NewAppError(AppErrInput, message, 5, cause)
}

// IsValidationAppError checks if an error is a validation error.
func IsValidationAppError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == AppErrValidation
	}

	return false
}

// IsGitError checks if an error is a Git error.
func IsGitError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == AppErrGit
	}

	return false
}

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == AppErrConfiguration
	}

	return false
}

// IsInputError checks if an error is an input error.
func IsInputError(err error) bool {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Code == AppErrInput
	}

	return false
}
