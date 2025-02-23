// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rule

import (
	"fmt"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
)

// DefaultMaxCommitHeaderLength is the default maximum number of characters
// allowed in a commit header.
const DefaultMaxCommitHeaderLength = 89

// HeaderLength enforces a maximum number of characters on the commit header.
type HeaderLength struct {
	headerLength int
	errors       []error
}

// Name returns the name of the check.
func (h *HeaderLength) Name() string {
	return "Header Length"
}

// Message returns the check message.
func (h *HeaderLength) Message() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return fmt.Sprintf("Header is %d characters", h.headerLength)
}

// Errors returns any violations of the check.
func (h *HeaderLength) Errors() []error {
	return h.errors
}

// ValidateHeaderLength checks the header length with proper UTF-8 handling.
func ValidateHeaderLength(message string, maxLength int) interfaces.Rule {
	// Use default max length if not specified
	if maxLength == 0 {
		maxLength = DefaultMaxCommitHeaderLength
	}

	// Compute UTF-8 aware length
	headerLength := utf8.RuneCountInString(message)

	rule := &HeaderLength{
		headerLength: headerLength,
	}

	// Validate length
	if headerLength > maxLength {
		rule.errors = append(rule.errors, fmt.Errorf(
			"commit header is too long: %d characters (maximum allowed: %d)",
			headerLength,
			maxLength,
		))
	}

	return rule
}
