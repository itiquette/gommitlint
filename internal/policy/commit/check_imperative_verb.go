// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"strings"

	"github.com/jdkato/prose/v3"
	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/policy"
)

// ImperativeCheck enforces that the first word of a commit message header is
// and imperative verb.
type ImperativeCheck struct {
	errors []error
}

// Name returns the name of the check.
func (i ImperativeCheck) Name() string {
	return "Imperative Mood"
}

// Message returns to check message.
func (i ImperativeCheck) Message() string {
	if len(i.errors) != 0 {
		return i.errors[0].Error()
	}

	return "Commit begins with imperative verb"
}

// Errors returns any violations of the check.
func (i ImperativeCheck) Errors() []error {
	return i.errors
}

// ValidateImperative checks the commit message for a GPG signature.
func (commit Commit) ValidateImperative() policy.Check { //nolint:ireturn
	check := &ImperativeCheck{}

	var (
		word string
		err  error
	)

	if word, err = commit.firstWord(); err != nil {
		check.errors = append(check.errors, err)

		return check
	}

	doc, err := prose.NewDocument("I " + strings.ToLower(word))
	if err != nil {
		check.errors = append(check.errors, errors.Errorf("Failed to create document: %v", err))

		return check
	}

	if len(doc.Tokens()) != 2 {
		check.errors = append(check.errors, errors.Errorf("Expected 2 tokens, got %d", len(doc.Tokens())))

		return check
	}

	tokens := doc.Tokens()
	tok := tokens[1]

	for _, tag := range []string{"VBD", "VBG", "VBZ"} {
		if tok.Tag == tag {
			check.errors = append(check.errors, errors.Errorf("First word of commit must be an imperative verb: %q is invalid", word))
		}
	}

	return check
}
