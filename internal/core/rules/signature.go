// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// SignatureRule validates that commits have cryptographic signatures.
type SignatureRule struct {
	baseRule           BaseRule
	requireSignature   bool
	allowedSigTypes    []string
	foundSignatureType string
}

// SignatureOption configures a SignatureRule.
type SignatureOption func(SignatureRule) SignatureRule

// WithRequireSignature configures whether signatures are required.
func WithRequireSignature(required bool) SignatureOption {
	return func(r SignatureRule) SignatureRule {
		result := r
		result.requireSignature = required

		return result
	}
}

// WithAllowedSignatureTypes configures the allowed signature types.
func WithAllowedSignatureTypes(types []string) SignatureOption {
	return func(r SignatureRule) SignatureRule {
		result := r
		// Create a deep copy of the types slice
		if len(types) > 0 {
			result.allowedSigTypes = make([]string, len(types))
			copy(result.allowedSigTypes, types)
		} else {
			result.allowedSigTypes = []string{}
		}

		return result
	}
}

// NewSignatureRule creates a new rule for validating commit signatures.
func NewSignatureRule(options ...SignatureOption) SignatureRule {
	// Create a rule with default values
	rule := SignatureRule{
		baseRule:         NewBaseRule("Signature"),
		requireSignature: true,
		allowedSigTypes:  []string{"gpg", "ssh"},
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate checks for the presence and format of cryptographic signatures
// using configuration from context, but options provided to the rule constructor take precedence.
func (r SignatureRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating signature using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the existing validation logic
	errors, _ := validateSignatureWithState(rule, commit)

	return errors
}

// Note: Rule constructor options take precedence over context configuration.
func (r SignatureRule) withContextConfig(ctx context.Context) SignatureRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Create a copy of the rule
	result := r

	// Update requireSignature from the context if not explicitly set via options
	// This ensures GPGRequired from the config is properly applied
	if cfg.Security.GPGRequired {
		result.requireSignature = true
	}

	// Only update settings from context if not explicitly set via options
	// Extract configuration values
	// NOTE: Rule constructor options should always take precedence over context config
	if len(cfg.Security.AllowedSignatureTypes) > 0 && len(result.allowedSigTypes) == 0 {
		result.allowedSigTypes = make([]string, len(cfg.Security.AllowedSignatureTypes))
		copy(result.allowedSigTypes, cfg.Security.AllowedSignatureTypes)
	}

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Bool("require_signature", result.requireSignature).
		Strs("allowed_sig_types", result.allowedSigTypes).
		Msg("Signature rule configuration (final)")

	return result
}

// validateSignatureWithState validates a commit signature and returns both the errors and an updated rule.
func validateSignatureWithState(rule SignatureRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SignatureRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Check if signature is required
	if !rule.requireSignature {
		return result.baseRule.Errors(), result
	}

	// Check if signature exists
	if commit.Signature == "" {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrMissingSignature,
			"Commit is missing a cryptographic signature",
		)
		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// Determine signature type
	sigType := determineSignatureType(commit.Signature)
	result.foundSignatureType = sigType

	// Validate signature type if allowed types are specified
	if len(rule.allowedSigTypes) > 0 {
		isAllowed := false

		for _, allowedType := range rule.allowedSigTypes {
			if strings.EqualFold(sigType, allowedType) {
				isAllowed = true

				break
			}
		}

		if !isAllowed {
			validationErr := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrDisallowedSigType,
				fmt.Sprintf("Signature type '%s' is not allowed", sigType),
			).WithContext("found_type", sigType).
				WithContext("allowed_types", strings.Join(rule.allowedSigTypes, ", "))
			result.baseRule = result.baseRule.WithError(validationErr)
		}
	}

	return result.baseRule.Errors(), result
}

// determineSignatureType determines the type of signature from its content.
func determineSignatureType(signature string) string {
	signature = strings.TrimSpace(signature)

	// Check for GPG signature
	if strings.HasPrefix(signature, "-----BEGIN PGP SIGNATURE-----") {
		return "gpg"
	}

	// Check for SSH signature
	if strings.Contains(signature, "-----BEGIN SSH SIGNATURE-----") {
		return "ssh"
	}

	// Unknown signature type
	return "unknown"
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return r.baseRule.Name()
}

// Errors returns all validation errors found by this rule.
func (r SignatureRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SignatureRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// Result returns a concise validation result.
func (r SignatureRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		if containsError(errors, appErrors.ErrMissingSignature) {
			return "❌ No signature"
		}

		if containsError(errors, appErrors.ErrDisallowedSigType) {
			return "❌ Disallowed signature type: " + r.foundSignatureType
		}

		return "❌ Invalid signature"
	}

	if r.foundSignatureType != "" {
		return fmt.Sprintf("✓ Valid %s signature", strings.ToUpper(r.foundSignatureType))
	}

	return "✓ Valid signature"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SignatureRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		if containsError(errors, appErrors.ErrMissingSignature) {
			return "❌ Commit is missing a cryptographic signature"
		}

		if containsError(errors, appErrors.ErrDisallowedSigType) {
			allowedTypes := strings.Join(r.allowedSigTypes, ", ")

			return fmt.Sprintf("❌ Signature type '%s' is not allowed (allowed: %s)",
				r.foundSignatureType, allowedTypes)
		}

		return "❌ Signature validation failed"
	}

	if r.foundSignatureType != "" {
		return fmt.Sprintf("✓ Commit has a valid %s signature", strings.ToUpper(r.foundSignatureType))
	}

	return "✓ Commit has a valid signature"
}

// Help returns guidance for fixing rule violations.
func (r SignatureRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var helpText strings.Builder

	if containsError(errors, appErrors.ErrMissingSignature) {
		helpText.WriteString("Your commit needs to be signed with a cryptographic signature.\n\n")
		helpText.WriteString("To sign commits with GPG:\n")
		helpText.WriteString("1. Set up a GPG key: `gpg --gen-key`\n")
		helpText.WriteString("2. Add it to Git: `git config --global user.signingkey YOUR_KEY_ID`\n")
		helpText.WriteString("3. Sign your commits: `git commit -S -m \"your message\"`\n")
		helpText.WriteString("\nTo sign commits with SSH:\n")
		helpText.WriteString("1. Configure Git to use your SSH key: `git config --global commit.gpgsign true`\n")
		helpText.WriteString("2. Set signing format: `git config --global gpg.format ssh`\n")
		helpText.WriteString("3. Specify your key: `git config --global user.signingkey ~/.ssh/id_ed25519.pub`\n")
	} else if containsError(errors, appErrors.ErrDisallowedSigType) {
		helpText.WriteString(fmt.Sprintf("Your commit is signed with '%s', but this signature type is not allowed.\n\n",
			r.foundSignatureType))
		helpText.WriteString(fmt.Sprintf("Allowed signature types: %s\n",
			strings.Join(r.allowedSigTypes, ", ")))

		for _, allowedType := range r.allowedSigTypes {
			if strings.EqualFold(allowedType, "gpg") {
				helpText.WriteString("\nTo sign with GPG:\n")
				helpText.WriteString("1. Set up a GPG key: `gpg --gen-key`\n")
				helpText.WriteString("2. Add it to Git: `git config --global user.signingkey YOUR_KEY_ID`\n")
				helpText.WriteString("3. Sign your commits: `git commit -S -m \"your message\"`\n")
			} else if strings.EqualFold(allowedType, "ssh") {
				helpText.WriteString("\nTo sign with SSH:\n")
				helpText.WriteString("1. Configure Git to use your SSH key: `git config --global commit.gpgsign true`\n")
				helpText.WriteString("2. Set signing format: `git config --global gpg.format ssh`\n")
				helpText.WriteString("3. Specify your key: `git config --global user.signingkey ~/.ssh/id_ed25519.pub`\n")
			}
		}
	} else {
		helpText.WriteString("Signature validation failed. Please check your signing configuration.\n")
	}

	return helpText.String()
}

// containsError checks if an error list contains an error with the given code.
func containsError(errors []appErrors.ValidationError, code appErrors.ValidationErrorCode) bool {
	for _, err := range errors {
		if err.Code == string(code) {
			return true
		}
	}

	return false
}
