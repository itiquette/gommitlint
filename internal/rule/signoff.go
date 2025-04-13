// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/model"
)

// SignOffRegex is the regular expression used to validate the Developer Certificate of Origin signature.
// It matches the standard format "Signed-off-by: Name <email@example.com>".
var SignOffRegex = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>$`)

// SignOff enforces the presence and format of a Developer Certificate of Origin (DCO) sign-off
// in commit messages.
//
// This rule helps ensure that all contributors formally certify they have the right to submit
// their code under the project's license, which is important for legal compliance and
// establishing clean provenance for all contributions.
//
// The Developer Certificate of Origin (DCO) is a lightweight alternative to more complex
// contributor license agreements, requiring a simple sign-off line in each commit message
// that attests to the contributor's right to submit the code.
//
// For a sign-off to be valid, it must:
//   - Appear on its own line in the commit message
//   - Follow the exact format: "Signed-off-by: Name <email@example.com>"
//   - Use the contributor's actual name and email address
//
// Examples:
//
//   - Valid commit message with sign-off:
//     ```
//     feat: add user authentication
//
//     Implement basic user authentication using JWT.
//
//     Signed-off-by: Jane Doe <jane@example.com>
//     ```
//
//   - Invalid commit message (missing sign-off):
//     ```
//     fix: resolve memory leak
//
//     Fixed memory leak in the connection pool.
//     ```
//
//   - Invalid sign-off format:
//     ```
//     feat: add new API endpoint
//
//     Added /api/v1/users endpoint for user management.
//
//     Signed by: John Smith (john@example.com)
//     ```
type SignOff struct {
	errors              []*model.ValidationError
	hasAttemptedSignOff bool   // Track if there was an attempt at signing off
	foundSignOff        string // Store the found sign-off for verbose output
}

// Name returns the name of the rule.
func (SignOff) Name() string {
	return "SignOff"
}

// Result returns a concise rule message.
func (rule SignOff) Result() string {
	if len(rule.errors) != 0 {
		return "Missing sign-off"
	}

	return "Sign-off exists"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (rule SignOff) VerboseResult() string {
	if len(rule.errors) != 0 {
		switch rule.errors[0].Code {
		case "empty_message":
			return "Commit message body is empty. Cannot find sign-off in an empty message."
		case "missing_signoff":
			return "No Developer Certificate of Origin sign-off found in commit message. Add 'Signed-off-by: Name <email@example.com>'."
		case "invalid_format":
			return "Attempted sign-off has incorrect format. Must be exactly: 'Signed-off-by: Name <email@example.com>'."
		default:
			return rule.errors[0].Error()
		}
	}

	if rule.foundSignOff != "" {
		return "Valid Developer Certificate of Origin sign-off found: " + rule.foundSignOff
	}

	return "Valid sign-off exists in commit message"
}

// Errors returns any violations of the rule.
func (rule SignOff) Errors() []*model.ValidationError {
	return rule.errors
}

// addError adds a structured validation error.
func (rule *SignOff) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("SignOff", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	rule.errors = append(rule.errors, err)
}

// Help returns a description of how to fix the rule violation.
func (rule SignOff) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

	// Check error code for more targeted help
	if len(rule.errors) > 0 {
		switch rule.errors[0].Code {
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

// ValidateSignOff checks if the commit message contains a valid Developer Certificate of Origin sign-off.
//
// Parameters:
//   - body: The commit message body to validate
//
// The function examines each line of the commit message, looking for a valid sign-off
// that matches the standard DCO format: "Signed-off-by: Name <email@example.com>".
//
// A valid sign-off certifies that the contributor has the right to submit the code
// under the project's license and adheres to the Developer Certificate of Origin
// (https://developercertificate.org/).
//
// Common validation failures include:
//   - Missing sign-off line entirely
//   - Incorrect format (e.g., "Signed by" instead of "Signed-off-by")
//   - Invalid email format
//   - Empty commit message
//
// Returns:
//   - A SignOff instance with validation results
func ValidateSignOff(body string) *SignOff {
	rule := &SignOff{}

	// Handle empty body
	if strings.TrimSpace(body) == "" {
		rule.addError(
			"empty_message",
			"commit message body is empty; no sign-off found",
			nil,
		)

		return rule
	}

	// Check each line for a sign-off
	allLines := strings.Split(body, "\n")
	for _, line := range allLines {
		trimmedLine := strings.TrimSpace(line)
		if SignOffRegex.MatchString(trimmedLine) {
			rule.foundSignOff = trimmedLine

			return rule // Found a valid sign-off
		}
	}

	// Check if there are any lines that attempt to be a sign-off but are formatted incorrectly
	rule.hasAttemptedSignOff = false

	for _, line := range allLines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(trimmedLine, "Signed") && strings.Contains(trimmedLine, "by:") {
			rule.hasAttemptedSignOff = true

			break
		}
	}

	// No valid sign-off found - distinguish between format issues and completely missing signoff
	if rule.hasAttemptedSignOff {
		rule.addError(
			"invalid_format",
			"commit must be signed-off using 'Signed-off-by: Name <email@example.com>' format",
			map[string]string{
				"message": body,
			},
		)
	} else {
		rule.addError(
			"missing_signoff",
			"commit must be signed-off using 'Signed-off-by: Name <email@example.com>' format",
			map[string]string{
				"message": body,
			},
		)
	}

	return rule
}
