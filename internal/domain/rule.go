// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import "context"

// SeverityLevel represents the severity of a rule violation.
type SeverityLevel string

const (
	// SeverityError indicates a rule violation that should fail validation.
	SeverityError SeverityLevel = "error"

	// SeverityWarning indicates a rule violation that should warn but not fail validation.
	SeverityWarning SeverityLevel = "warning"

	// SeverityInfo indicates informational output that isn't a violation.
	SeverityInfo SeverityLevel = "info"
)

// RuleMetadata provides information about a validation rule.
type RuleMetadata struct {
	// ID is the unique identifier for the rule.
	ID string

	// Name is the human-readable name of the rule.
	Name string

	// Description is a detailed description of what the rule validates.
	Description string

	// Severity indicates how severe violations of this rule are.
	Severity SeverityLevel
}

// ValidationErrorCode represents standardized error codes for validation errors.
type ValidationErrorCode string

// Standard validation error codes.
const (
	ValidationErrorInvalidFormat      ValidationErrorCode = "invalid_format"
	ValidationErrorSpacing            ValidationErrorCode = "spacing_error"
	ValidationErrorInvalidType        ValidationErrorCode = "invalid_type"
	ValidationErrorInvalidScope       ValidationErrorCode = "invalid_scope"
	ValidationErrorEmptyDescription   ValidationErrorCode = "empty_description"
	ValidationErrorDescriptionTooLong ValidationErrorCode = "description_too_long"
	ValidationErrorTooLong            ValidationErrorCode = "description_too_long" // Standard alias
	ValidationErrorInvalidCase        ValidationErrorCode = "invalid_case"
	ValidationErrorInvalidSuffix      ValidationErrorCode = "invalid_suffix"
	ValidationErrorNonImperative      ValidationErrorCode = "non_imperative"
	ValidationErrorMissingJira        ValidationErrorCode = "missing_jira"
	ValidationErrorInvalidSignature   ValidationErrorCode = "invalid_signature"
	ValidationErrorMissingSignature   ValidationErrorCode = "missing_signature"
	ValidationErrorDisallowedSigType  ValidationErrorCode = "disallowed_signature_type"
	ValidationErrorUnknownSigFormat   ValidationErrorCode = "unknown_signature_format"
	ValidationErrorIncompleteGPGSig   ValidationErrorCode = "incomplete_gpg_signature"
	ValidationErrorIncompleteSSHSig   ValidationErrorCode = "incomplete_ssh_signature"
	ValidationErrorInvalidGPGFormat   ValidationErrorCode = "invalid_gpg_format"
	ValidationErrorInvalidSSHFormat   ValidationErrorCode = "invalid_ssh_format"
	ValidationErrorMissingSignoff     ValidationErrorCode = "missing_signoff"
	ValidationErrorSpelling           ValidationErrorCode = "spelling_error"
	ValidationErrorInvalidBody        ValidationErrorCode = "invalid_body"
	ValidationErrorTooManyCommits     ValidationErrorCode = "too_many_commits"
	ValidationErrorEmptyMessage       ValidationErrorCode = "empty_message"
	ValidationErrorNonVerb            ValidationErrorCode = "non_verb"
	ValidationErrorPastTense          ValidationErrorCode = "past_tense"
	ValidationErrorGerund             ValidationErrorCode = "gerund"
	ValidationErrorThirdPerson        ValidationErrorCode = "third_person"
	ValidationErrorMissingSubject     ValidationErrorCode = "missing_subject"
	ValidationErrorNoFirstWord        ValidationErrorCode = "no_first_word"
	ValidationErrorUnknown            ValidationErrorCode = "unknown_error"
)

// ValidationError represents an error detected during rule validation.
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

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return e.Message
}

// WithContext adds context information to a ValidationError.
func (e *ValidationError) WithContext(key, value string) *ValidationError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}

	e.Context[key] = value

	return e
}

// NewValidationError creates a new ValidationError.
func NewValidationError(rule, code, message string) *ValidationError {
	return &ValidationError{
		Rule:    rule,
		Code:    code,
		Message: message,
		Context: make(map[string]string),
	}
}

// NewStandardValidationError creates a validation error with a standardized error code.
func NewStandardValidationError(rule string, code ValidationErrorCode, message string) *ValidationError {
	return NewValidationError(rule, string(code), message)
}

// NewValidationErrorWithContext creates a validation error with context in one step.
func NewValidationErrorWithContext(rule, code, message string, context map[string]string) *ValidationError {
	err := NewValidationError(rule, code, message)

	for key, value := range context {
		err = err.WithContext(key, value)
	}

	return err
}

// Rule defines the interface for all validation rules.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation against a commit and returns any errors.
	Validate(commit *CommitInfo) []*ValidationError

	// Result returns a concise result message.
	Result() string

	// VerboseResult returns a detailed result message.
	VerboseResult() string

	// Help returns guidance on how to fix rule violations.
	Help() string

	// Errors returns all validation errors found by this rule.
	Errors() []*ValidationError
}

// ContextualRule extends Rule with context-aware methods.
type ContextualRule interface {
	Rule

	// ValidateWithContext performs validation with context.
	ValidateWithContext(ctx context.Context, commit *CommitInfo) []*ValidationError
}

// RuleProvider defines an interface for retrieving validation rules.
type RuleProvider interface {
	// GetRules returns all available validation rules.
	GetRules() []Rule

	// GetActiveRules returns all active validation rules based on configuration.
	GetActiveRules() []Rule
}
