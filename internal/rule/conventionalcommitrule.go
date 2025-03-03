// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

var SubjectRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// ConventionalCommitRule ensures that the commit message is a valid
// conventional commit.
type ConventionalCommitRule struct {
	errors []error
}

// Name returns the name of the rule.
func (c ConventionalCommitRule) Name() string {
	return "ConventionalCommitRule"
}

// Result returns the validation results.
func (c ConventionalCommitRule) Result() string {
	if len(c.errors) != 0 {
		return c.errors[0].Error()
	}

	return "Commit message is a valid conventional commit"
}

// Errors returns validation errors.
func (c ConventionalCommitRule) Errors() []error {
	return c.errors
}

func ValidateConventionalCommit(subject string, types []string, scopes []string, descLength int) *ConventionalCommitRule {
	rule := &ConventionalCommitRule{}
	subjectGroups := parseSubject(subject)

	if len(subjectGroups) != 5 {
		rule.errors = append(rule.errors, errors.Errorf("Invalid conventional commits format: %q", subject))

		return rule
	}

	// [0] - Full match (entire commit message subject)
	// [1] - Type (feat, fix, etc.)
	// [2] - Scope (without parentheses)
	// [3] - Breaking change marker (!)
	// [4] - Description
	// conventional commit sections
	ccFull := subjectGroups[0]
	ccType := subjectGroups[1]
	ccScope := subjectGroups[2]
	ccDesc := subjectGroups[4]

	isValidType := slices.Contains(types, ccType)

	if !isValidType {
		rule.errors = append(rule.errors, errors.Errorf("Invalid type %q: allowed types are %v", ccType, types))

		return rule
	}

	// Scope is optional.
	if ccScope != "" && len(scopes) > 0 {
		ccScopes := strings.Split(ccScope, ",")
		for _, scope := range ccScopes {
			isValidScope := slices.Contains(scopes, scope)

			if !isValidScope {
				rule.errors = append(rule.errors, errors.Errorf("Invalid scope %q: allowed scopes are %v", scope, scopes))

				return rule
			}
		}
	}

	// Description is not optional, neither should be only whitespace
	if strings.TrimSpace(ccDesc) == "" {
		rule.errors = append(rule.errors, errors.Errorf("Invalid description %q: description must be at least one non whitespace char", subjectGroups[4]))

		return rule
	}

	var OneSpaceRegex = regexp.MustCompile(`^.*:[ ][^ ].*$`)

	if !OneSpaceRegex.MatchString(ccFull) {
		rule.errors = append(rule.errors, errors.Errorf("Space between type: description %q must be one", subjectGroups[0]))

		return rule
	}

	// Provide a good default value for DescriptionLength
	if descLength == 0 {
		descLength = 72
	}

	if len(ccDesc) <= descLength && len(ccDesc) != 0 {
		return rule
	}

	rule.errors = append(rule.errors, errors.Errorf("Invalid description: %s", ccDesc))

	return rule
}

func parseSubject(msg string) []string {
	subject := strings.Split(msg, "\n")[0]
	groups := SubjectRegex.FindStringSubmatch(subject)

	return groups
}
