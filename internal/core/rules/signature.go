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
	"github.com/itiquette/gommitlint/internal/domain/crypto"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SignatureRule validates that commits have cryptographic signatures.
// Uses the crypto domain for signature handling.
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

// Validate checks for the presence and format of cryptographic signatures.
func (r SignatureRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating signature using context configuration", "rule", r.Name(), "commit_hash", commit.Hash)

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Check if signature is required
	if !rule.requireSignature {
		return nil
	}

	// Create a signature object from the commit
	signature := crypto.NewSignature(commit.Signature)

	// Check if signature exists
	if signature.IsEmpty() {
		return []appErrors.ValidationError{
			appErrors.NewSignatureError(
				appErrors.ErrMissingSignature,
				"Signature",
				"Commit must be cryptographically signed",
				"Sign your commits using 'git commit -S' for GPG or 'git commit --signoff' for DCO",
			),
		}
	}

	// Get signature type as string for comparison
	sigType := string(signature.Type())
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

// withContextConfig creates a new rule with configuration from context.
func (r SignatureRule) withContextConfig(ctx context.Context) SignatureRule {
	// Create a copy of the rule
	result := r

	// In a real implementation, this would read config from context
	// Just keeping defaults from constructor for now

	// Log configuration at debug level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Signature rule configuration (final)",
		"require_gpg", result.requireSignature,
		"allowed_sig_types", result.allowedSigTypes)

	return result
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return r.name
}
