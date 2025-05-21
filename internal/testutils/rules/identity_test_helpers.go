// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules provides test helpers for testing rule components.
package rules

import (
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/core/rules"
)

// GetRuleRepository is a test-only function to access the repository from an IdentityRule.
// This should only be used in test code.
func GetRuleRepository(r rules.IdentityRule) crypto.KeyRepository {
	return r.GetRepository()
}

// SetRuleRepository is a test-only function to set the repository in an IdentityRule.
// This should only be used in test code.
func SetRuleRepository(r *rules.IdentityRule, repo crypto.KeyRepository) {
	r.SetRepository(repo)
}

// SetRuleVerifier is a test-only function to set the verifier in an IdentityRule.
// This should only be used in test code.
func SetRuleVerifier(r *rules.IdentityRule, verifier *crypto.VerificationAdapter) {
	r.SetVerifier(verifier)
}
