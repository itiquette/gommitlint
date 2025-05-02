// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
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
	baseRule            BaseRule
	requireSignOff      bool
	allowMultiple       bool
	customRegex         *regexp.Regexp
	hasAttemptedSignOff bool   // Track if there was an attempt at signing off
	foundSignOff        string // Store the found sign-off for verbose output
	errors              []appErrors.ValidationError
}

// SignOffOption is a function that modifies a SignOffRule.
type SignOffOption func(SignOffRule) SignOffRule

// WithRequireSignOff configures whether the sign-off is mandatory.
func WithRequireSignOff(require bool) SignOffOption {
	return func(rule SignOffRule) SignOffRule {
		result := rule
		result.requireSignOff = require

		return result
	}
}

// WithAllowMultipleSignOffs configures whether multiple sign-offs are allowed.
func WithAllowMultipleSignOffs(allow bool) SignOffOption {
	return func(rule SignOffRule) SignOffRule {
		result := rule
		result.allowMultiple = allow

		return result
	}
}

// WithCustomSignOffRegex sets a custom regular expression for validating sign-offs.
func WithCustomSignOffRegex(regex *regexp.Regexp) SignOffOption {
	return func(rule SignOffRule) SignOffRule {
		result := rule
		if regex != nil {
			result.customRegex = regex
		}

		return result
	}
}

// NewSignOffRule creates a new SignOffRule with the specified options.
func NewSignOffRule(options ...SignOffOption) SignOffRule {
	rule := SignOffRule{
		baseRule:       NewBaseRule("SignOff"),
		requireSignOff: true, // By default, sign-off is mandatory
		allowMultiple:  true, // By default, allow multiple sign-offs
		customRegex:    nil,  // By default, use the standard regex
		errors:         make([]appErrors.ValidationError, 0),
	}
	// Apply provided options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// NewSignOffRuleWithConfig creates a SignOffRule using configuration.
func NewSignOffRuleWithConfig(config domain.SecurityConfigProvider) SignOffRule {
	// Build options based on the configuration
	var options []SignOffOption

	// Set whether sign-offs are required
	options = append(options, WithRequireSignOff(config.SignOffRequired()))

	// Set whether multiple sign-offs are allowed
	options = append(options, WithAllowMultipleSignOffs(config.AllowMultipleSignOffs()))

	return NewSignOffRule(options...)
}

// SetErrors sets the validation errors for this rule and returns a new instance.
func (r SignOffRule) SetErrors(errors []appErrors.ValidationError) SignOffRule {
	result := r
	result.errors = errors

	// Also update baseRule for consistency
	baseRule := r.baseRule
	for _, err := range errors {
		baseRule = baseRule.WithError(err)
	}

	result.baseRule = baseRule

	return result
}

// SetSignOffInfo sets additional sign-off information (attempted, found) and returns a new instance.
func (r SignOffRule) SetSignOffInfo(attempted bool, found string) SignOffRule {
	result := r
	result.hasAttemptedSignOff = attempted
	result.foundSignOff = found

	return result
}

// Errors returns all validation errors.
func (r SignOffRule) Errors() []appErrors.ValidationError {
	return r.errors
}

// HasErrors checks if the rule has any validation errors.
func (r SignOffRule) HasErrors() bool {
	return len(r.errors) > 0
}

// Name returns the rule name.
func (r SignOffRule) Name() string {
	return r.baseRule.Name()
}

// Validate validates a commit message against the rule.
func (r SignOffRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Create local variables for tracking state - don't modify r directly
	errors := make([]appErrors.ValidationError, 0)

	// If sign-off is not required, pass immediately
	if !r.requireSignOff {
		return errors
	}

	body := commit.Body
	// Handle empty body
	if strings.TrimSpace(body) == "" {
		// Create error context with rich information
		errorCtx := appErrors.NewContext()

		helpMessage := `Empty Commit Body Error: Cannot find sign-off in an empty message.

A sign-off line must be present in the commit body to certify you have the right to submit
your contribution under the project's license (Developer Certificate of Origin).

✅ CORRECT FORMAT:
- A commit with a proper body and sign-off:

  feat: add new feature
  
  This commit adds a new feature that does something useful.
  Here's more description about what it does.
  
  Signed-off-by: Your Name <your.email@example.com>

❌ INCORRECT FORMAT:
- Your commit has an empty body with no sign-off line

WHY THIS MATTERS:
- Sign-offs provide a legal attestation that you have the right to contribute
- They help establish a clear chain of provenance for all code
- They're required by many open source projects for legal compliance

NEXT STEPS:
1. Add a meaningful commit message body with a description of your changes
2. Add a sign-off line at the end using one of these methods:
   - Use 'git commit --signoff' or 'git commit -s' to add it automatically
   - Manually add 'Signed-off-by: Your Name <your.email@example.com>' at the end`

		err := appErrors.CreateRichError(
			r.Name(),
			appErrors.ErrEmptyMessage,
			"commit message body is empty; no sign-off found",
			helpMessage,
			errorCtx,
		)

		// Store the original code for detailed error information
		err = err.WithContext("original_code", "empty_message")

		errors = append(errors, err)

		return errors
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
			errors = append(errors, r.createError(
				"multiple_signoffs",
				"commit has multiple sign-offs but only one is allowed",
				map[string]string{
					"message":       body,
					"signoff_count": strconv.Itoa(len(validSignOffs)),
				},
			))
		}

		return errors
	}

	// Check if there are any lines that attempt to be a sign-off but are formatted incorrectly
	hasAttemptedSignOff := false

	for _, line := range allLines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(trimmedLine, "Signed") && strings.Contains(trimmedLine, "by:") {
			hasAttemptedSignOff = true

			break
		}
	}

	// No valid sign-off found - distinguish between format issues and completely missing signoff
	if hasAttemptedSignOff {
		// Create error context with rich information
		errorCtx := appErrors.NewContext()

		helpMessage := `Invalid Sign-off Format Error: The sign-off line format is incorrect.

Your commit appears to have an attempted sign-off line, but it doesn't match the required format.

✅ CORRECT FORMAT:
- The sign-off line must follow this exact format:
  
  Signed-off-by: Your Name <your.email@example.com>
  
  Examples:
  Signed-off-by: Jane Doe <jane.doe@example.com>
  Signed-off-by: John Smith <john.smith@company.org>

❌ COMMON FORMATTING MISTAKES:
- Misspelled "Signed-off-by:" (e.g., "Signed by:" or "Singed-off-by:")
- Missing hyphen (e.g., "Signed off by:")
- Using parentheses instead of angle brackets (e.g., "Signed-off-by: Name (email@example.com)")
- Incorrect email format (e.g., missing @ or domain)
- Extra spaces in the wrong places
- Special characters in the wrong places

WHY THIS MATTERS:
- Precise sign-off formatting enables automatic verification tools
- The format follows the Developer Certificate of Origin standard
- Consistent formatting makes review and compliance checking easier

NEXT STEPS:
1. Add a correctly formatted sign-off:
   Signed-off-by: Your Name <your.email@example.com>

2. Or use Git's built-in signoff feature:
   git commit --signoff   (or the shorthand -s flag)
   
3. Ensure your Git user.name and user.email are configured correctly:
   git config --global user.name "Your Name"
   git config --global user.email "your.email@example.com"`

		err := appErrors.CreateRichError(
			r.Name(),
			appErrors.ErrInvalidFormat,
			"Invalid sign-off format. Must use 'Signed-off-by: Name <email@example.com>' format",
			helpMessage,
			errorCtx,
		)

		// Store message context and original code for detailed error information
		err = err.WithContext("message", body)
		err = err.WithContext("original_code", "invalid_format")

		errors = append(errors, err)
	} else {
		// Create error context with rich information
		errorCtx := appErrors.NewContext()

		helpMessage := `Missing Sign-off Error: No Developer Certificate of Origin sign-off found.

Your commit message is missing a required Developer Certificate of Origin (DCO) sign-off line.

✅ CORRECT FORMAT:
- A commit with a proper sign-off looks like this:
  
  feat: add new feature
  
  This commit adds a new feature that does something useful.
  Here's more description about what it does.
  
  Signed-off-by: Your Name <your.email@example.com>

❌ INCORRECT FORMAT:
- Your commit has no sign-off line at all
- The sign-off is not at the end of the commit message

WHY THIS MATTERS:
- The sign-off certifies you have the right to submit your contribution
- It's a legal attestation that your contribution complies with the project's licensing
- It helps establish clear provenance for all contributed code

NEXT STEPS:
1. Add a sign-off line at the end of your commit message using one of these methods:
   - Use 'git commit --signoff' or 'git commit -s' to add it automatically
   - Manually add 'Signed-off-by: Your Name <your.email@example.com>' at the end
   
2. Ensure your Git user name and email are configured correctly:
   git config --global user.name "Your Name"
   git config --global user.email "your.email@example.com"`

		err := appErrors.CreateRichError(
			r.Name(),
			appErrors.ErrMissingSignoff,
			"Missing Signed-off-by line. Commit must be signed-off using 'Signed-off-by: Name <email@example.com>' format",
			helpMessage,
			errorCtx,
		)

		// Store message context and original code for detailed error information
		err = err.WithContext("message", body)
		err = err.WithContext("original_code", "missing_signoff")

		errors = append(errors, err)
	}

	return errors
}

// Result returns a concise rule message.
func (r SignOffRule) Result(errors []errors.ValidationError) string {
	if r.HasErrors() {
		return "Missing sign-off"
	}

	return "Sign-off is present"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SignOffRule) VerboseResult(errors []errors.ValidationError) string {
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
func (r SignOffRule) Help(errors []errors.ValidationError) string {
	if !r.HasErrors() {
		return "No errors to fix. This rule checks that commits have a proper Signed-off-by line indicating Developer Certificate of Origin (DCO) agreement."
	}
	// Check error code for more targeted help
	if len(errors) > 0 {
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

// createError creates a structured validation error without modifying the rule's state.
func (r SignOffRule) createError(code, message string, context map[string]string) appErrors.ValidationError {
	if context == nil {
		context = make(map[string]string)
	}

	// Extract message body if available (or use empty string)
	messageBody := ""
	if body, ok := context["message"]; ok {
		messageBody = body
	}

	// Get error code based on the original code
	var errorCode appErrors.ValidationErrorCode

	var helpMessage string

	if code == "missing_signoff" {
		errorCode = appErrors.ErrMissingSignoff
		helpMessage = `Missing Sign-off Error: No Developer Certificate of Origin sign-off found.

Your commit message is missing a required Developer Certificate of Origin (DCO) sign-off line.

✅ CORRECT FORMAT:
- A commit with a proper sign-off looks like this:
  
  feat: add new feature
  
  This commit adds a new feature that does something useful.
  Here's more description about what it does.
  
  Signed-off-by: Your Name <your.email@example.com>

❌ INCORRECT FORMAT:
- Your commit has no sign-off line at all
- The sign-off is not at the end of the commit message

WHY THIS MATTERS:
- The sign-off certifies you have the right to submit your contribution
- It's a legal attestation that your contribution complies with the project's licensing
- It helps establish clear provenance for all contributed code

NEXT STEPS:
1. Add a sign-off line at the end of your commit message using one of these methods:
   - Use 'git commit --signoff' or 'git commit -s' to add it automatically
   - Manually add 'Signed-off-by: Your Name <your.email@example.com>' at the end
   
2. Ensure your Git user name and email are configured correctly:
   git config --global user.name "Your Name"
   git config --global user.email "your.email@example.com"`
	} else if code == "invalid_format" {
		errorCode = appErrors.ErrInvalidFormat
		message = "Invalid sign-off format. Must use 'Signed-off-by: Name <email@example.com>' format"
		helpMessage = `Invalid Sign-off Format Error: The sign-off line format is incorrect.

Your commit appears to have an attempted sign-off line, but it doesn't match the required format.

✅ CORRECT FORMAT:
- The sign-off line must follow this exact format:
  
  Signed-off-by: Your Name <your.email@example.com>
  
  Examples:
  Signed-off-by: Jane Doe <jane.doe@example.com>
  Signed-off-by: John Smith <john.smith@company.org>

❌ COMMON FORMATTING MISTAKES:
- Misspelled "Signed-off-by:" (e.g., "Signed by:" or "Singed-off-by:")
- Missing hyphen (e.g., "Signed off by:")
- Using parentheses instead of angle brackets (e.g., "Signed-off-by: Name (email@example.com)")
- Incorrect email format (e.g., missing @ or domain)
- Extra spaces in the wrong places
- Special characters in the wrong places

WHY THIS MATTERS:
- Precise sign-off formatting enables automatic verification tools
- The format follows the Developer Certificate of Origin standard
- Consistent formatting makes review and compliance checking easier

NEXT STEPS:
1. Add a correctly formatted sign-off:
   Signed-off-by: Your Name <your.email@example.com>

2. Or use Git's built-in signoff feature:
   git commit --signoff   (or the shorthand -s flag)
   
3. Ensure your Git user.name and user.email are configured correctly:
   git config --global user.name "Your Name"
   git config --global user.email "your.email@example.com"`
	} else if code == "multiple_signoffs" {
		errorCode = appErrors.ErrMissingSignoff
		helpMessage = `Multiple Sign-offs Error: Too many sign-off lines detected.

Your commit contains multiple Developer Certificate of Origin sign-off lines, but the project configuration only allows one.

✅ CORRECT FORMAT:
- A commit should have exactly one sign-off line:
  
  Signed-off-by: Your Name <your.email@example.com>

❌ INCORRECT FORMAT:
- Your commit has multiple sign-off lines:
  
  Signed-off-by: First Person <first.person@example.com>
  Signed-off-by: Second Person <second.person@example.com>

WHY THIS MATTERS:
- Single sign-offs establish clear accountability for contributions
- Multiple sign-offs can create confusion about responsibility
- The project has explicitly configured validation to require a single sign-off

NEXT STEPS:
1. Edit your commit message to include only one sign-off line:
   - Choose the primary contributor's sign-off to keep
   - Remove all other sign-off lines
   
2. Use 'git commit --amend' to edit your most recent commit

3. If you believe multiple sign-offs should be allowed, contact the project maintainers
   to discuss changing the configuration policy`
	} else if code == "empty_message" {
		errorCode = appErrors.ErrEmptyMessage
		helpMessage = `Empty Commit Body Error: Cannot find sign-off in an empty message.

A sign-off line must be present in the commit body to certify you have the right to submit
your contribution under the project's license (Developer Certificate of Origin).

✅ CORRECT FORMAT:
- A commit with a proper body and sign-off:

  feat: add new feature
  
  This commit adds a new feature that does something useful.
  Here's more description about what it does.
  
  Signed-off-by: Your Name <your.email@example.com>

❌ INCORRECT FORMAT:
- Your commit has an empty body with no sign-off line

WHY THIS MATTERS:
- Sign-offs provide a legal attestation that you have the right to contribute
- They help establish a clear chain of provenance for all code
- They're required by many open source projects for legal compliance

NEXT STEPS:
1. Add a meaningful commit message body with a description of your changes
2. Add a sign-off line at the end using one of these methods:
   - Use 'git commit --signoff' or 'git commit -s' to add it automatically
   - Manually add 'Signed-off-by: Your Name <your.email@example.com>' at the end`
	} else {
		// For other codes, use default unknown error
		errorCode = appErrors.ErrUnknown
		helpMessage = `Sign-off Error: There was an issue with your commit's sign-off line.

Your commit message has an issue related to the Developer Certificate of Origin sign-off line.

✅ CORRECT FORMAT:
- A commit should have a properly formatted sign-off line:
  
  Signed-off-by: Your Name <your.email@example.com>

WHY THIS MATTERS:
- Sign-offs provide a legal attestation that you have the right to contribute
- They help establish a clear chain of provenance for all code
- They're required by many open source projects for legal compliance

NEXT STEPS:
1. Ensure your commit message has a properly formatted sign-off line
2. Add a sign-off line using 'git commit --signoff' or 'git commit -s'
3. Manually add 'Signed-off-by: Your Name <your.email@example.com>' at the end of your commit message`
	}

	// Use the new SignOffError helper function
	return appErrors.SignOffError(
		r.Name(),
		errorCode,
		message,
		helpMessage,
		code,
		messageBody,
		context,
	)
}
