// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/policy"
)

// HeaderCaseCheck enforces the case of the first word in the header.
type HeaderCaseCheck struct {
	headerCase string
	errors     []error
}

// Name returns the name of the check.
func (h HeaderCaseCheck) Name() string {
	return "Header Case"
}

// Message returns to check message.
func (h HeaderCaseCheck) Message() string {
	if len(h.errors) != 0 {
		return h.errors[0].Error()
	}

	return "Header case is valid"
}

// Errors returns any violations of the check.
func (h HeaderCaseCheck) Errors() []error {
	return h.errors
}

// ValidateHeaderCase checks the header length.
func (commit Commit) ValidateHeaderCase() policy.Check { //nolint:ireturn
	check := &HeaderCaseCheck{headerCase: commit.Header.Case}

	firstWord, err := commit.firstWord()
	if err != nil {
		check.errors = append(check.errors, err)

		return check
	}

	first, _ := utf8.DecodeRuneInString(firstWord)
	if first == utf8.RuneError {
		check.errors = append(check.errors, errors.New("Header does not start with valid UTF-8 text"))

		return check
	}

	var valid bool

	switch commit.Header.Case {
	case "upper":
		valid = unicode.IsUpper(first)
	case "lower":
		valid = unicode.IsLower(first)
	default:
		check.errors = append(check.errors, errors.Errorf("Invalid configured case %s", commit.Header.Case))

		return check
	}

	if !valid {
		check.errors = append(check.errors, errors.Errorf("Commit header case is not %s", commit.Header.Case))
	}

	return check
}
