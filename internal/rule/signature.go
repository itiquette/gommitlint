// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rule

import (
	"github.com/pkg/errors"

	"github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/interfaces"
)

// Signature ensures that the commit is cryptographically signed using GPG.
type Signature struct {
	RuleErrors []error
}

// Name returns the name of the check.
func (g Signature) Name() string {
	return "GPG"
}

// Message returns to check message.
func (g Signature) Message() string {
	if len(g.RuleErrors) != 0 {
		return g.RuleErrors[0].Error()
	}

	return "GPG signature found"
}

func (g Signature) Errors() []error {
	return g.RuleErrors
}

// ValidateSignature checks the commit message for a GPG signature.
func ValidateSignature(gitPtr *git.Git) interfaces.Rule { //nolint:ireturn
	rule := &Signature{}

	isOk, err := gitPtr.HasGPGSignature()
	if err != nil {
		rule.RuleErrors = append(rule.RuleErrors, err)

		return rule
	}

	if !isOk {
		rule.RuleErrors = append(rule.RuleErrors, errors.Errorf("Commit does not have a GPG signature"))

		return rule
	}

	return rule
}
