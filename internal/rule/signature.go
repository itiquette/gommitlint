// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"github.com/pkg/errors"

	"github.com/itiquette/gommitlint/internal/interfaces"
)

// Signature ensures that the commit is cryptographically signed using GPG.
type Signature struct {
	RuleErrors []error
}

// Name returns the name of the check.
func (g Signature) Name() string {
	return "Signature"
}

// Result returns to check message.
func (g Signature) Result() string {
	if len(g.RuleErrors) != 0 {
		return g.RuleErrors[0].Error()
	}

	return "SSH/GPG signature found"
}

func (g Signature) Errors() []error {
	return g.RuleErrors
}

// ValidateSignature checks the commit message for a GPG signature.
func ValidateSignature(signature string) interfaces.CommitRule { //nolint:ireturn
	rule := &Signature{}

	if signature == "" {
		rule.RuleErrors = append(rule.RuleErrors, errors.Errorf("Commit does not have a SSH/GPG-signature"))

		return rule
	}

	return rule
}
