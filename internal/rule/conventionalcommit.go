// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"regexp"
	"slices"
	"strings"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

var SubjectRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

const (
	// TypeFeat is a commit of the type fix patches a bug in your codebase
	// (this correlates with MINOR in semantic versioning).
	TypeFeat = "feat"

	// TypeFix is a commit of the type feat introduces a new feature to the
	// codebase (this correlates with PATCH in semantic versioning).
	TypeFix = "fix"
)

// ConventionalCommitCheck ensures that the commit message is a valid
// conventional commit.
type ConventionalCommitCheck struct {
	errors []error
}

// Name returns the name of the check.
func (c ConventionalCommitCheck) Name() string {
	return "Conventional Commit"
}

// Result returns to check message.
func (c ConventionalCommitCheck) Result() string {
	if len(c.errors) != 0 {
		return c.errors[0].Error()
	}

	return "Commit message is a valid conventional commit"
}

// Errors returns any violations of the check.
func (c ConventionalCommitCheck) Errors() []error {
	return c.errors
}

// ValidateConventionalCommit returns the commit type.
func ValidateConventionalCommit(subject string, types []string, scopes []string, descLength int) interfaces.CommitRule { //nolint:ireturn
	rule := &ConventionalCommitCheck{}
	groups := parseSubject(subject)

	if len(groups) != 5 {
		rule.errors = append(rule.errors, errors.Errorf("Invalid conventional commits format: %q", subject))

		return rule
	}

	// [0] - Full match (entire commit message subject)
	// [1] - Type (feat, fix, etc.)
	// [2] - Scope (without parentheses)
	// [3] - Breaking change marker (!)
	// [4] - Description
	// conventional commit sections
	ccFull := groups[0]
	ccType := groups[1]
	ccScope := groups[2]
	ccDesc := groups[4]

	isValidType := slices.Contains(types, ccType)

	if !isValidType {
		rule.errors = append(rule.errors, errors.Errorf("Invalid type %q: allowed types are %v", groups[1], types))

		return rule
	}

	// Scope is optional.
	if ccScope != "" && len(scopes) > 0 {
		ccScopes := strings.Split(ccScope, ",")
		for _, scope := range ccScopes {
			isValidScope := slices.Contains(scopes, scope)

			if !isValidScope {
				rule.errors = append(rule.errors, errors.Errorf("Invalid scope %q: allowed scopes are %v", groups[3], scopes))

				return rule
			}
		}
	}

	// Description is not optional, neither should be only whitespace
	if strings.TrimSpace(ccDesc) == "" {
		rule.errors = append(rule.errors, errors.Errorf("Invalid description %q: description must be at least one non whitespace char", groups[4]))

		return rule
	}

	var OneSpaceRegex = regexp.MustCompile(`^.*:[ ][^ ].*$`)

	if !OneSpaceRegex.MatchString(ccFull) {
		rule.errors = append(rule.errors, errors.Errorf("Space between type: description %q must be one", groups[0]))

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
