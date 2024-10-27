// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/git"
	"github.com/janderssonse/gommitlint/internal/interfaces"
)

// GPGCheck ensures that the commit is cryptographically signed using GPG.
type GPGCheck struct {
	errors []error
}

// Status returns the name of the check.
func (g GPGCheck) Status() string {
	return "GPG"
}

// Message returns to check message.
func (g GPGCheck) Message() string {
	if len(g.errors) != 0 {
		return g.errors[0].Error()
	}

	return "GPG signature found"
}

// Errors returns any violations of the check.
func (g GPGCheck) Errors() []error {
	return g.errors
}

// ValidateGPGSign checks the commit message for a GPG signature.
func ValidateGPGSign(gitPtr *git.Git) interfaces.Check { //nolint:ireturn
	check := &GPGCheck{}

	isOk, err := gitPtr.HasGPGSignature()
	if err != nil {
		check.errors = append(check.errors, err)

		return check
	}

	if !isOk {
		check.errors = append(check.errors, errors.Errorf("Commit does not have a GPG signature"))

		return check
	}

	return check
}
