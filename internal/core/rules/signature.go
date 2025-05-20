// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SignatureRule validates that commits have cryptographic signatures.
type SignatureRule struct {
	name               string
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
		name:             "Signature",
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
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating signature using context configuration", "rule", r.Name(), "commit_hash", commit.Hash)

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Check if signature is required
	if !rule.requireSignature {
		return nil
	}

	// Check if signature exists
	if commit.Signature == "" {
		return []appErrors.ValidationError{
			appErrors.NewSignatureError(
				appErrors.ErrMissingSignature,
				"Signature",
				"Commit must be cryptographically signed",
				"Sign your commits using 'git commit -S' for GPG or 'git commit --signoff' for DCO",
			),
		}
	}

	// Determine signature type
	sigType := determineSignatureType(commit.Signature)
	r.foundSignatureType = sigType

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
			return []appErrors.ValidationError{
				appErrors.NewSignatureError(
					appErrors.ErrDisallowedSigType,
					"Signature",
					fmt.Sprintf("Signature type '%s' is not allowed; allowed types: %s", sigType, strings.Join(rule.allowedSigTypes, ", ")),
					"Use one of the allowed signature types: "+strings.Join(rule.allowedSigTypes, ", "),
				).WithContextMap(map[string]string{
					"found_type":    sigType,
					"allowed_types": strings.Join(rule.allowedSigTypes, ", "),
				}),
			}
		}
	}

	return nil
}

// Note: Rule constructor options take precedence over context configuration.
func (r SignatureRule) withContextConfig(ctx context.Context) SignatureRule {
	// Create a copy of the rule
	result := r

	// We're skipping config from context because our adapter doesn't have security settings
	// Just keeping defaults from constructor

	// Log configuration at debug level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Signature rule configuration (final)",
		"require_signature", result.requireSignature,
		"allowed_sig_types", result.allowedSigTypes)

	return result
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return r.name
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
