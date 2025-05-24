// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package rules provides test helpers for testing rule components.
package rules

import (
	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
)

// WithTestKeyDirectory returns an IdentityOption that sets the key directory.
// This is the recommended replacement for the removed WithKeyDirectory.
func WithTestKeyDirectory(dir string) rules.IdentityOption {
	// Create a new repository and verifier with the specified directory
	repo := crypto.NewFileSystemKeyRepository(dir)
	verifier := crypto.NewVerificationAdapter(repo)

	// Return a composite option that applies both repository and verifier
	return func(rule rules.IdentityRule) rules.IdentityRule {
		// Apply repository
		rule = rules.WithKeyRepository(repo)(rule)
		// Apply verifier
		rule = rules.WithVerifier(verifier)(rule)

		return rule
	}
}

// WithTestRepository returns an IdentityOption that sets the repository.
func WithTestRepository(repo domain.CryptoKeyRepository) rules.IdentityOption {
	return rules.WithKeyRepository(repo)
}

// WithTestVerifier returns an IdentityOption that sets the verifier.
func WithTestVerifier(verifier domain.CryptoVerifier) rules.IdentityOption {
	return rules.WithVerifier(verifier)
}
