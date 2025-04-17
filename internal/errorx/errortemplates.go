// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errorx

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
)

// ValidationErrorCode represents specific error codes for signature validation.
type ValidationErrorCode string

// Signature validation error codes.
const (
	ErrCommitNil              ValidationErrorCode = "commit_nil"
	ErrNoKeyDir               ValidationErrorCode = "no_key_dir"
	ErrInvalidKeyDir          ValidationErrorCode = "invalid_key_dir"
	ErrMissingSignature       ValidationErrorCode = "missing_signature"
	ErrInvalidSignatureFormat ValidationErrorCode = "invalid_signature_format"
	ErrUnknownSigFormat       ValidationErrorCode = "unknown_signature_format"
	ErrKeyNotTrusted          ValidationErrorCode = "key_not_trusted"
	ErrWeakKey                ValidationErrorCode = "weak_key"
	ErrVerificationFailed     ValidationErrorCode = "verification_failed"
	ErrInvalidCommit          ValidationErrorCode = "invalid_commit"
)

// ErrorTemplate defines standard formats for validation errors.
type ErrorTemplate struct {
	// Code is the ValidationErrorCode for this error template
	Code domain.ValidationErrorCode

	// Format is the string format for the error message
	Format string

	// HelpFmt is the string format for help messages
	HelpFmt string
}

// Common templates for validation errors.
var (
	// Format errors.
	ErrInvalidFormat = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidFormat,
		Format:  "Invalid format: %s does not follow the required pattern",
		HelpFmt: "Ensure %s follows the pattern: %s",
	}

	// Subject errors.
	ErrSubjectTooLong = ErrorTemplate{
		Code:    domain.ValidationErrorTooLong,
		Format:  "Subject exceeds maximum length of %d characters (found %d)",
		HelpFmt: "Keep the subject line to %d characters or less",
	}

	ErrSubjectCase = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidCase,
		Format:  "Subject has invalid case: %s (expected: %s)",
		HelpFmt: "Use %s case for the subject line",
	}

	ErrSubjectSuffix = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidSuffix,
		Format:  "Subject ends with invalid suffix: %q",
		HelpFmt: "Avoid ending the subject with %q",
	}

	ErrMissingSubject = ErrorTemplate{
		Code:    domain.ValidationErrorMissingSubject,
		Format:  "Missing commit subject",
		HelpFmt: "Provide a subject line for the commit message",
	}

	// Body errors.
	ErrInvalidBody = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidBody,
		Format:  "Invalid commit body: %s",
		HelpFmt: "Format the commit body correctly: %s",
	}

	// Conventional commit errors.
	ErrInvalidType = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidType,
		Format:  "Invalid commit type: %q (allowed types: %s)",
		HelpFmt: "Use one of the allowed commit types: %s",
	}

	ErrInvalidScope = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidScope,
		Format:  "Invalid commit scope: %q (allowed scopes: %s)",
		HelpFmt: "Use one of the allowed commit scopes: %s",
	}

	ErrEmptyDescription = ErrorTemplate{
		Code:    domain.ValidationErrorEmptyDescription,
		Format:  "Missing commit description",
		HelpFmt: "Add a description after the commit type",
	}

	// Imperative errors.
	ErrNonImperative = ErrorTemplate{
		Code:    domain.ValidationErrorNonImperative,
		Format:  "Subject does not use imperative mood: %q",
		HelpFmt: "Use imperative mood in the subject line (e.g., 'Add' instead of 'Added')",
	}

	ErrNonVerb = ErrorTemplate{
		Code:    domain.ValidationErrorNonVerb,
		Format:  "Subject does not start with a verb: %q",
		HelpFmt: "Start the subject with a verb in imperative form",
	}

	ErrPastTense = ErrorTemplate{
		Code:    domain.ValidationErrorPastTense,
		Format:  "Subject uses past tense verb: %q",
		HelpFmt: "Use present tense verbs in commit messages (e.g., 'Add' not 'Added')",
	}

	ErrGerund = ErrorTemplate{
		Code:    domain.ValidationErrorGerund,
		Format:  "Subject uses gerund verb form: %q",
		HelpFmt: "Avoid -ing verb forms in commit messages (e.g., 'Add' not 'Adding')",
	}

	ErrThirdPerson = ErrorTemplate{
		Code:    domain.ValidationErrorThirdPerson,
		Format:  "Subject uses third person verb form: %q",
		HelpFmt: "Avoid third-person singular verbs (e.g., 'Add' not 'Adds')",
	}

	// Jira errors.
	ErrMissingJira = ErrorTemplate{
		Code:    domain.ValidationErrorMissingJira,
		Format:  "no Jira issue key found in the commit message",
		HelpFmt: "Include a JIRA issue reference (e.g., PROJ-123) in the commit message",
	}

	// Signature errors.
	MissingSignatureTemplate = ErrorTemplate{
		Code:    domain.ValidationErrorMissingSignature,
		Format:  "Missing commit signature",
		HelpFmt: "Sign your commit using --gpg-sign or --sign",
	}

	InvalidSignatureTemplate = ErrorTemplate{
		Code:    domain.ValidationErrorInvalidSignature,
		Format:  "Invalid signature: %s",
		HelpFmt: "Fix the signature issue: %s",
	}

	DisallowedSigTypeTemplate = ErrorTemplate{
		Code:    domain.ValidationErrorDisallowedSigType,
		Format:  "Disallowed signature type: %s (allowed: %s)",
		HelpFmt: "Use one of the allowed signature types: %s",
	}

	// Signoff errors.
	ErrMissingSignoff = ErrorTemplate{
		Code:    domain.ValidationErrorMissingSignoff,
		Format:  "Missing Signed-off-by line",
		HelpFmt: "Add a Signed-off-by line using git commit --signoff",
	}

	// Spelling errors.
	ErrSpelling = ErrorTemplate{
		Code:    domain.ValidationErrorSpelling,
		Format:  "Spelling error detected: %q in %s",
		HelpFmt: "Correct the spelling of %q in the commit %s",
	}

	// Commits ahead errors.
	ErrTooManyCommits = ErrorTemplate{
		Code:    domain.ValidationErrorTooManyCommits,
		Format:  "Too many commits ahead of base branch: %d (maximum: %d)",
		HelpFmt: "Reduce the number of commits to %d or less (consider squashing or rebasing)",
	}
)

// FormatError creates a consistently formatted error message.
func FormatError(template ErrorTemplate, args ...interface{}) string {
	return fmt.Sprintf(template.Format, args...)
}

// FormatHelp creates a consistently formatted help message.
func FormatHelp(template ErrorTemplate, args ...interface{}) string {
	return fmt.Sprintf(template.HelpFmt, args...)
}

// NewValidationError creates a new ValidationError using a template.
func NewValidationError(ruleName string, template ErrorTemplate, args ...interface{}) *domain.ValidationError {
	return domain.NewStandardValidationError(
		ruleName,
		template.Code,
		FormatError(template, args...),
	)
}

func NewSignatureValidationError(ruleName string, code ValidationErrorCode, message string) *domain.ValidationError {
	return domain.NewValidationError(
		ruleName,
		string(code),
		message,
	)
}

// WithContext adds context to a ValidationError and returns it.
func WithContext(err *domain.ValidationError, context map[string]string) *domain.ValidationError {
	for key, value := range context {
		err = err.WithContext(key, value)
	}

	return err
}

// NewErrorWithContext creates a validation error with context in one step.
// This simplifies the common pattern of creating an error and adding context.
func NewErrorWithContext(ruleName string, template ErrorTemplate, contextMap map[string]string, args ...interface{}) *domain.ValidationError {
	err := NewValidationError(ruleName, template, args...)

	return WithContext(err, contextMap)
}

// NewSignatureErrorWithContext creates a signature validation error with context in one step.
func NewSignatureErrorWithContext(ruleName string, code ValidationErrorCode, message string, contextMap map[string]string) *domain.ValidationError {
	err := NewSignatureValidationError(ruleName, code, message)

	return WithContext(err, contextMap)
}
