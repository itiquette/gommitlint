// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rules

import (
	"fmt"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
)

// DefaultMaxCommitHeaderLength is the default maximum number of characters
// allowed in a commit header.
const DefaultMaxCommitHeaderLength = 89

// HeaderLengthCheck enforces a maximum number of characters on the commit header.
type HeaderLengthCheck struct {
	headerLength int
	errors       []error
}

// Status returns the name of the check.
func (h *HeaderLengthCheck) Status() string {
	return "Header Length"
}

// Message returns the check message.
func (h *HeaderLengthCheck) Message() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return fmt.Sprintf("Header is %d characters", h.headerLength)
}

// Errors returns any violations of the check.
func (h *HeaderLengthCheck) Errors() []error {
	return h.errors
}

// ValidateHeaderLength checks the header length with proper UTF-8 handling.
func ValidateHeaderLength(message string, maxLength int) interfaces.Check {
	// Use default max length if not specified
	if maxLength == 0 {
		maxLength = DefaultMaxCommitHeaderLength
	}

	// Compute UTF-8 aware length
	headerLength := utf8.RuneCountInString(message)

	check := &HeaderLengthCheck{
		headerLength: headerLength,
	}

	// Validate length
	if headerLength > maxLength {
		check.errors = append(check.errors, fmt.Errorf(
			"commit header is too long: %d characters (maximum allowed: %d)",
			headerLength,
			maxLength,
		))
	}

	return check
}
