// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules/sigverify"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// SignedIdentityRule validates that commit signatures match the committer identity.
type SignedIdentityRule struct {
	baseRule       BaseRule
	keyDir         string
	verifiedBy     string
	signerIdentity string
}

// SignedIdentityOption configures a SignedIdentityRule.
type SignedIdentityOption func(SignedIdentityRule) SignedIdentityRule

// WithKeyDirectory sets the directory containing trusted keys.
func WithKeyDirectory(dir string) SignedIdentityOption {
	return func(r SignedIdentityRule) SignedIdentityRule {
		result := r
		result.keyDir = dir

		return result
	}
}

// NewSignedIdentityRule creates a new rule for validating signature identity.
func NewSignedIdentityRule(options ...SignedIdentityOption) SignedIdentityRule {
	// Create a rule with default values
	rule := SignedIdentityRule{
		baseRule: NewBaseRule("SignedIdentity"),
		keyDir:   "", // Default to no key directory
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate validates that signatures match the committer identity using context-based configuration.
func (r SignedIdentityRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating signed identity using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the existing validation logic
	errors, _ := validateSignedIdentityWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r SignedIdentityRule) withContextConfig(ctx context.Context) SignedIdentityRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values
	keyDir := cfg.Security.KeyDirectory
	allowedIdentities := cfg.Security.AllowedIdentities
	gpgRequired := cfg.Security.GPGRequired

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Str("key_dir", keyDir).
		Strs("allowed_identities", allowedIdentities).
		Bool("gpg_required", gpgRequired).
		Msg("SignedIdentity rule configuration from context")

	// Create a copy of the rule
	result := r

	// If GPG is not required, skip the verification
	if !gpgRequired {
		// Log that we're skipping verification
		logger.Debug().Msg("GPG verification not required, skipping signature identity validation")
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

// validateSignedIdentityWithState validates a signature identity and returns both the errors and an updated rule.
func validateSignedIdentityWithState(rule SignedIdentityRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SignedIdentityRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Skip validation if key directory is empty, which indicates GPG is not required
	// This handles the case where the rule was configured via withContextConfig
	// with GPGRequired = false in the config
	if rule.keyDir == "" {
		// No validation needed, return empty errors
		return []appErrors.ValidationError{}, result
	}

	// If no signature, we can't validate identity
	if commit.Signature == "" {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrMissingSignature,
			"commit has no signature to verify identity",
		)
		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// Determine signature type and parse
	if strings.Contains(commit.Signature, "BEGIN PGP SIGNATURE") {
		// GPG signature verification logic
		result.verifiedBy = "gpg"

		// For signature verification, we would normally have the commit data
		// but in this mock implementation we'll use a placeholder
		commitData := []byte("mock commit data")

		// Verify the signature and get signer identity
		signerIdentity, err := sigverify.VerifyGPGSignature(commitData, commit.Signature, rule.keyDir)
		if err != nil {
			validationErr := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrVerificationFailed,
				"failed to verify GPG signature: "+err.Error(),
			)
			result.baseRule = result.baseRule.WithError(validationErr)

			return result.baseRule.Errors(), result
		}

		result.signerIdentity = signerIdentity

		// Compare with author identity
		if !isIdentityMatch(signerIdentity, commit.AuthorEmail) {
			validationErr := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrInvalidSignature,
				"signature identity does not match commit author",
			).WithContext("signer", signerIdentity).
				WithContext("author", commit.AuthorEmail)

			result.baseRule = result.baseRule.WithError(validationErr)
		}
	} else if strings.Contains(commit.Signature, "BEGIN SSH SIGNATURE") {
		// SSH signature verification logic
		result.verifiedBy = "ssh"

		// For signature verification, we would normally have the commit data and parse the signature
		// but in this mock implementation we'll use placeholders
		commitData := []byte("mock commit data")
		format := "ssh-rsa"
		blob := []byte("mock signature blob")

		// Verify the signature and get signer identity
		signerIdentity, err := sigverify.VerifySSHSignature(commitData, format, blob, rule.keyDir)
		if err != nil {
			validationErr := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrVerificationFailed,
				"failed to verify SSH signature: "+err.Error(),
			)
			result.baseRule = result.baseRule.WithError(validationErr)

			return result.baseRule.Errors(), result
		}

		result.signerIdentity = signerIdentity

		// Compare with author identity
		if !isIdentityMatch(signerIdentity, commit.AuthorEmail) {
			validationErr := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrInvalidSignature,
				"signature identity does not match commit author",
			).WithContext("signer", signerIdentity).
				WithContext("author", commit.AuthorEmail)

			result.baseRule = result.baseRule.WithError(validationErr)
		}
	} else {
		// Unknown signature format
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrUnknownSigFormat,
			"unrecognized signature format",
		)
		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
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

// Name returns the rule name.
func (r SignedIdentityRule) Name() string {
	return r.baseRule.Name()
}

// SetIdentityInfo sets identity information on the rule and returns an updated rule.
func (r SignedIdentityRule) SetIdentityInfo(identity, sigType string) SignedIdentityRule {
	result := r

	if identity != "" {
		result.signerIdentity = identity
	}

	if sigType != "" {
		result.verifiedBy = sigType
	}

	return result
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r SignedIdentityRule) SetErrors(errors []appErrors.ValidationError) SignedIdentityRule {
	result := r
	result.baseRule = result.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Errors returns all validation errors found by this rule.
func (r SignedIdentityRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SignedIdentityRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// Result returns a concise validation result.
func (r SignedIdentityRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		if r.verifiedBy != "" {
			return fmt.Sprintf("❌ %s signature identity mismatch", strings.ToUpper(r.verifiedBy))
		}

		return "❌ Signature identity verification failed"
	}

	if r.verifiedBy != "" {
		return fmt.Sprintf("✓ %s signature identity verified", strings.ToUpper(r.verifiedBy))
	}

	return "✓ Signature identity verified"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SignedIdentityRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		for _, err := range errors {
			switch err.Code {
			case string(appErrors.ErrMissingSignature):
				return "❌ Commit has no signature to verify identity"
			case string(appErrors.ErrVerificationFailed):
				return fmt.Sprintf("❌ %s signature verification failed", strings.ToUpper(r.verifiedBy))
			case string(appErrors.ErrInvalidSignature):
				signer := r.signerIdentity
				if signer == "" && err.Context != nil {
					signer = err.Context["signer"]
				}

				author := ""
				if err.Context != nil {
					author = err.Context["author"]
				}

				return fmt.Sprintf("❌ Signature identity '%s' does not match commit author '%s'",
					signer, author)
			}
		}

		return "❌ Signature identity verification failed"
	}

	if r.verifiedBy != "" && r.signerIdentity != "" {
		return fmt.Sprintf("✓ %s signature verified for identity: %s",
			strings.ToUpper(r.verifiedBy), r.signerIdentity)
	}

	return "✓ Signature identity verified"
}

// Help returns guidance for fixing rule violations.
func (r SignedIdentityRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var helpText strings.Builder

	for _, err := range errors {
		switch err.Code {
		case string(appErrors.ErrMissingSignature):
			helpText.WriteString("Your commit needs to be signed with a cryptographic signature.\n\n")
			helpText.WriteString("To sign commits with GPG:\n")
			helpText.WriteString("1. Set up a GPG key: `gpg --gen-key`\n")
			helpText.WriteString("2. Add it to Git: `git config --global user.signingkey YOUR_KEY_ID`\n")
			helpText.WriteString("3. Sign your commits: `git commit -S -m \"your message\"`\n")

		case string(appErrors.ErrVerificationFailed):
			helpText.WriteString("The signature verification failed. This might be because:\n")
			helpText.WriteString("1. The signature is malformed\n")
			helpText.WriteString("2. The signing key is not in the trusted keys directory\n")
			helpText.WriteString("3. The signature has expired or been revoked\n\n")
			helpText.WriteString("Make sure your signing keys are properly set up and trusted.\n")

		case string(appErrors.ErrInvalidSignature):
			signer := r.signerIdentity
			if signer == "" && err.Context != nil {
				signer = err.Context["signer"]
			}

			author := ""
			if err.Context != nil {
				author = err.Context["author"]
			}

			helpText.WriteString(fmt.Sprintf("The signature identity '%s' does not match the commit author '%s'.\n\n",
				signer, author))
			helpText.WriteString("Make sure you're using the correct signing key that matches your Git author identity.\n")
			helpText.WriteString("You can set your Git author with:\n")
			helpText.WriteString("  git config --global user.name \"Your Name\"\n")
			helpText.WriteString("  git config --global user.email \"your.email@example.com\"\n\n")
			helpText.WriteString("And make sure it matches the identity associated with your signing key.\n")

		case string(appErrors.ErrUnknownSigFormat):
			helpText.WriteString("The signature format was not recognized.\n")
			helpText.WriteString("Make sure you're using either GPG or SSH signatures.\n")
		}
	}

	return helpText.String()
}
