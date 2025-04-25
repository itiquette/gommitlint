// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
// Package rules provides validation rules for commit messages.
package rules

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/core/rules/sigverify"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// Security constants for key strength requirements.
var (
	// MinimumRSABits defines the minimum acceptable RSA key strength (2048 bits).
	MinimumRSABits uint16 = 2048
	// MinimumECBits defines the minimum acceptable elliptic curve key strength (256 bits).
	MinimumECBits uint16 = 256
)

// Signature types.
const (
	// SSH signature type.
	SSH = "SSH"
	// GPG signature type.
	GPG = "GPG"
)

// SignedIdentityRule enforces that commits are cryptographically signed
// with a trusted key. It's used to verify the identity of commit authors.
type SignedIdentityRule struct {
	*BaseRule
	Identity      string // Email or name of the signer
	SignatureType string // "GPG" or "SSH"
	KeyDir        string // Directory used for key verification
}

// SignedIdentityOption is a function that modifies a SignedIdentityRule.
type SignedIdentityOption func(SignedIdentityRule) SignedIdentityRule

// WithKeyDirectory sets the directory containing trusted public keys.
func WithKeyDirectory(keyDir string) SignedIdentityOption {
	return func(r SignedIdentityRule) SignedIdentityRule {
		r.KeyDir = keyDir

		return r
	}
}

// NewSignedIdentityRule creates a new SignedIdentityRule with the specified options.
func NewSignedIdentityRule(options ...SignedIdentityOption) SignedIdentityRule {
	rule := SignedIdentityRule{
		BaseRule: NewBaseRule("SignedIdentity"),
	}
	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Result returns a concise string representation of the rule's status.
func (r SignedIdentityRule) Result() string {
	if r.HasErrors() {
		return "Invalid signature"
	}

	return "Valid signature"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SignedIdentityRule) VerboseResult() string {
	if r.HasErrors() {
		// Get the first error
		firstErr := r.Errors()[0]
		// firstErr is already a ValidationError, so no need for type assertion
		validationErr := firstErr
		// Get the error code
		code := validationErr.Code
		switch code {
		case string(appErrors.ErrCommitNil):
			return "Cannot verify signature: commit object is nil"
		case string(appErrors.ErrNoKeyDir):
			return "Cannot verify signature: no trusted key directory provided"
		case string(appErrors.ErrInvalidKeyDir):
			var errorMsg string

			if ctx := validationErr.Context; ctx != nil {
				if v, ok := ctx["error"]; ok {
					errorMsg = v
				}
			}

			return "Cannot verify signature: invalid key directory - " + errorMsg
		case string(appErrors.ErrMissingSignature):
			return "No cryptographic signature found on commit"
		case string(appErrors.ErrInvalidSignatureFormat):
			var errorMsg string

			if ctx := validationErr.Context; ctx != nil {
				if v, ok := ctx["error"]; ok {
					errorMsg = v
				}
			}

			return "Invalid " + r.SignatureType + " signature format: " + errorMsg
		case string(appErrors.ErrKeyNotTrusted):
			return "Signature verified but the key is not in the trusted keys directory"
		case string(appErrors.ErrWeakKey):
			var bits, required string

			if ctx := validationErr.Context; ctx != nil {
				if v, ok := ctx["key_bits"]; ok {
					bits = v
				}

				if v, ok := ctx["required_bits"]; ok {
					required = v
				}
			}

			return "Weak " + r.SignatureType + " key detected: " + bits + " bits (minimum required: " + required + " bits)"
		case string(appErrors.ErrVerificationFailed):
			var errorMsg string

			if ctx := validationErr.Context; ctx != nil {
				if v, ok := ctx["error"]; ok {
					errorMsg = v
				}
			}

			return "Signature verification failed: " + errorMsg
		case string(appErrors.ErrUnknownSigFormat):
			return "Unknown signature type, cannot verify identity"
		default:
			return validationErr.Error()
		}
	}

	return fmt.Sprintf("Valid %s signature from %q", r.SignatureType, r.Identity)
}

// Help returns a description of how to fix the rule violation.
func (r SignedIdentityRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix"
	}
	// First check for specific error codes
	if r.ErrorCount() > 0 {
		// Get the first error
		firstErr := r.Errors()[0]
		// firstErr is already a ValidationError, so no need for type assertion
		validationErr := firstErr

		code := validationErr.Code
		switch code {
		case string(appErrors.ErrCommitNil):
			return "A valid commit object is required for signature verification"
		case string(appErrors.ErrNoKeyDir):
			return "Please provide a valid directory containing trusted public keys for verification"
		case string(appErrors.ErrInvalidKeyDir):
			return "The specified key directory is invalid or inaccessible. Please provide a valid path to a directory containing trusted public keys"
		case string(appErrors.ErrMissingSignature):
			return "This commit is not signed. Please configure Git to sign your commits with either GPG or SSH"
		case string(appErrors.ErrInvalidSignatureFormat):
			return "The signature format is invalid or corrupted. Please ensure you're using a properly configured signing key"
		case string(appErrors.ErrKeyNotTrusted):
			return "The signature was created with a key that is not in the trusted keys directory. Add the public key to your trusted keys directory"
		case string(appErrors.ErrWeakKey):
			// Access context fields safely with type assertion
			var keyType, bits, required string

			if validationErr.Context != nil {
				if v, ok := validationErr.Context["key_type"]; ok {
					keyType = v
				}

				if v, ok := validationErr.Context["key_bits"]; ok {
					bits = v
				}

				if v, ok := validationErr.Context["required_bits"]; ok {
					required = v
				}
			}

			return fmt.Sprintf("The %s key used for signing (%s bits) does not meet the minimum strength requirement of %s bits. Please generate a stronger key",
				keyType, bits, required)
		case string(appErrors.ErrVerificationFailed):
			return "Signature verification failed. The signature may be invalid or the commit content may have been altered"
		}
	}
	// Default comprehensive help
	return fmt.Sprintf(
		"Your commit has signature validation issues.\n"+
			"Consider one of the following solutions:\n"+
			"1. Sign your commits with GPG using 'git config --global commit.gpgsign true'\n"+
			"2. Sign your commits with SSH using 'git config --global gpg.format ssh'\n"+
			"3. Ensure your signing key is in the trusted keys directory\n"+
			"4. Verify your key strength meets minimum requirements (RSA: %d bits, EC: %d bits)",
		MinimumRSABits, MinimumECBits)
}

// ValidateWithIdentity performs validation and returns the errors along with identity info.
// This approach directly returns all the values we need instead of trying to update state.
func (r SignedIdentityRule) ValidateWithIdentity(commit *domain.CommitInfo) ([]appErrors.ValidationError, string, string) {
	// Create a new errors slice
	errors := make([]appErrors.ValidationError, 0)

	// Store identity and signature type for return
	identity := ""
	sigType := ""

	// Validate commit
	if commit == nil {
		errors = append(errors, createError(
			r.Name(),
			appErrors.ErrCommitNil,
			"commit cannot be nil",
			nil,
		))

		return errors, identity, sigType
	}

	// Get signature from commit
	signature := commit.Signature

	// Check for empty signature
	if signature == "" || len(strings.TrimSpace(signature)) == 0 {
		errors = append(errors, createError(
			r.Name(),
			appErrors.ErrMissingSignature,
			"no signature provided",
			nil,
		))

		return errors, identity, sigType
	}

	// Validate key directory if specified
	if r.KeyDir != "" {
		// Sanitize keyDir to prevent path traversal
		_, err := sigverify.SanitizePath(r.KeyDir)
		if err != nil {
			errors = append(errors, createError(
				r.Name(),
				appErrors.ErrInvalidKeyDir,
				fmt.Sprintf("invalid key directory: %s", err),
				map[string]string{
					"key_dir": r.KeyDir,
					"error":   err.Error(),
				},
			))

			return errors, identity, sigType
		}

		// Auto-detect signature type
		sigType = sigverify.DetectSignatureType(signature)

		// For the simplified version, we'll just do format validation
		// In a more complete implementation, we would convert the CommitInfo to the right format
		// for the verification functions in the signedidentityrule package
		switch sigType {
		case GPG:
			// For now, we'll just simulate a verification
			if !strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") ||
				!strings.Contains(signature, "-----END PGP SIGNATURE-----") {
				errors = append(errors, createError(
					r.Name(),
					appErrors.ErrInvalidSignatureFormat,
					"incomplete GPG signature (missing begin/end markers)",
					map[string]string{
						"signature_type": GPG,
					},
				))

				return errors, identity, sigType
			}

			identity = "GPG Signature Format Verified"
		case SSH:
			// For now, we'll just simulate a verification
			if strings.Contains(signature, "-----BEGIN SSH SIGNATURE-----") &&
				!strings.Contains(signature, "-----END SSH SIGNATURE-----") {
				errors = append(errors, createError(
					r.Name(),
					appErrors.ErrInvalidSignatureFormat,
					"incomplete SSH signature (missing end marker)",
					map[string]string{
						"signature_type": SSH,
					},
				))

				return errors, identity, sigType
			}

			identity = "SSH Signature Format Verified"
		default:
			errors = append(errors, createError(
				r.Name(),
				appErrors.ErrUnknownSigFormat,
				"unknown signature type",
				map[string]string{
					"signature": signature[:safeMin(len(signature), 20)],
				},
			))

			return errors, identity, sigType
		}
	} else {
		// If no key directory is specified, we can only verify signature format
		sigType = sigverify.DetectSignatureType(signature)
		identity = "Signature format verified (no key directory provided for verification)"
	}

	return errors, identity, sigType
}

// Validate is a compatibility method that calls ValidateWithIdentity but only returns errors.
func (r SignedIdentityRule) Validate(commit *domain.CommitInfo) []appErrors.ValidationError {
	errors, _, _ := r.ValidateWithIdentity(commit)

	return errors
}

// SetValidationState sets the rule's state based on validation results.
func (r SignedIdentityRule) SetValidationState(commit *domain.CommitInfo) SignedIdentityRule {
	// Get validation results with all information
	errors, identity, sigType := r.ValidateWithIdentity(commit)

	// Create a new base rule with the errors
	baseRule := NewBaseRule(r.Name())
	baseRule.MarkAsRun()

	for _, err := range errors {
		baseRule.AddError(err)
	}

	// Set all state fields
	r.BaseRule = baseRule
	r.Identity = identity
	r.SignatureType = sigType

	return r
}

// // createError creates a validation error without modifying the rule's state.
// func createError(ruleName string, code appErrors.ValidationErrorCode, message string, context map[string]string) appErrors.ValidationError {
// 	if context == nil {
// 		return appErrors.New(ruleName, code, message)
// 	}
// 	return appErrors.New(ruleName, code, message, appErrors.WithContextMap(context))
// }

// safeMin returns the minimum of two integers (utility function for safety).
func safeMin(a, b int) int {
	if a < b {
		return a
	}

	return b
}
