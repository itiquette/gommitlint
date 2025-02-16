// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rules

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// HeaderSuffixCheck enforces that the last character of the header isn't in a specified set.
type HeaderSuffixCheck struct {
	errors []error
}

// Status returns the name of the check.
func (h *HeaderSuffixCheck) Status() string {
	return "Header Last Character"
}

// Message returns the check message.
func (h *HeaderSuffixCheck) Message() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return "Header last character is valid"
}

// Errors returns any violations of the check.
func (h *HeaderSuffixCheck) Errors() []error {
	return h.errors
}

// ValidateHeaderSuffix checks the last character of the header.
func ValidateHeaderSuffix(header, invalidSuffixes string) interfaces.Check {
	check := &HeaderSuffixCheck{}

	// Handle empty header
	if header == "" {
		return check
	}

	// Decode the last rune
	last, _ := utf8.DecodeLastRuneInString(header)

	// Check for invalid UTF-8
	if last == utf8.RuneError {
		check.errors = append(check.errors, errors.New("header does not end with valid UTF-8 text"))

		return check
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(invalidSuffixes, last) {
		check.errors = append(check.errors, fmt.Errorf("commit header ends with invalid character %q", last))
	}

	return check
}
