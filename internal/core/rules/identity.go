// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/crypto"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	domainCrypto "github.com/itiquette/gommitlint/internal/domain/crypto"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// IdentityRule validates that commit signatures match the committer identity.
// Uses the crypto domain for identity and signature handling.
type IdentityRule struct {
	name       string
	keyDir     string
	verifier   *crypto.VerificationAdapter
	repository crypto.KeyRepository
}

// IdentityOption configures an IdentityRule.
type IdentityOption func(IdentityRule) IdentityRule

// WithKeyDirectory sets the directory containing trusted keys.
// This is deprecated and will be removed in a future version.
// Use testutils/rules.WithTestKeyDirectory for tests.
// Production code should use configuration-based paths.
func WithKeyDirectory(dir string) IdentityOption {
	return func(r IdentityRule) IdentityRule {
		result := r
		result.keyDir = dir

		// Update repository if needed
		if result.repository != nil {
			result.repository = crypto.NewFileSystemKeyRepository(dir)
			result.verifier = crypto.NewVerificationAdapter(result.repository)
		}

		return result
	}
}

// GetRepository returns the key repository. Used in test helpers only.
// Note: Previously exported but moved to testutils/rules for proper separation.
func (r IdentityRule) GetRepository() crypto.KeyRepository {
	return r.repository
}

// SetRepository sets the key repository. Used in test helpers only.
// Note: Previously exported but moved to testutils/rules for proper separation.
func (r *IdentityRule) SetRepository(repo crypto.KeyRepository) {
	r.repository = repo
}

// SetVerifier sets the verification adapter. Used in test helpers only.
// Note: Previously exported but moved to testutils/rules for proper separation.
func (r *IdentityRule) SetVerifier(verifier *crypto.VerificationAdapter) {
	r.verifier = verifier
}

// NewIdentityRule creates a new rule for validating signature identity.
func NewIdentityRule(options ...IdentityOption) IdentityRule {
	// Create default repository and verifier
	repository := crypto.NewFileSystemKeyRepository("") // Empty directory as default
	verifier := crypto.NewVerificationAdapter(repository)

	// Create a rule with default values
	rule := IdentityRule{
		name:       "SignedIdentity",
		keyDir:     "", // Default to no key directory
		repository: repository,
		verifier:   verifier,
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// WithContext implements the ConfigurableRule interface for IdentityRule.
// It returns a new rule with configuration from the provided context.
func (r IdentityRule) WithContext(ctx context.Context) domain.Rule {
	// Get configuration directly from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r
	}

	// Create a copy of the rule
	result := r

	// Get keyDir from configuration if available
	keyDir := cfg.GetString("signing.key_directory")
	if keyDir != "" && keyDir != r.keyDir {
		result.keyDir = keyDir
		// Update repository and verifier
		result.repository = crypto.NewFileSystemKeyRepository(keyDir)
		result.verifier = crypto.NewVerificationAdapter(result.repository)
	}

	return result
}

// Validate validates that signatures match the committer identity.
func (r IdentityRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return nil
	}

	// Check if rule is enabled (use domain rule registry)
	priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())
	enabled := cfg.GetStringSlice("rules.enabled")
	disabled := cfg.GetStringSlice("rules.disabled")

	if !priorityService.IsRuleEnabled(ctx, r.Name(), enabled, disabled) {
		return nil
	}

	// Get allowed signers from configuration
	allowedSigners := cfg.GetStringSlice("signing.allowed_signers")
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
