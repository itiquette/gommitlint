// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
)

// IdentityRule validates that commit signatures match the committer identity.
// Uses domain crypto interfaces for clean architecture.
type IdentityRule struct {
	name           string
	keyDir         string
	verifier       domain.CryptoVerifier
	repository     domain.CryptoKeyRepository
	allowedSigners []string
}

// NewIdentityRule creates a new rule for validating signature identity from config.
func NewIdentityRule(cfg config.Config, deps domain.RuleDependencies) IdentityRule {
	return IdentityRule{
		name:           "SignedIdentity",
		keyDir:         cfg.Signing.KeyDirectory,
		verifier:       deps.CryptoVerifier,
		repository:     deps.CryptoRepository,
		allowedSigners: cfg.Signing.AllowedSigners,
	}
}

// Validate validates that signatures match the committer identity.
func (r IdentityRule) Validate(ctx context.Context, commit domain.CommitInfo) []domain.ValidationError {
	// Use allowed signers from config
	if len(r.allowedSigners) > 0 {
		// Create author identity
		authorIdentity := domain.NewIdentity(commit.AuthorName, commit.AuthorEmail)

		// Convert allowed signers config to Identity objects using functional approach
		allowedIdentities := domain.MapSliceCompat(r.allowedSigners, domain.NewIdentityFromString)

		// Check if author is in allowed signers
		isAllowed := false

		for _, allowed := range allowedIdentities {
			if authorIdentity.Matches(allowed) {
				isAllowed = true

				break
			}
		}

		if !isAllowed {
			return []domain.ValidationError{
				domain.New(
					r.Name(),
					domain.ErrInvalidSignature,
					"author identity not in allowed signers list",
				).WithHelp("Add the author to the allowed signers list or use an authorized identity").WithContextMap(map[string]string{
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
		return []domain.ValidationError{
			domain.New(
				"Identity",
				domain.ErrMissingSignature,
				"commit is not signed",
			).WithHelp("Sign commits with: git commit -S"),
		}
	}

	// Verify signature using our adapter
	result, err := r.verifier.VerifyCommit(ctx, commit)
	if err != nil {
		return []domain.ValidationError{
			domain.New(
				"Identity",
				domain.ErrVerificationFailed,
				fmt.Sprintf("failed to verify signature: %s", err),
			).WithHelp("Check signature format and key availability"),
		}
	}

	// Check verification result
	if !result.IsVerified() {
		return []domain.ValidationError{
			domain.New(
				"Identity",
				domain.ErrVerificationFailed,
				result.ErrorMessage(),
			).WithHelp("Ensure your signing key is valid and properly configured").WithContextMap(map[string]string{
				"error_code": result.ErrorCode(),
			}),
		}
	}

	// Get the verified identity
	signerIdentity := result.Identity()

	// Compare with author identity
	authorIdentity := domain.NewIdentity(commit.AuthorName, commit.AuthorEmail)
	if !signerIdentity.Matches(authorIdentity) {
		return []domain.ValidationError{
			domain.New(
				"Identity",
				domain.ErrInvalidSignature,
				"signature identity mismatch",
			).WithHelp("Commit author must match signature identity").WithContextMap(map[string]string{
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
