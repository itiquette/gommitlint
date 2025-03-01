// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

// SubjectRegex Format: type(scope)!: description.
var SubjectRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// ConventionalCommitRule validates that commit messages follow the
// conventional commit specification.
type ConventionalCommitRule struct {
	errors []error
}

// Name returns the rule identifier.
func (c ConventionalCommitRule) Name() string {
	return "ConventionalCommitRule"
}

// Result returns a string representation of the validation result.
func (c ConventionalCommitRule) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return "Commit message is a valid conventional commit"
}

// Errors returns all validation errors.
func (c ConventionalCommitRule) Errors() []error {
	return c.errors
}

// ValidateConventionalCommitRule checks if a commit subject follows conventional format.
// It validates the type, scope (if provided), and description length.
// If descLength is 0, defaults to 72 characters.
func ValidateConventionalCommitRule(subject string, types []string, scopes []string, descLength int) *ConventionalCommitRule {
	rule := &ConventionalCommitRule{}

	// Default description length if not specified
	if descLength == 0 {
		descLength = 72
	}

	// Get first line of commit message
	subjectLine := strings.SplitN(subject, "\n", 2)[0]

	// Parse the subject according to conventional commit format
	matches := SubjectRegex.FindStringSubmatch(subjectLine)
	if len(matches) != 5 {
		rule.errors = append(rule.errors, fmt.Errorf("invalid conventional commit format: %q", subjectLine))

		return rule
	}

	// Extract components
	commitType := matches[1]
	scope := matches[2]
	description := matches[4]

	// Validate type
	if len(types) > 0 && !slices.Contains(types, commitType) {
		rule.errors = append(rule.errors, fmt.Errorf("invalid type %q: allowed types are %v", commitType, types))

		return rule
	}

	// Validate scope if provided and scope list is defined
	if scope != "" && len(scopes) > 0 {
		scopesList := strings.Split(scope, ",")
		for _, s := range scopesList {
			if !slices.Contains(scopes, s) {
				rule.errors = append(rule.errors, fmt.Errorf("invalid scope %q: allowed scopes are %v", s, scopes))

				return rule
			}
		}
	}

	// Validate description content
	if strings.TrimSpace(description) == "" {
		rule.errors = append(rule.errors, errors.New("empty description: description must contain non-whitespace characters"))

		return rule
	}

	// Validate description length
	if len(description) > descLength {
		rule.errors = append(rule.errors, fmt.Errorf("description too long: %d characters (max: %d)", len(description), descLength))

		return rule
	}

	// Validate spacing after colon
	if !regexp.MustCompile(`^.*:[ ][^ ].*$`).MatchString(subjectLine) {
		rule.errors = append(rule.errors, errors.New("spacing error: must have exactly one space after colon"))

		return rule
	}

	return rule
}
