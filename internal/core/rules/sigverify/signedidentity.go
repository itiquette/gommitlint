// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package sigverify

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// Security requirement constants.
var (
	MinimumRSABits uint16 = 2048
	MinimumECBits  uint16 = 256
)

// Signature types.
const (
	SSH = "SSH"
	GPG = "GPG"
)

// VerificationResult contains the outcome of signature verification.
type VerificationResult struct {
	Identity      string
	SignatureType string
	Errors        []appErrors.ValidationError
}

// NewVerificationResult creates a new empty verification result.
func NewVerificationResult() VerificationResult {
	return VerificationResult{
		Errors: make([]appErrors.ValidationError, 0),
	}
}

// DetectSignatureType determines whether a signature is GPG or SSH based on its format.
func DetectSignatureType(signature string) string {
	// Check for SSH signature format (format:blob)
	if strings.Contains(signature, ":") && strings.HasPrefix(signature, "ssh-") {
		return SSH
	}

	// Check for GPG signature format (PGP block)
	if strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") {
		return GPG
	}

	// Check for other common SSH format prefixes
	sshPrefixes := []string{"ecdsa-", "sk-ssh-", "ssh-ed25519"}
	for _, prefix := range sshPrefixes {
		if strings.HasPrefix(signature, prefix) {
			return SSH
		}
	}

	// Default to GPG for other formats
	return GPG
}

// VerifySignature verifies a commit signature and returns the verification result.
func VerifySignature(commit *object.Commit, keyDir string) VerificationResult {
	result := NewVerificationResult()

	// Validate inputs
	if commit == nil {
		result.Errors = append(result.Errors, createError(
			appErrors.ErrCommitNil,
			"commit cannot be nil",
			nil,
		))

		return result
	}

	if keyDir == "" {
		result.Errors = append(result.Errors, createError(
			appErrors.ErrNoKeyDir,
			"no key directory provided",
			nil,
		))

		return result
	}

	// Sanitize keyDir to prevent path traversal
	sanitizedKeyDir, err := SanitizePath(keyDir)
	if err != nil {
		result.Errors = append(result.Errors, createError(
			appErrors.ErrInvalidKeyDir,
			fmt.Sprintf("invalid key directory: %s", err),
			map[string]string{
				"key_dir": keyDir,
				"error":   err.Error(),
			},
		))

		return result
	}

	// Get the signature from the commit
	signature := commit.PGPSignature
	if signature == "" {
		result.Errors = append(result.Errors, createError(
			appErrors.ErrMissingSignature,
			"no signature provided",
			nil,
		))

		return result
	}

	// Get commit data
	commitBytes, err := getCommitBytes(commit)
	if err != nil {
		result.Errors = append(result.Errors, createError(
			appErrors.ErrInvalidCommit,
			fmt.Sprintf("failed to prepare commit data: %s", err),
			map[string]string{
				"error": err.Error(),
			},
		))

		return result
	}

	// Detect signature type
	sigType := DetectSignatureType(signature)
	result.SignatureType = sigType

	// Verify based on signature type
	switch sigType {
	case GPG:
		identity, err := VerifyGPGSignature(commitBytes, signature, sanitizedKeyDir)
		if err != nil {
			result.Errors = append(result.Errors, handleVerificationError(err, GPG))

			return result
		}

		result.Identity = identity

	case SSH:
		// Parse SSH signature from string
		format, blob, err := ParseSSHSignature(signature)
		if err != nil {
			result.Errors = append(result.Errors, createError(
				appErrors.ErrInvalidSignatureFormat,
				fmt.Sprintf("invalid SSH signature format: %s", err),
				map[string]string{
					"signature_type": SSH,
					"error":          err.Error(),
				},
			))

			return result
		}

		identity, err := VerifySSHSignature(commitBytes, format, blob, sanitizedKeyDir)
		if err != nil {
			result.Errors = append(result.Errors, handleVerificationError(err, SSH))

			return result
		}

		result.Identity = identity

	default:
		result.Errors = append(result.Errors, createError(
			appErrors.ErrUnknownSigFormat,
			"unknown signature type",
			map[string]string{
				"signature": signature,
			},
		))
	}

	return result
}

// handleVerificationError creates the appropriate validation error based on the error type.
func handleVerificationError(err error, sigType string) appErrors.ValidationError {
	if err == nil {
		// This shouldn't happen, but if it does, return a generic error
		return createError(
			appErrors.ErrUnknown,
			"unknown verification error",
			nil,
		)
	}

	errMsg := err.Error()

	// Determine error type and create appropriate validation error
	switch {
	case strings.Contains(errMsg, "not verified with any trusted key"):
		return createError(
			appErrors.ErrKeyNotTrusted,
			errMsg,
			map[string]string{
				"signature_type": sigType,
				"error":          errMsg,
			},
		)

	case strings.Contains(errMsg, "key strength"):
		// Extract key bits from error message
		var bits, required string

		if parts := strings.Split(errMsg, "bits (required: "); len(parts) > 1 {
			keyBitsParts := strings.Split(parts[0], "key strength: ")
			if len(keyBitsParts) > 1 {
				bits = strings.TrimSpace(keyBitsParts[1])
			}

			required = strings.TrimSuffix(parts[1], " bits)")
		}

		return createError(
			appErrors.ErrWeakKey,
			errMsg,
			map[string]string{
				"signature_type": sigType,
				"key_type":       sigType,
				"key_bits":       bits,
				"required_bits":  required,
			},
		)

	default:
		return createError(
			appErrors.ErrVerificationFailed,
			errMsg,
			map[string]string{
				"signature_type": sigType,
				"error":          errMsg,
			},
		)
	}
}

// createError creates a validation error.
func createError(code appErrors.ValidationErrorCode, message string, context map[string]string) appErrors.ValidationError {
	ruleName := "SignedIdentity"

	// Create a basic error
	err := appErrors.NewValidationError(code, ruleName, message, message)

	// Add context if provided
	for key, value := range context {
		err = err.WithContext(key, value)
	}

	return err
}
