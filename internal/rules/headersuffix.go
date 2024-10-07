// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"
	"unicode/utf8"

	"github.com/janderssonse/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// HeaderSuffixCheck enforces that the last character of the header isn't in some set.
type HeaderSuffixCheck struct {
	errors []error
}

// Status returns the name of the check.
func (h HeaderSuffixCheck) Status() string {
	return "Header Last Character"
}

// Message returns to check message.
func (h HeaderSuffixCheck) Message() string {
	if len(h.errors) != 0 {
		return h.errors[0].Error()
	}

	return "Header last character is valid"
}

// Errors returns any violations of the check.
func (h HeaderSuffixCheck) Errors() []error {
	return h.errors
}

// ValidateHeaderSuffix checks the last character of the header.
func ValidateHeaderSuffix(header string, invalidSuffix string) interfaces.Check { //nolint:ireturn
	check := &HeaderSuffixCheck{}

	switch last, _ := utf8.DecodeLastRuneInString(header); {
	case last == utf8.RuneError:
		check.errors = append(check.errors, errors.New("Header does not end with valid UTF-8 text"))
	case strings.ContainsRune(invalidSuffix, last):
		check.errors = append(check.errors, errors.Errorf("Commit header ends in %q", last))
	}

	return check
}
