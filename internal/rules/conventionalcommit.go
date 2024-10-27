// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"
	"strings"

	"github.com/janderssonse/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// HeaderRegex is the regular expression used for Conventional Commits 1.0.0.
var HeaderRegex = regexp.MustCompile(`^(\w*)(\(([^)]+)\))?(!)?:\s{1}(.*)($|\n{2})`)

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

// Status returns the name of the check.
func (c ConventionalCommitCheck) Status() string {
	return "Conventional Commit"
}

// Message returns to check message.
func (c ConventionalCommitCheck) Message() string {
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
func ValidateConventionalCommit(message string, types []string, scopes []string, descLength int) interfaces.Check { //nolint:ireturn
	check := &ConventionalCommitCheck{}
	groups := parseHeader(message)

	if len(groups) != 7 {
		check.errors = append(check.errors, errors.Errorf("Invalid conventional commits format: %q", message))

		return check
	}

	// conventional commit sections
	ccType := groups[1]
	ccScope := groups[3]
	ccDesc := groups[5]

	types = append(types, TypeFeat, TypeFix)
	typeIsValid := false

	for _, t := range types {
		if t == ccType {
			typeIsValid = true
		}
	}

	if !typeIsValid {
		check.errors = append(check.errors, errors.Errorf("Invalid type %q: allowed types are %v", groups[1], types))

		return check
	}

	// Scope is optional.
	if ccScope != "" {
		scopeIsValid := false

		for _, scope := range scopes {
			re := regexp.MustCompile(scope)
			if re.MatchString(ccScope) {
				scopeIsValid = true

				break
			}
		}

		if !scopeIsValid {
			check.errors = append(check.errors, errors.Errorf("Invalid scope %q: allowed scopes are %v", groups[3], scopes))

			return check
		}
	}

	// Provide a good default value for DescriptionLength
	if descLength == 0 {
		descLength = 72
	}

	if len(ccDesc) <= descLength && len(ccDesc) != 0 {
		return check
	}

	check.errors = append(check.errors, errors.Errorf("Invalid description: %s", ccDesc))

	return check
}

func parseHeader(msg string) []string {
	// To circumvent any commit violation due to the leading \n that GitHub
	// prefixes to the commit message on a squash merge, we remove it from the
	// message.
	header := strings.Split(strings.TrimPrefix(msg, "\n"), "\n")[0]
	groups := HeaderRegex.FindStringSubmatch(header)

	return groups
}
