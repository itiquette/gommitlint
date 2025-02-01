// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"unicode"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// HeaderCaseCheck enforces the case of the first word in the header.
type HeaderCaseCheck struct {
	headerCase string
	errors     []error
}

// Status returns the name of the check.
func (h HeaderCaseCheck) Status() string {
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
func ValidateHeaderCase(isConventional bool, message string, caseChoice string) interfaces.Check { //nolint:ireturn
	check := &HeaderCaseCheck{headerCase: caseChoice}

	firstWord, err := firstWord(isConventional, message)
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

	switch caseChoice {
	case "upper":
		valid = unicode.IsUpper(first)
	case "lower":
		valid = unicode.IsLower(first)
	default:
		check.errors = append(check.errors, errors.Errorf("Invalid configured case %s", caseChoice))

		return check
	}

	if !valid {
		check.errors = append(check.errors, errors.Errorf("Commit header case is not %s", caseChoice))
	}

	return check
}
