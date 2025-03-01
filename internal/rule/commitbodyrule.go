// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"strings"

	"github.com/pkg/errors"
)

// CommitBodyRule validates the presence of a commit message body.
type CommitBodyRule struct {
	errors []error
}

// Name returns the identifier of this check.
func (c CommitBodyRule) Name() string {
	return "Commit Body"
}

// Result returns the check message.
func (c CommitBodyRule) Result() string {
	if len(c.errors) != 0 {
		return c.errors[0].Error()
	}

	return "Commit body is valid"
}

// Errors returns any violations of the check.
func (c CommitBodyRule) Errors() []error {
	return c.errors
}

// ValidateCommitBody validates the commit message body.
// It ensures the body is not empty (contains at least one non-DCO line)
// and has exactly one empty line between subject and body.
func ValidateCommitBody(message string) *CommitBodyRule {
	rule := &CommitBodyRule{}
	lines := strings.Split(strings.TrimPrefix(message, "\n"), "\n")

	// A valid commit body must have at least 3 lines:
	// - subject
	// - empty line
	// - body
	if len(lines) < 3 {
		rule.errors = append(rule.errors, errors.New("Commit requires a descriptive body explaining the changes"))

		return rule
	}

	// Second line must be empty
	if lines[1] != "" {
		rule.errors = append(rule.errors, errors.New("Commit message must have exactly one empty line between subject and body"))

		return rule
	}

	// Third line must be not empty
	if lines[2] == "" {
		rule.errors = append(rule.errors, errors.New("Commit message must have a non empty body text"))

		return rule
	}

	bodyContent := []string{}

	for _, line := range lines[2:] {
		if SignOffRegex.MatchString(strings.TrimSpace(line)) {
			continue
		}

		if line != "" {
			bodyContent = append(bodyContent, line)
		}
	}

	if len(bodyContent) == 0 {
		rule.errors = append(rule.errors, errors.New(`Commit body is required. 
Be specific, descriptive, and explain the why behind changes while staying brief.

Example - A commit subject with body and sign-off:

feat: update password validation to meet NIST guidelines

- Increases minimum length to 12 characters
- Adds check against compromised password database

Signed-off-by: Humpty Dumpty <supercommiter@example.com>`))

		return rule
	}

	return rule
}
