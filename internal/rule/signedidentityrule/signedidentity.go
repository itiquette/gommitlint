// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/model"
)

var MinimumRSABits uint16 = 2048
var MinimumECBits uint16 = 256

const (
	SSH = "SSH"
	GPG = "GPG"
)

// SignedIdentity validates that a commit is properly signed with either GPG or SSH.
// This rule helps ensure that code changes are securely authenticated and attributable
// to a verified identity, which is crucial for maintaining supply chain security and
// establishing an audit trail of code changes.
//
// The rule checks:
// - Whether the commit has a cryptographic signature (GPG or SSH)
// - If the signature can be verified against trusted keys
// - That the key used for signing meets minimum security requirements
//
// Example usage:
//
//	// Get the signature from a commit
//	signature := commit.PGPSignature
//
//	// Verify against trusted keys stored in a specific directory
//	rule := VerifyCommitSignature(commit, signature, "/path/to/trusted/keys")
//
//	if len(rule.Errors()) > 0 {
//	    // Handle validation failure
//	    fmt.Println(rule.Help())
//	} else {
//	    // Display who signed the commit
//	    fmt.Printf("Commit verified: %s\n", rule.Result())
//	}
type SignedIdentity struct {
	errors        []*model.ValidationError
	Identity      string // Email or name of the signer
	SignatureType string // "GPG" or "SSH"
	KeyDir        string // Directory used for key verification
}

// Name returns the rule identifier.
func (s SignedIdentity) Name() string {
	return "SignedIdentity"
}

// Result returns a concise string representation of the rule's status.
func (s SignedIdentity) Result() string {
	if len(s.errors) > 0 {
		return "Invalid signature"
	}

	return "Valid signature"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (s SignedIdentity) VerboseResult() string {
	if len(s.errors) > 0 {
		switch s.errors[0].Code {
		case "commit_nil":
			return "Cannot verify signature: commit object is nil"
		case "no_key_dir":
			return "Cannot verify signature: no trusted key directory provided"
		case "invalid_key_dir":
			var errorMsg string

			for k, v := range s.errors[0].Context {
				if k == "error" {
					errorMsg = v

					break
				}
			}

			return "Cannot verify signature: invalid key directory - " + errorMsg
		case "no_signature":
			return "No cryptographic signature found on commit"
		case "invalid_signature_format":
			var errorMsg string

			for k, v := range s.errors[0].Context {
				if k == "error" {
					errorMsg = v

					break
				}
			}

			return "Invalid " + s.SignatureType + " signature format: " + errorMsg
		case "key_not_trusted":
			return "Signature verified but the key is not in the trusted keys directory"
		case "weak_key":
			var bits, required string

			for k, v := range s.errors[0].Context {
				if k == "key_bits" {
					bits = v
				} else if k == "required_bits" {
					required = v
				}
			}

			return "Weak " + s.SignatureType + " key detected: " + bits + " bits (minimum required: " + required + " bits)"
		case "verification_failed":
			var errorMsg string

			for k, v := range s.errors[0].Context {
				if k == "error" {
					errorMsg = v

					break
				}
			}

			return "Signature verification failed: " + errorMsg
		case "unknown_signature_type":
			return "Unknown signature type, cannot verify identity"
		default:
			return s.errors[0].Error()
		}
	}

	return fmt.Sprintf("Valid %s signature from %q", s.SignatureType, s.Identity)
}

// addError adds a structured validation error.
func (s *SignedIdentity) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("SignedIdentity", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	s.errors = append(s.errors, err)
}

// Errors returns validation errors.
func (s SignedIdentity) Errors() []*model.ValidationError {
	return s.errors
}

// Help returns a description of how to fix the rule violation.
func (s SignedIdentity) Help() string {
	if len(s.errors) == 0 {
		return "No errors to fix"
	}

	// First check for specific error codes
	if len(s.errors) > 0 {
		switch s.errors[0].Code {
		case "commit_nil":
			return "A valid commit object is required for signature verification"
		case "no_key_dir":
			return "Please provide a valid directory containing trusted public keys for verification"
		case "invalid_key_dir":
			return "The specified key directory is invalid or inaccessible. Please provide a valid path to a directory containing trusted public keys"
		case "no_signature":
			return "This commit is not signed. Please configure Git to sign your commits with either GPG or SSH"
		case "invalid_signature_format":
			return "The signature format is invalid or corrupted. Please ensure you're using a properly configured signing key"
		case "key_not_trusted":
			return "The signature was created with a key that is not in the trusted keys directory. Add the public key to your trusted keys directory"
		case "weak_key":
			keyType := s.errors[0].Context["key_type"]
			bits := s.errors[0].Context["key_bits"]
			required := s.errors[0].Context["required_bits"]

			return fmt.Sprintf("The %s key used for signing (%s bits) does not meet the minimum strength requirement of %s bits. Please generate a stronger key",
				keyType, bits, required)
		case "verification_failed":
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

// VerifySignatureIdentity checks if a commit is signed with a trusted key.
// It automatically detects whether the signature is GPG or SSH based on its format
// and validates it against trusted public keys stored in the specified directory.
//
// The function performs several security checks:
//   - Validates that the signature corresponds to the commit content
//   - Verifies the signature against trusted keys in keyDir
//   - Checks that the signing key meets minimum strength requirements
//     (RSA: 2048 bits, EC: 256 bits by default)
func VerifySignatureIdentity(commit *object.Commit, signature string, keyDir string) *SignedIdentity {
	rule := &SignedIdentity{
		KeyDir: keyDir,
	}

	if commit == nil {
		rule.addError(
			"commit_nil",
			"commit cannot be nil",
			map[string]string{},
		)

		return rule
	}

	if keyDir == "" {
		rule.addError(
			"no_key_dir",
			"no key directory provided",
			map[string]string{},
		)

		return rule
	}

	// Sanitize keyDir to prevent path traversal
	sanitizedKeyDir, err := sanitizePath(keyDir)
	if err != nil {
		rule.addError(
			"invalid_key_dir",
			fmt.Sprintf("invalid key directory: %s", err),
			map[string]string{
				"key_dir": keyDir,
				"error":   err.Error(),
			},
		)

		return rule
	}

	if signature == "" {
		rule.addError(
			"no_signature",
			"no signature provided",
			map[string]string{},
		)

		return rule
	}

	// Get commit data
	commitBytes, err := getCommitBytes(commit)
	if err != nil {
		rule.addError(
			"commit_data_error",
			fmt.Sprintf("failed to prepare commit data: %s", err),
			map[string]string{
				"error": err.Error(),
			},
		)

		return rule
	}

	// Auto-detect signature type
	sigType := detectSignatureType(signature)
	rule.SignatureType = sigType

	// Helper function to handle verification errors
	handleVerificationError := func(err error, sigType string) bool {
		if err == nil {
			return false
		}

		// Determine error type and add appropriate validation error
		if strings.Contains(err.Error(), "not verified with any trusted key") {
			rule.addError(
				"key_not_trusted",
				err.Error(),
				map[string]string{
					"signature_type": sigType,
					"error":          err.Error(),
				},
			)
		} else if strings.Contains(err.Error(), "key strength") {
			// Extract key bits from error message
			var bits, required string

			if parts := strings.Split(err.Error(), "bits (required: "); len(parts) > 1 {
				keyBitsParts := strings.Split(parts[0], "key strength: ")
				if len(keyBitsParts) > 1 {
					bits = strings.TrimSpace(keyBitsParts[1])
				}

				required = strings.TrimSuffix(parts[1], " bits)")
			}

			rule.addError(
				"weak_key",
				err.Error(),
				map[string]string{
					"signature_type": sigType,
					"key_type":       sigType,
					"key_bits":       bits,
					"required_bits":  required,
				},
			)
		} else {
			rule.addError(
				"verification_failed",
				err.Error(),
				map[string]string{
					"signature_type": sigType,
					"error":          err.Error(),
				},
			)
		}

		return true
	}

	// Verify based on signature type
	switch sigType {
	case GPG:
		identity, err := verifyGPGSignature(commitBytes, signature, sanitizedKeyDir)
		if handleVerificationError(err, GPG) {
			return rule
		}

		rule.Identity = identity

	case SSH:
		// Parse SSH signature from string
		format, blob, err := parseSSHSignature(signature)
		if err != nil {
			rule.addError(
				"invalid_signature_format",
				fmt.Sprintf("invalid SSH signature format: %s", err),
				map[string]string{
					"signature_type": SSH,
					"error":          err.Error(),
				},
			)

			return rule
		}

		identity, err := verifySSHSignature(commitBytes, format, blob, sanitizedKeyDir)
		if handleVerificationError(err, SSH) {
			return rule
		}

		rule.Identity = identity

	default:
		rule.addError(
			"unknown_signature_type",
			"unknown signature type",
			map[string]string{
				"signature": signature,
			},
		)
	}

	return rule
}

// detectSignatureType determines whether a signature is GPG or SSH based on its format.
func detectSignatureType(signature string) string {
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

// The following functions are assumed to be defined elsewhere in the package:
// sanitizePath(path string) (string, error)
// getCommitBytes(commit *object.Commit) ([]byte, error)
// parseSSHSignature(signature string) (string, []byte, error)
// verifyGPGSignature(commitData []byte, signature string, keyDir string) (string, error)
// verifySSHSignature(commit []byte, format string, blob []byte, keyDir string) (string, error)
