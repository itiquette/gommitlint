// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	domainCrypto "github.com/itiquette/gommitlint/internal/domain/crypto"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// IdentityConfig holds configuration for identity validation.
type IdentityConfig struct {
	EnabledRules     []string
	DisabledRules    []string
	AllowedSigners   []string
	RequireSignature bool
}

// IdentityRule validates that commit signatures match the committer identity.
// Uses domain crypto interfaces for clean architecture.
type IdentityRule struct {
	name       string
	keyDir     string
	verifier   domain.CryptoVerifier
	repository domain.CryptoKeyRepository
	config     IdentityConfig
}

// IdentityOption configures an IdentityRule.
type IdentityOption func(IdentityRule) IdentityRule

// WithVerifier sets the crypto verifier for the rule.
func WithVerifier(verifier domain.CryptoVerifier) IdentityOption {
	return func(r IdentityRule) IdentityRule {
		r.verifier = verifier

		return r
	}
}

// WithKeyRepository sets the key repository for the rule.
func WithKeyRepository(repository domain.CryptoKeyRepository) IdentityOption {
	return func(r IdentityRule) IdentityRule {
		r.repository = repository

		return r
	}
}

// WithConfig sets the configuration for the rule.
func WithConfig(config IdentityConfig) IdentityOption {
	return func(r IdentityRule) IdentityRule {
		r.config = config

		return r
	}
}

// NewIdentityRule creates a new rule for validating signature identity.
// Dependencies must be injected via options.
func NewIdentityRule(options ...IdentityOption) IdentityRule {
	// Create a rule with default values - dependencies injected via options
	rule := IdentityRule{
		name:   "SignedIdentity",
		keyDir: "", // Default to no key directory
		// verifier and repository must be injected
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate validates that signatures match the committer identity.
func (r IdentityRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Use injected configuration
	allowedSigners := r.config.AllowedSigners
	if len(allowedSigners) > 0 {
		// Create author identity
		authorIdentity := domainCrypto.NewIdentity(commit.AuthorName, commit.AuthorEmail)

		// Convert allowed signers config to Identity objects
		var allowedIdentities []domainCrypto.Identity
		for _, allowed := range allowedSigners {
			allowedIdentities = append(allowedIdentities, domainCrypto.NewIdentityFromString(allowed))
		}

		// Check if author is in allowed signers
		isAllowed := false

		for _, allowed := range allowedIdentities {
			if authorIdentity.Matches(allowed) {
				isAllowed = true

				break
			}
		}

		if !isAllowed {
			return []appErrors.ValidationError{
				appErrors.NewIdentityError(
					appErrors.ErrInvalidSignature,
					r.Name(),
					"author identity not in allowed signers list",
					"Add the author to the allowed signers list or use an authorized identity",
				).WithContextMap(map[string]string{
					"author": authorIdentity.String(),
				}),
			}
		}
	}

	// Skip validation if crypto dependencies are not available
	if r.verifier == nil || r.repository == nil {
		return nil
	}

	// Skip validation if key directory is empty
	if r.keyDir == "" && r.repository.GetKeyDirectory() == "" {
		return nil
	}

	// If no signature, we can't validate identity
	if commit.Signature == "" {
		return []appErrors.ValidationError{
			appErrors.NewIdentityError(
				appErrors.ErrMissingSignature,
				"Identity",
				"commit is not signed",
				"Sign commits with: git commit -S",
			),
		}
	}

	// Verify signature using our adapter
	result, err := r.verifier.VerifyCommit(ctx, commit)
	if err != nil {
		return []appErrors.ValidationError{
			appErrors.NewIdentityError(
				appErrors.ErrVerificationFailed,
				"Identity",
				fmt.Sprintf("failed to verify signature: %s", err),
				"Check signature format and key availability",
			),
		}
	}

	// Check verification result
	if !result.IsVerified() {
		return []appErrors.ValidationError{
			appErrors.NewIdentityError(
				appErrors.ErrVerificationFailed,
				"Identity",
				result.ErrorMessage(),
				"Ensure your signing key is valid and properly configured",
			).WithContextMap(map[string]string{
				"error_code": result.ErrorCode(),
			}),
		}
	}

	// Get the verified identity
	signerIdentity := result.Identity()

	// Compare with author identity
	authorIdentity := domainCrypto.NewIdentity(commit.AuthorName, commit.AuthorEmail)
	if !signerIdentity.Matches(authorIdentity) {
		return []appErrors.ValidationError{
			appErrors.NewIdentityError(
				appErrors.ErrInvalidSignature,
				"Identity",
				"signature identity mismatch",
				"Commit author must match signature identity",
			).WithContextMap(map[string]string{
				"signer": signerIdentity.String(),
				"author": authorIdentity.String(),
			}),
		}
	}

	return nil
}

// Name returns the rule name.
func (r IdentityRule) Name() string {
	return r.name
}
