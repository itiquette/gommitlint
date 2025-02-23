// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// HeaderSuffix enforces that the last character of the header isn't in a specified set.
type HeaderSuffix struct {
	errors []error
}

// Name returns the name of the rule.
func (h *HeaderSuffix) Name() string {
	return "Header Last Character"
}

// Message returns the check message.
func (h *HeaderSuffix) Message() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return "Header last character is valid"
}

// Errors returns any violations of the check.
func (h *HeaderSuffix) Errors() []error {
	return h.errors
}

// ValidateHeaderSuffix checks the last character of the header.
func ValidateHeaderSuffix(header, invalidSuffixes string) interfaces.Rule {
	rule := &HeaderSuffix{}

	// Handle empty header
	if header == "" {
		return rule
	}

	// Decode the last rune
	last, _ := utf8.DecodeLastRuneInString(header)

	// Check for invalid UTF-8
	if last == utf8.RuneError {
		rule.errors = append(rule.errors, errors.New("header does not end with valid UTF-8 text"))

		return rule
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(invalidSuffixes, last) {
		rule.errors = append(rule.errors, fmt.Errorf("commit header ends with invalid character %q", last))
	}

	return rule
}
