// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SignatureRule validates that commits have cryptographic signatures.
// Uses the crypto domain for signature handling.
type SignatureRule struct {
	name             string
	requireSignature bool
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

// NewSignatureRule creates a new rule for validating commit signatures.
func NewSignatureRule(options ...SignatureOption) SignatureRule {
	// Create a rule with default values
	rule := SignatureRule{
		name:             "Signature",
		requireSignature: true, // Default to requiring signatures
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate checks for the presence and format of cryptographic signatures.
func (r SignatureRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Simple validation based on rule settings
	// Configuration is now handled by RuleRegistry, not here
	// If signatures are not required, skip validation
	if !r.requireSignature {
		return nil
	}

	// If signature is missing and required, return error
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

	return nil
}

// WithContext returns a new rule with configuration from the provided context.
// This implements domain.ConfigurableRule interface.
func (r SignatureRule) WithContext(ctx context.Context) domain.Rule {
	// Create a copy of the rule
	result := r

	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return result
	}

	// Apply configuration settings related to this rule
	requireSignature := cfg.GetBool("signing.require_signature")

	// Only override if explicitly set to false (keep default otherwise)
	if !requireSignature {
		result.requireSignature = false
	}

	return result
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return r.name
}

// RequireSignature returns whether signatures are required by this rule.
func (r SignatureRule) RequireSignature() bool {
	return r.requireSignature
}
