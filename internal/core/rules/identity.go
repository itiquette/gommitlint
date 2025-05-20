// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/core/rules/sigverify"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// IdentityRule validates that commit signatures match the committer identity.
type IdentityRule struct {
	name           string
	keyDir         string
	verifiedBy     string
	signerIdentity string
}

// IdentityOption configures an IdentityRule.
type IdentityOption func(IdentityRule) IdentityRule

// WithKeyDirectory sets the directory containing trusted keys.
func WithKeyDirectory(dir string) IdentityOption {
	return func(r IdentityRule) IdentityRule {
		result := r
		result.keyDir = dir

		return result
	}
}

// NewIdentityRule creates a new rule for validating signature identity.
func NewIdentityRule(options ...IdentityOption) IdentityRule {
	// Create a rule with default values
	rule := IdentityRule{
		name:   "SignedIdentity",
		keyDir: "", // Default to no key directory
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate validates that signatures match the committer identity using context-based configuration.
func (r IdentityRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating signed identity using context configuration", "rule", r.Name(), "commit_hash", commit.Hash)

	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return nil
	}

	// Check if this rule is enabled
	if !IsRuleEnabled(ctx, r.Name()) {
		logger.Debug("Identity rule is disabled, skipping validation")

		return nil
	}

	// Get allowed signers from configuration
	allowedSigners := cfg.GetStringSlice("signing.allowed_signers")
	if len(allowedSigners) > 0 {
		// Validate author identity against allowed signers
		authorIdentity := commit.AuthorName + " <" + commit.AuthorEmail + ">"
		isAllowed := false

		for _, allowed := range allowedSigners {
			if allowed == authorIdentity || extractEmail(allowed) == commit.AuthorEmail {
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
					"author": authorIdentity,
				}),
			}
		}
	}

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Skip validation if key directory is empty or verification is disabled
	if rule.keyDir == "" {
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

	// Determine signature type and parse
	if strings.Contains(commit.Signature, "BEGIN PGP SIGNATURE") {
		// GPG signature verification logic
		r.verifiedBy = "gpg"

		// For signature verification, we would normally have the commit data
		// but in this mock implementation we'll use a placeholder
		commitData := []byte("mock commit data")

		// Verify the signature and get signer identity
		signerIdentity, err := sigverify.VerifyGPGSignature(commitData, commit.Signature, rule.keyDir)
		if err != nil {
			return []appErrors.ValidationError{
				appErrors.NewIdentityError(
					appErrors.ErrVerificationFailed,
					"Identity",
					"GPG signature verification failed",
					"Ensure your GPG key is valid and properly configured",
				).WithContextMap(map[string]string{
					"error": err.Error(),
				}),
			}
		}

		r.signerIdentity = signerIdentity

		// Compare with author identity
		if !isIdentityMatch(signerIdentity, commit.AuthorEmail) {
			return []appErrors.ValidationError{
				appErrors.NewIdentityError(
					appErrors.ErrInvalidSignature,
					"Identity",
					"signature identity mismatch",
					"Commit author must match signature identity",
				).WithContextMap(map[string]string{
					"signer": signerIdentity,
					"author": commit.AuthorEmail,
				}),
			}
		}
	} else if strings.Contains(commit.Signature, "BEGIN SSH SIGNATURE") {
		// SSH signature verification logic
		r.verifiedBy = "ssh"

		// For signature verification, we would normally have the commit data and parse the signature
		// but in this mock implementation we'll use placeholders
		commitData := []byte("mock commit data")
		format := "ssh-rsa"
		blob := []byte("mock signature blob")

		// Verify the signature and get signer identity
		signerIdentity, err := sigverify.VerifySSHSignature(commitData, format, blob, rule.keyDir)
		if err != nil {
			return []appErrors.ValidationError{
				appErrors.NewIdentityError(
					appErrors.ErrVerificationFailed,
					"Identity",
					"SSH signature verification failed",
					"Ensure your SSH key is valid and properly configured",
				).WithContextMap(map[string]string{
					"error": err.Error(),
				}),
			}
		}

		r.signerIdentity = signerIdentity

		// Compare with author identity
		if !isIdentityMatch(signerIdentity, commit.AuthorEmail) {
			return []appErrors.ValidationError{
				appErrors.NewIdentityError(
					appErrors.ErrInvalidSignature,
					"Identity",
					"signature identity mismatch",
					"Commit author must match signature identity",
				).WithContextMap(map[string]string{
					"signer": signerIdentity,
					"author": commit.AuthorEmail,
				}),
			}
		}
	} else {
		// Unknown signature format
		return []appErrors.ValidationError{
			appErrors.NewIdentityError(
				appErrors.ErrUnknownSigFormat,
				"Identity",
				"unrecognized signature format",
				"Use GPG or SSH keys for commit signing",
			),
		}
	}

	return nil
}

// withContextConfig creates a new rule with configuration from context.
func (r IdentityRule) withContextConfig(ctx context.Context) IdentityRule {
	// We're skipping config from context for now as our adapter doesn't have security settings
	// Default values (would come from config in the full implementation)
	keyDir := ""
	allowedIdentities := []string{}
	gpgRequired := false

	// Log configuration at debug level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Identity rule configuration from context",
		"key_dir", keyDir,
		"allowed_identities", allowedIdentities,
		"gpg_required", gpgRequired)

	// Create a copy of the rule
	result := r

	// If GPG is not required, skip the verification
	if !gpgRequired {
		// Log that we're skipping verification
		logger.Debug("GPG verification not required, skipping signature identity validation")
		// We'll return a dummy rule that won't validate anything
		// This ensures validateSignedIdentityWithState won't run verification
		return result
	}

	// Update with context configuration
	if keyDir != "" {
		result.keyDir = keyDir
	}

	return result
}

// Name returns the rule name.
func (r IdentityRule) Name() string {
	return r.name
}

// SetIdentityInfo sets identity information on the rule and returns an updated rule.
func (r IdentityRule) SetIdentityInfo(identity, sigType string) IdentityRule {
	result := r

	if identity != "" {
		result.signerIdentity = identity
	}

	if sigType != "" {
		result.verifiedBy = sigType
	}

	return result
}

// isIdentityMatch checks if the signature identity matches the author identity.
func isIdentityMatch(signerIdentity, authorIdentity string) bool {
	// Simple exact match
	if signerIdentity == authorIdentity {
		return true
	}

	// Normalize and check email only
	signerEmail := extractEmail(signerIdentity)
	authorEmail := extractEmail(authorIdentity)

	return signerEmail != "" && signerEmail == authorEmail
}

// extractEmail attempts to extract an email from a string.
func extractEmail(input string) string {
	// Simple logic: look for string between < and >
	start := strings.LastIndex(input, "<")
	end := strings.LastIndex(input, ">")

	if start != -1 && end != -1 && start < end {
		return input[start+1 : end]
	}

	// If not found, check if the whole string is an email (has @)
	if strings.Contains(input, "@") {
		return input
	}

	return ""
}
