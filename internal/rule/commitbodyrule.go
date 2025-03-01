// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"strings"

	"github.com/pkg/errors"
)

// CommitBodyRule validates the presence and format of a commit message body.
// It ensures the body contains meaningful content and follows proper formatting.
type CommitBodyRule struct {
	errors []error
}

// Name returns the rule identifier.
func (c CommitBodyRule) Name() string {
	return "CommitBodyRule"
}

// Result returns a string representation of the validation result.
func (c CommitBodyRule) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return "Commit body is valid"
}

// Errors returns all validation errors found.
func (c CommitBodyRule) Errors() []error {
	return c.errors
}

// ValidateCommitBodyRule checks that the commit message has a proper body
// with one empty line between subject and body. It filters out sign-off lines.
func ValidateCommitBodyRule(message string) *CommitBodyRule {
	rule := &CommitBodyRule{}
	lines := strings.Split(message, "\n")

	if !hasValidStructure(lines, rule) {
		return rule
	}

	if !hasNonEmptyBody(lines, rule) {
		return rule
	}

	return rule
}

// hasValidStructure checks if the commit message has the minimum required
// structure: subject line, empty line, and at least one body line.
func hasValidStructure(lines []string, rule *CommitBodyRule) bool {
	if len(lines) < 3 {
		rule.errors = append(rule.errors, errors.New("Commit message requires a body explaining the changes"))

		return false
	}

	if lines[1] != "" {
		rule.errors = append(rule.errors, errors.New("Commit message must have exactly one empty line between the subject and the body"))

		return false
	}

	if lines[2] == "" {
		rule.errors = append(rule.errors, errors.New("Commit message must have a non-empty body text"))

		return false
	}

	return true
}

// hasNonEmptyBody checks if the commit message body has meaningful content
// beyond just sign-off lines.
func hasNonEmptyBody(lines []string, rule *CommitBodyRule) bool {
	var bodyContent []string

	for _, line := range lines[2:] {
		trimmedLine := strings.TrimSpace(line)
		if SignOffRegex.MatchString(trimmedLine) {
			continue
		}

		if trimmedLine != "" {
			bodyContent = append(bodyContent, line)
		}
	}

	if len(bodyContent) == 0 {
		rule.errors = append(rule.errors, errors.New(`Commit message body is required.
Be specific, descriptive, and explain the why behind changes while staying brief.

Example - A commit subject with a body and a sign-off:

feat: update password validation to meet NIST guidelines

- Increases minimum length to 12 characters
- Adds check against compromised password database

Signed-off-by: Laval Lion <laval@cavora.org>`))

		return false
	}

	return true
}
