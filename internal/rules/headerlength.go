// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"fmt"

	"github.com/janderssonse/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// MaxNumberOfCommitCharacters is the default maximium number of characters
// allowed in a commit header.
var MaxNumberOfCommitCharacters = 89

// HeaderLengthCheck enforces a maximum number of charcters on the commit
// header.
type HeaderLengthCheck struct {
	headerLength int
	errors       []error
}

// Status returns the name of the check.
func (h HeaderLengthCheck) Status() string {
	return "Header Length"
}

// Message returns to check message.
func (h HeaderLengthCheck) Message() string {
	return fmt.Sprintf("Header is %d characters", h.headerLength)
}

// Errors returns any violations of the check.
func (h HeaderLengthCheck) Errors() []error {
	return h.errors
}

// ValidateHeaderLength checks the header length.
func ValidateHeaderLength(ruleLength int, actualLength int) interfaces.Check { //nolint:ireturn
	check := &HeaderLengthCheck{}

	if ruleLength != 0 {
		MaxNumberOfCommitCharacters = actualLength
	}

	check.headerLength = actualLength
	if check.headerLength > MaxNumberOfCommitCharacters {
		check.errors = append(check.errors, errors.Errorf("Commit header is %d characters", check.headerLength))
	}

	return check
}
