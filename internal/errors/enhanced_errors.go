// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// EnhanceValidationError adds help text to an existing ValidationError.
func EnhanceValidationError(err ValidationError, help string) ValidationError {
	// Create a copy of the existing error
	result := err

	// Add help to the context
	result = result.WithContext("help", help)

	return result
}

// WithCommitSHA adds a commit SHA to the error context.
func (e ValidationError) WithCommitSHA(commitSHA string) ValidationError {
	return e.WithContext("commit_sha", commitSHA)
}

// GetHelp retrieves the help message from the error context.
func (e ValidationError) GetHelp() string {
	if help, ok := e.Context["help"]; ok {
		return help
	}

	return ""
}

// FormatAsText formats the error as text for CLI output.
func (e ValidationError) FormatAsText(verbose bool) string {
	if !verbose {
		return fmt.Sprintf("[%s] %s", e.Rule, e.Message)
	}

	// More detailed output for verbose mode
	var builder strings.Builder

	fmt.Fprintf(&builder, "Rule:    %s\n", e.Rule)
	fmt.Fprintf(&builder, "Code:    %s\n", e.Code)
	fmt.Fprintf(&builder, "Message: %s\n", e.Message)

	help := e.GetHelp()
	if help != "" {
		fmt.Fprintf(&builder, "Help:    %s\n", help)
	}

	// Skip help when printing context
	if len(e.Context) > 0 {
		fmt.Fprintf(&builder, "Context:\n")

		for k, v := range e.Context {
			if k != "help" {
				fmt.Fprintf(&builder, "  %s: %s\n", k, v)
			}
		}
	}

	return builder.String()
}

// FormatAsJSON formats the error as JSON.
func (e ValidationError) FormatAsJSON() ([]byte, error) {
	// Create a representation with help extracted from context
	representation := struct {
		Rule    string            `json:"rule"`
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Help    string            `json:"help,omitempty"`
		Context map[string]string `json:"context,omitempty"`
	}{
		Rule:    e.Rule,
		Code:    e.Code,
		Message: e.Message,
		Help:    e.GetHelp(),
		Context: make(map[string]string),
	}

	// Copy context excluding help
	for k, v := range e.Context {
		if k != "help" {
			representation.Context[k] = v
		}
	}

	return json.Marshal(representation)
}

// FormatAsMarkdown formats the error as Markdown for documentation.
func (e ValidationError) FormatAsMarkdown() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "### %s: %s\n\n", e.Rule, e.Message)
	fmt.Fprintf(&builder, "**Code:** `%s`\n\n", e.Code)

	help := e.GetHelp()
	if help != "" {
		fmt.Fprintf(&builder, "**Help:** %s\n\n", help)
	}

	// Add context section if needed (excluding help)
	hasContext := false

	for k := range e.Context {
		if k != "help" {
			hasContext = true

			break
		}
	}

	if hasContext {
		fmt.Fprintf(&builder, "**Context:**\n\n")

		for k, v := range e.Context {
			if k != "help" {
				fmt.Fprintf(&builder, "- `%s`: %s\n", k, v)
			}
		}

		fmt.Fprintf(&builder, "\n")
	}

	return builder.String()
}

// ===== Common Validation Error Helpers =====

// FormatError creates a rich error for format validation failures.
func FormatError(ruleName string, message string, helpText string, subject string) ValidationError {
	ctx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrInvalidFormat,
		message,
		helpText,
		ctx,
	)

	return err.WithContext("subject", subject)
}

// EmptyError creates a rich error for empty content validation failures.
func EmptyError(ruleName string, message string, helpText string) ValidationError {
	ctx := NewContext()

	return CreateRichError(
		ruleName,
		ErrEmptyDescription,
		message,
		helpText,
		ctx,
	)
}

// CaseError creates a rich error for case validation failures.
func CaseError(ruleName string, message string, helpText string,
	expectedCase string, actualCase string, firstWord string, subject string) ValidationError {
	ctx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrSubjectCase,
		message,
		helpText,
		ctx,
	)

	// Add standardized context information
	err = err.WithContext("expected_case", expectedCase)
	err = err.WithContext("actual_case", actualCase)
	err = err.WithContext("first_word", firstWord)
	err = err.WithContext("subject", subject)

	return err
}

// ReferenceError creates a rich error for reference validation failures (like JIRA issues).
func ReferenceError(ruleName string, message string, helpText string, context map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrInvalidReference,
		message,
		helpText,
		errorCtx,
	)

	// Add detailed context information
	for k, v := range context {
		err = err.WithContext(k, v)
	}

	return err
}

// LengthError creates a rich error for length validation failures.
func LengthError(ruleName string, message string, helpText string,
	actualLength int, maxLength int, subject string) ValidationError {
	ctx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrMaxLengthExceeded,
		message,
		helpText,
		ctx,
	)

	// Add standardized context information
	err = err.WithContext("actual_length", strconv.Itoa(actualLength))
	err = err.WithContext("max_length", strconv.Itoa(maxLength))
	err = err.WithContext("subject", subject)

	return err
}

// SignatureError creates a rich error for signature validation failures.
func SignatureError(ruleName string, message string, helpText string, context map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrInvalidSignature,
		message,
		helpText,
		errorCtx,
	)

	// Add detailed context information
	for k, v := range context {
		err = err.WithContext(k, v)
	}

	return err
}

// UTF8Error creates a rich error for invalid UTF-8 encoding validation failures.
func UTF8Error(ruleName string, message string, helpText string, subject string) ValidationError {
	ctx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrInvalidFormat,
		message,
		helpText,
		ctx,
	)

	return err.WithContext("subject", subject)
}

// BodyError creates a rich error for commit body validation failures.
func BodyError(ruleName string, message string, helpText string, context map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrInvalidBody,
		message,
		helpText,
		errorCtx,
	)

	// Add detailed context information
	for k, v := range context {
		err = err.WithContext(k, v)
	}

	return err
}

// SpellError creates a rich error for spelling validation failures.
func SpellError(ruleName string, message string, helpText string,
	originalWord string, correctedWord string, additionalContext map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrSpelling,
		message,
		helpText,
		errorCtx,
	)

	// Add standardized context information
	err = err.WithContext("original", originalWord)
	err = err.WithContext("corrected", correctedWord)

	// Add additional context if provided
	for k, v := range additionalContext {
		err = err.WithContext(k, v)
	}

	return err
}

// ConfigError creates a rich error for configuration validation failures.
func ConfigError(ruleName string, message string, helpText string, context map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		ErrInvalidConfig,
		message,
		helpText,
		errorCtx,
	)

	// Add detailed context information
	for k, v := range context {
		err = err.WithContext(k, v)
	}

	return err
}

// ImperativeError creates a rich error for imperative verb validation failures.
func ImperativeError(ruleName string, errorCode ValidationErrorCode, message string, helpText string,
	word string, verbType string, suggestedForm string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		errorCode,
		message,
		helpText,
		errorCtx,
	)

	// Add standardized context information
	err = err.WithContext("word", word)
	err = err.WithContext("type", verbType)

	// Add suggested form if provided
	if suggestedForm != "" {
		err = err.WithContext("suggested_form", suggestedForm)
	}

	return err
}

// SignOffError creates a rich error for sign-off validation failures.
func SignOffError(ruleName string, errorCode ValidationErrorCode, message string, helpText string,
	errorType string, body string, additionalContext map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		errorCode,
		message,
		helpText,
		errorCtx,
	)

	// Add standardized context information
	err = err.WithContext("original_code", errorType)
	err = err.WithContext("message", body)

	// Add additional context if provided
	for k, v := range additionalContext {
		if k != "original_code" && k != "message" { // Don't overwrite standardized context
			err = err.WithContext(k, v)
		}
	}

	return err
}

// JiraError creates a rich error for Jira reference validation failures.
func JiraError(ruleName string, errorCode ValidationErrorCode, message string, helpText string,
	jiraKey string, subject string, additionalContext map[string]string) ValidationError {
	errorCtx := NewContext()
	err := CreateRichError(
		ruleName,
		errorCode,
		message,
		helpText,
		errorCtx,
	)

	// Add standardized context information
	if jiraKey != "" {
		err = err.WithContext("key", jiraKey)
	}

	if subject != "" {
		err = err.WithContext("subject", subject)
	}

	// Add additional context if provided
	for k, v := range additionalContext {
		if k != "key" && k != "subject" { // Don't overwrite standardized context
			err = err.WithContext(k, v)
		}
	}

	return err
}
