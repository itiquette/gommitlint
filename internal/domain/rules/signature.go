// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
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
				domain.New(r.Name(), domain.ErrMissingSignature, "Missing cryptographic signature").
					WithContextMap(map[string]string{
						"current_state":    "No signature present",
						"supported_types":  "GPG, SSH",
						"signing_commands": "git commit -S (GPG) or configure SSH signing",
					}).
					WithHelp("Sign your commits using 'git commit -S' for GPG or configure SSH signing"),
			}
		}

		// Check signature format and provide context if verbose
		if r.requireSigningKey && signature != "" {
			signatureType := detectSignatureType(signature)
			if signatureType == "unknown" {
				return []domain.ValidationError{
					domain.New(r.Name(), domain.ErrInvalidSignatureFormat, "Invalid signature format").
						WithContextMap(map[string]string{
							"current_signature": fmt.Sprintf("%.50s...", signature),
							"supported_formats": "GPG: -----BEGIN PGP SIGNATURE-----, SSH: -----BEGIN SSH SIGNATURE-----",
						}).
						WithHelp("Ensure your signing key is properly configured"),
				}
			}
		}
	}

	return nil
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return "Signature"
}

// detectSignatureType detects the type of signature (GPG, SSH, or unknown).
func detectSignatureType(signature string) string {
	if strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") {
		return "GPG"
	}

	if strings.Contains(signature, "-----BEGIN SSH SIGNATURE-----") {
		return "SSH"
	}

	return "unknown"
}
