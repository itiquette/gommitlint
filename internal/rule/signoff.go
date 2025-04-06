// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"regexp"
	"strings"
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
	errors []error
}

// Name returns the name of the rule.
func (SignOff) Name() string {
	return "SignOff"
}

// Result returns the rule message.
func (rule SignOff) Result() string {
	if len(rule.errors) != 0 {
		return rule.errors[0].Error()
	}

	return "Sign-off exists"
}

// Errors returns any violations of the rule.
func (rule SignOff) Errors() []error {
	return rule.errors
}

// addErrorf adds an error to the rule's errors slice.
func (rule *SignOff) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}

// Help returns a description of how to fix the rule violation.
func (rule SignOff) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

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
		rule.addErrorf("commit message body is empty; no sign-off found")

		return rule
	}

	// Check each line for a sign-off
	for _, line := range strings.Split(body, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if SignOffRegex.MatchString(trimmedLine) {
			return rule // Found a valid sign-off
		}
	}

	// No valid sign-off found
	rule.addErrorf("commit must be signed-off using 'Signed-off-by: Name <email@example.com>' format")

	return rule
}
