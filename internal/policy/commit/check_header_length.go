// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/policy"
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

// Name returns the name of the check.
func (h HeaderLengthCheck) Name() string {
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
func (commit Commit) ValidateHeaderLength() policy.Check { //nolint:ireturn
	check := &HeaderLengthCheck{}

	if commit.Header.Length != 0 {
		MaxNumberOfCommitCharacters = commit.Header.Length
	}

	check.headerLength = len(commit.header())
	if check.headerLength > MaxNumberOfCommitCharacters {
		check.errors = append(check.errors, errors.Errorf("Commit header is %d characters", check.headerLength))
	}

	return check
}
