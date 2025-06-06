// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SignatureRule validates that commits have cryptographic signatures when required.
type SignatureRule struct {
	requireSignature  bool
	requireSigningKey bool
}

// NewSignatureRule creates a new rule for validating commit signatures from config.
func NewSignatureRule(cfg config.Config) SignatureRule {
	return SignatureRule{
		requireSignature:  cfg.Signing.RequireSignature,
		requireSigningKey: cfg.Signing.RequireVerification,
	}
}

// Validate checks if a commit has the required cryptographic signature.
func (r SignatureRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Check if signatures are required
	if r.requireSignature {
		// Check for any signature (GPG or SSH format)
		signature := strings.TrimSpace(commit.Signature)
		if signature == "" {
			return []domain.ValidationError{
				domain.New(r.Name(), domain.ErrMissingSignature, "Commit must be cryptographically signed").
					WithHelp("Sign your commits using 'git commit -S' for GPG or 'git commit --signoff' for DCO"),
			}
		}
	}

	return nil
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return "Signature"
}
