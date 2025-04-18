// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SignOffRegex is the regular expression used to validate the Developer Certificate of Origin signature.
// It matches the standard format "Signed-off-by: Name <email@example.com>".
var SignOffRegex = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>]+@[^<>]+\.[^<>]+)>$`)

// SignOffRule enforces the presence and format of a Developer Certificate of Origin (DCO) sign-off
// in commit messages.
//
// This rule helps ensure that all contributors formally certify they have the right to submit
// their code under the project's license, which is important for legal compliance and
// establishing clean provenance for all contributions.
type SignOffRule struct {
	*BaseRule
	requireSignOff      bool
	allowMultiple       bool
	customRegex         *regexp.Regexp
	hasAttemptedSignOff bool   // Track if there was an attempt at signing off
	foundSignOff        string // Store the found sign-off for verbose output
}

// SignOffOption is a function that modifies a SignOffRule.
type SignOffOption func(*SignOffRule)

// WithRequireSignOff configures whether the sign-off is mandatory.
func WithRequireSignOff(require bool) SignOffOption {
	return func(rule *SignOffRule) {
		rule.requireSignOff = require
	}
}

// WithAllowMultipleSignOffs configures whether multiple sign-offs are allowed.
func WithAllowMultipleSignOffs(allow bool) SignOffOption {
	return func(rule *SignOffRule) {
		rule.allowMultiple = allow
	}
}

// WithCustomSignOffRegex sets a custom regular expression for validating sign-offs.
func WithCustomSignOffRegex(regex *regexp.Regexp) SignOffOption {
	return func(rule *SignOffRule) {
		if regex != nil {
			rule.customRegex = regex
		}
	}
}

// NewSignOffRule creates a new SignOffRule with the specified options.
func NewSignOffRule(options ...SignOffOption) *SignOffRule {
	rule := &SignOffRule{
		BaseRule:       NewBaseRule("SignOff"),
		requireSignOff: true, // By default, sign-off is mandatory
		allowMultiple:  true, // By default, allow multiple sign-offs
		customRegex:    nil,  // By default, use the standard regex
	}

	// Apply provided options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Validate validates a commit message against the rule.
func (r *SignOffRule) Validate(commit *domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors
	r.ClearErrors()
	r.MarkAsRun()
	r.hasAttemptedSignOff = false
	r.foundSignOff = ""

	// If sign-off is not required, pass immediately
	if !r.requireSignOff {
		// Set a dummy sign-off for the verbose output
		r.foundSignOff = "Not required"

		return r.Errors()
	}

	body := commit.Body

	// Handle empty body
	if strings.TrimSpace(body) == "" {
		r.addError(
			"empty_message",
			"commit message body is empty; no sign-off found",
			nil,
		)

		return r.Errors()
	}

	// Determine which regex to use
	signOffRegex := SignOffRegex
	if r.customRegex != nil {
		signOffRegex = r.customRegex
	}

	// Check each line for a sign-off
	allLines := strings.Split(body, "\n")
	validSignOffs := []string{}

	for _, line := range allLines {
		trimmedLine := strings.TrimSpace(line)
		if signOffRegex.MatchString(trimmedLine) {
			validSignOffs = append(validSignOffs, trimmedLine)
		}
	}

	// Handle different cases based on configuration
	if len(validSignOffs) > 0 {
		// Found at least one valid sign-off
		if !r.allowMultiple && len(validSignOffs) > 1 {
			r.addError(
				"multiple_signoffs",
				"commit has multiple sign-offs but only one is allowed",
				map[string]string{
					"message":       body,
					"signoff_count": strconv.Itoa(len(validSignOffs)),
				},
			)

			return r.Errors()
		}

		// Store the first valid sign-off for verbose output
		r.foundSignOff = validSignOffs[0]

		return r.Errors()
	}

	// Check if there are any lines that attempt to be a sign-off but are formatted incorrectly
	r.hasAttemptedSignOff = false

	for _, line := range allLines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(trimmedLine, "Signed") && strings.Contains(trimmedLine, "by:") {
			r.hasAttemptedSignOff = true

			break
		}
	}

	// No valid sign-off found - distinguish between format issues and completely missing signoff
	if r.hasAttemptedSignOff {
		r.addError(
			"invalid_format",
			"Invalid sign-off format. Must use 'Signed-off-by: Name <email@example.com>' format",
			map[string]string{
				"message": body,
			},
		)
	} else {
		r.addError(
			"missing_signoff",
			"Missing Signed-off-by line. Commit must be signed-off using 'Signed-off-by: Name <email@example.com>' format",
			map[string]string{
				"message": body,
			},
		)
	}

	return r.Errors()
}

// Result returns a concise rule message.
func (r *SignOffRule) Result() string {
	if r.HasErrors() {
		return "Missing sign-off"
	}

	return "Sign-off is present"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *SignOffRule) VerboseResult() string {
	if r.HasErrors() {
		errors := r.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		// Check original code in context for more specific message
		if originalCode, ok := validationErr.Context["original_code"]; ok {
			switch originalCode {
			case "empty_message":
				return "Commit message body is empty. Cannot find sign-off in an empty message."
			case "missing_signoff":
				return "No Developer Certificate of Origin sign-off found in commit message. Add 'Signed-off-by: Name <email@example.com>'."
			case "invalid_format":
				return "Attempted sign-off has incorrect format. Must be exactly: 'Signed-off-by: Name <email@example.com>'."
			case "multiple_signoffs":
				return "Multiple Developer Certificate of Origin sign-offs found, but configuration only allows one."
			}
		}

		// If no original_code in context or not recognized, fall back to error message
		return validationErr.Error()
	}

	if r.foundSignOff == "Not required" {
		return "Sign-off is not required by configuration. Valid Developer Certificate of Origin sign-off skipped."
	} else if r.foundSignOff != "" {
		return "Sign-off is present. Valid Developer Certificate of Origin sign-off found: " + r.foundSignOff
	}

	return "Sign-off is present in commit message"
}

// Help returns a description of how to fix the rule violation.
func (r *SignOffRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix"
	}

	// Check error code for more targeted help
	errors := r.Errors()
	if len(errors) > 0 {
		// Try to cast to ValidationError
		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		// First check context for original error code
		if originalCode, ok := validationErr.Context["original_code"]; ok {
			switch originalCode {
			case "empty_message":
				return `Add a Developer Certificate of Origin sign-off to your commit message.
Your commit message is currently empty. First, provide a meaningful commit message,
then add a sign-off line at the end.

You can add a sign-off automatically using 'git commit -s' or manually add:
Signed-off-by: Your Name <your.email@example.com>

The Developer Certificate of Origin is a statement that you have the right to 
submit this contribution under the project's license.`

			case "missing_signoff":
				return `Add a Developer Certificate of Origin sign-off to your commit message.
You can do this by:
1. Using 'git commit -s' which will automatically add the sign-off
2. Manually adding a line at the end of your commit message:
   Signed-off-by: Your Name <your.email@example.com>

The sign-off certifies you have the right to submit your contribution 
under the project's license and follows the Developer Certificate of Origin.

Example of a complete commit message with sign-off:
feat: introduce rate limiting for API endpoints

Adds rate limiting to prevent API abuse:
- Implements token bucket algorithm
- Configurable limits per endpoint

Signed-off-by: Jane Doe <jane.doe@example.org>`

			case "invalid_format":
				return `Add a correctly formatted Developer Certificate of Origin sign-off to your commit message.
The sign-off line must follow this exact format:
Signed-off-by: Your Name <your.email@example.com>

Common issues include:
- Misspelling "Signed-off-by"
- Using parentheses instead of angle brackets for email
- Using incorrect email format

You can add a correct sign-off automatically using 'git commit -s'`

			case "multiple_signoffs":
				return `Your commit has multiple Developer Certificate of Origin sign-offs, but the configuration only allows one.
Remove all but one sign-off line to comply with the project's requirements.

Example of a correctly formatted sign-off:
Signed-off-by: Your Name <your.email@example.com>

If you need to keep multiple sign-offs, contact the project maintainers to update the configuration.`
			}
		}

		// Fall back to checking the error code if no original_code in context
		switch validationErr.Code {
		case string(appErrors.ErrEmptyMessage):
			return `Add a Developer Certificate of Origin sign-off to your commit message.
Your commit message is currently empty. First, provide a meaningful commit message,
then add a sign-off line at the end.

You can add a sign-off automatically using 'git commit -s' or manually add:
Signed-off-by: Your Name <your.email@example.com>

The Developer Certificate of Origin is a statement that you have the right to 
submit this contribution under the project's license.`

		case string(appErrors.ErrMissingSignoff):
			return `Add a Developer Certificate of Origin sign-off to your commit message.
You can do this by:
1. Using 'git commit -s' which will automatically add the sign-off
2. Manually adding a line at the end of your commit message:
   Signed-off-by: Your Name <your.email@example.com>

The sign-off certifies you have the right to submit your contribution 
under the project's license and follows the Developer Certificate of Origin.

Example of a complete commit message with sign-off:
feat: introduce rate limiting for API endpoints

Adds rate limiting to prevent API abuse:
- Implements token bucket algorithm
- Configurable limits per endpoint

Signed-off-by: Jane Doe <jane.doe@example.org>`

		case string(appErrors.ErrInvalidFormat):
			return `Add a correctly formatted Developer Certificate of Origin sign-off to your commit message.
The sign-off line must follow this exact format:
Signed-off-by: Your Name <your.email@example.com>

Common issues include:
- Misspelling "Signed-off-by"
- Using parentheses instead of angle brackets for email
- Using incorrect email format

You can add a correct sign-off automatically using 'git commit -s'`
		}
	}

	// Default help text
	return `Add a Developer Certificate of Origin sign-off to your commit message.
You can do this by:
1. Using 'git commit -s' which will automatically add the sign-off
2. Manually adding a line at the end of your commit message:
   Signed-off-by: Your Name <your.email@example.com>

The sign-off certifies you have the right to submit your contribution 
under the project's license and follows the Developer Certificate of Origin.

Example of a complete commit message with sign-off:
feat: introduce rate limiting for API endpoints

Adds rate limiting to prevent API abuse:
- Implements token bucket algorithm
- Configurable limits per endpoint

Signed-off-by: Jane Doe <jane.doe@example.org>`
}

// addError adds a structured validation error.
func (r *SignOffRule) addError(code, message string, context map[string]string) {
	if context == nil {
		context = make(map[string]string)
	}

	// Store original code in context for test validation
	context["original_code"] = code

	if code == "missing_signoff" {
		// Use the error template for missing signoff with context in one step
		r.AddErrorWithContext(appErrors.ErrMissingSignoff, message, context)
	} else if code == "invalid_format" {
		// Use a more consistent error message for invalid format
		standardMessage := "Invalid sign-off format. Must use 'Signed-off-by: Name <email@example.com>' format"
		r.AddErrorWithContext(appErrors.ErrInvalidFormat, standardMessage, context)
	} else if code == "multiple_signoffs" {
		// Use appropriate error code for multiple signoffs
		r.AddErrorWithContext(appErrors.ErrMissingSignoff, message, context)
	} else if code == "empty_message" {
		// Use appropriate error code for empty message
		r.AddErrorWithContext(appErrors.ErrEmptyMessage, message, context)
	} else {
		// For other codes, use context if available
		r.AddErrorWithContext(appErrors.ErrUnknown, message, context)
	}
}
