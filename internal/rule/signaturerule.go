// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"github.com/pkg/errors"
)

// SignatureRule ensures that the commit is cryptographically signed.
type SignatureRule struct {
	RuleErrors []error
}

// Name returns the name of the rule.
func (g SignatureRule) Name() string {
	return "SignatureRule"
}

// Result returns validation results.
func (g SignatureRule) Result() string {
	if len(g.RuleErrors) != 0 {
		return g.RuleErrors[0].Error()
	}

	return "SSH/GPG signature found"
}

func (g SignatureRule) Errors() []error {
	return g.RuleErrors
}

func ValidateSignatureRule(signature string) *SignatureRule {
	rule := &SignatureRule{}

	if signature == "" {
		rule.RuleErrors = append(rule.RuleErrors, errors.Errorf("Commit does not have a SSH/GPG-signature"))

		return rule
	}

	return rule
}
