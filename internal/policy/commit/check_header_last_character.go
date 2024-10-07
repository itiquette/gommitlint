// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/policy"
)

// HeaderLastCharacterCheck enforces that the last character of the header isn't in some set.
type HeaderLastCharacterCheck struct {
	errors []error
}

// Name returns the name of the check.
func (h HeaderLastCharacterCheck) Name() string {
	return "Header Last Character"
}

// Message returns to check message.
func (h HeaderLastCharacterCheck) Message() string {
	if len(h.errors) != 0 {
		return h.errors[0].Error()
	}

	return "Header last character is valid"
}

// Errors returns any violations of the check.
func (h HeaderLastCharacterCheck) Errors() []error {
	return h.errors
}

// ValidateHeaderLastCharacter checks the last character of the header.
func (commit Commit) ValidateHeaderLastCharacter() policy.Check { //nolint:ireturn
	check := &HeaderLastCharacterCheck{}

	switch last, _ := utf8.DecodeLastRuneInString(commit.header()); {
	case last == utf8.RuneError:
		check.errors = append(check.errors, errors.New("Header does not end with valid UTF-8 text"))
	case strings.ContainsRune(commit.Header.InvalidLastCharacters, last):
		check.errors = append(check.errors, errors.Errorf("Commit header ends in %q", last))
	}

	return check
}
