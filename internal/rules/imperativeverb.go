// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"regexp"
	"strings"

	"github.com/janderssonse/gommitlint/internal/interfaces"
	"github.com/jdkato/prose/v3"
	"github.com/pkg/errors"
)

// ImperativeCheck enforces that the first word of a commit message header is
// and imperative verb.
type ImperativeCheck struct {
	errors []error
}

// Status returns the name of the check.
func (i ImperativeCheck) Status() string {
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

// ValidateImperative checks the commit message for a imperative first word.
func ValidateImperative(isConventional bool, message string) interfaces.Check { //nolint:ireturn
	check := &ImperativeCheck{}

	var (
		word string
		err  error
	)

	if word, err = firstWord(isConventional, message); err != nil {
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

func firstWord(isConventional bool, message string) (string, error) {
	var (
		groups []string
		msg    string
	)

	if isConventional {
		groups = parseHeader(message)
		if len(groups) != 7 {
			return "", errors.Errorf("Invalid conventional commit format")
		}

		msg = groups[5]
	} else {
		msg = message
	}

	if msg == "" {
		return "", errors.Errorf("Invalid msg: %s", msg)
	}

	if groups = FirstWordRegex.FindStringSubmatch(msg); groups == nil {
		return "", errors.Errorf("Invalid msg: %s", msg)
	}

	return groups[0], nil
}

// FirstWordRegex is theregular expression used to find the first word in a
// commit.
var FirstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)
