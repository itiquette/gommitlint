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
	requireSignature bool
	verifyFormat     bool
	allowedSigners   []string
}

// NewSignatureRule creates a new rule for validating commit signatures from config.
func NewSignatureRule(cfg config.Config) SignatureRule {
	return SignatureRule{
		requireSignature: cfg.Signature.Required,
		verifyFormat:     cfg.Signature.VerifyFormat,
		allowedSigners:   cfg.Signature.AllowedSigners,
	}
}

// Validate checks if a commit has the required cryptographic signature.
func (r SignatureRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Skip if signature not required
	if !r.requireSignature {
		return nil
	}

	// Functional composition of validations
	var errors []domain.ValidationError
	errors = append(errors, r.validatePresence(commit)...)

	// Only validate format and signers if signature exists
	if commit.Signature != "" {
		errors = append(errors, r.validateFormat(commit)...)
		errors = append(errors, r.validateSigners(commit)...)
	}

	return errors
}

// Name returns the rule name.
func (r SignatureRule) Name() string {
	return "Signature"
}

// validatePresence checks if signature is present when required.
func (r SignatureRule) validatePresence(commit domain.Commit) []domain.ValidationError {
	signature := strings.TrimSpace(commit.Signature)
	if signature == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingSignature, "Missing cryptographic signature").
				WithContextMap(map[string]string{
					"actual":   "no signature",
					"expected": "signed commit",
				}).
				WithHelp("Sign your commits using 'git commit -S' for GPG or configure SSH signing"),
		}
	}

	return nil
}

// validateFormat validates signature format using domain.Signature.
func (r SignatureRule) validateFormat(commit domain.Commit) []domain.ValidationError {
	if !r.verifyFormat {
		return nil
	}

	signature := domain.NewSignature(commit.Signature)
	if signature.IsEmpty() {
		return nil
	}

	// Check if signature type is recognized
	if signature.Type() == domain.SignatureTypeUnknown {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrUnknownSigFormat, "Unknown signature format").
				WithContextMap(map[string]string{
					"actual":   "unknown format",
					"expected": "GPG or SSH signature",
				}).
				WithHelp("Ensure your signing key is properly configured"),
		}
	}

	// Validate signature completeness based on type
	if signature.Type() == domain.SignatureTypeGPG {
		if !isCompleteGPGSignature(signature.Data()) {
			return []domain.ValidationError{
				domain.New(r.Name(), domain.ErrIncompleteGPGSig, "Incomplete GPG signature").
					WithContextMap(map[string]string{
						"actual":   "incomplete GPG signature",
						"expected": "complete GPG signature",
					}).
					WithHelp("Ensure GPG signature includes complete BEGIN/END markers"),
			}
		}
	} else if signature.Type() == domain.SignatureTypeSSH {
		if !isCompleteSSHSignature(signature.Data()) {
			return []domain.ValidationError{
				domain.New(r.Name(), domain.ErrIncompleteSSHSig, "Incomplete SSH signature").
					WithContextMap(map[string]string{
						"actual":   "incomplete SSH signature",
						"expected": "complete SSH signature",
					}).
					WithHelp("Ensure SSH signature includes complete BEGIN/END markers"),
			}
		}
	}
	// SignatureTypeUnknown is already handled above

	return nil
}

// validateSigners validates that the signer is in the allowed list.
func (r SignatureRule) validateSigners(commit domain.Commit) []domain.ValidationError {
	if len(r.allowedSigners) == 0 {
		return nil
	}

	// Use author email as signer identity
	signer := commit.AuthorEmail
	if signer == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrVerificationFailed, "Cannot determine signer identity").
				WithContextMap(map[string]string{
					"actual":   "no author email",
					"expected": "valid author email",
				}).
				WithHelp("Ensure commit has valid author information"),
		}
	}

	// Check if signer is in allowed list
	for _, allowed := range r.allowedSigners {
		if signer == allowed {
			return nil
		}
	}

	return []domain.ValidationError{
		domain.New(r.Name(), domain.ErrKeyNotTrusted, "Signer not in allowed signers list").
			WithContextMap(map[string]string{
				"actual":   signer,
				"expected": strings.Join(r.allowedSigners, ", "),
			}).
			WithHelp("Contact your repository administrator to add your signing key"),
	}
}

// isCompleteGPGSignature checks if a GPG signature has the required components.
func isCompleteGPGSignature(signature string) bool {
	return strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") &&
		strings.Contains(signature, "-----END PGP SIGNATURE-----") &&
		len(strings.TrimSpace(signature)) > 100 // Basic sanity check for content
}

// isCompleteSSHSignature checks if an SSH signature has the required components.
func isCompleteSSHSignature(signature string) bool {
	return strings.Contains(signature, "-----BEGIN SSH SIGNATURE-----") &&
		strings.Contains(signature, "-----END SSH SIGNATURE-----") &&
		len(strings.TrimSpace(signature)) > 50 // Basic sanity check for content
}
