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
	errors          []appErrors.ValidationError
	name            string
	hasRun          bool
	Identity        string              // Email or name of the signer
	SignatureType   string              // "GPG" or "SSH"
	KeyDir          string              // Directory used for key verification
	allowedTypes    []string            // Allowed signature types (defaults to [GPG, SSH])
	allowedKeyTypes map[string][]string // Allowed key types for each signature type
}

// Name returns the rule identifier.
func (s SignedIdentity) Name() string {
	return "SignedIdentity"
}

// HasErrors returns true if the rule has found any errors.
func (s SignedIdentity) HasErrors() bool {
	return len(s.errors) > 0
}

// Errors returns validation errors.
func (s SignedIdentity) Errors() []appErrors.ValidationError {
	return s.errors
}

// Result returns a concise string representation of the rule's status.
func (s SignedIdentity) Result() string {
	if !s.hasRun {
		return "Rule has not been run"
	}

	if s.HasErrors() {
		return "Invalid signature"
	}

	if s.SignatureType != "" && s.Identity != "" {
		return fmt.Sprintf("Valid %s signature from %q", s.SignatureType, s.Identity)
	}

	return "Valid signature"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (s SignedIdentity) VerboseResult() string {
	if !s.hasRun {
		return "SignedIdentity: Rule has not been run"
	}

	if s.HasErrors() {
		errors := s.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		switch errors[0].Code {
		case string(appErrors.ErrCommitNil):
			return "Cannot verify signature: commit object is nil"
		case string(appErrors.ErrNoKeyDir):
			return "Cannot verify signature: no trusted key directory provided"
		case string(appErrors.ErrInvalidKeyDir):
			var errorMsg string

			for k, v := range errors[0].Context {
				if k == "error" {
					errorMsg = v

					break
				}
			}

			return "Cannot verify signature: invalid key directory - " + errorMsg
		case string(appErrors.ErrMissingSignature):
			return "No cryptographic signature found on commit"
		case string(appErrors.ErrInvalidSignatureFormat):
			var errorMsg string

			for k, v := range errors[0].Context {
				if k == "error" {
					errorMsg = v

					break
				}
			}

			return "Invalid " + s.SignatureType + " signature format: " + errorMsg
		case string(appErrors.ErrKeyNotTrusted):
			return "Signature verified but the key is not in the trusted keys directory"
		case string(appErrors.ErrWeakKey):
			var bits, required string

			for k, v := range errors[0].Context {
				if k == "key_bits" {
					bits = v
				} else if k == "required_bits" {
					required = v
				}
			}

			return "Weak " + s.SignatureType + " key detected: " + bits + " bits (minimum required: " + required + " bits)"
		case string(appErrors.ErrVerificationFailed):
			var errorMsg string

			for k, v := range errors[0].Context {
				if k == "error" {
					errorMsg = v

					break
				}
			}

			return "Signature verification failed: " + errorMsg
		case string(appErrors.ErrUnknownSigFormat):
			return "Unknown signature type, cannot verify identity"
		default:
			return errors[0].Error()
		}
	}

	return fmt.Sprintf("Valid %s signature from %q", s.SignatureType, s.Identity)
}

// addError adds a structured validation error using the standard error system.
func (s *SignedIdentity) addError(code appErrors.ValidationErrorCode, message string, context map[string]string) {
	var err appErrors.ValidationError
	if context != nil {
		err = appErrors.New("SignedIdentity", code, message, appErrors.WithContextMap(context))
	} else {
		err = appErrors.New("SignedIdentity", code, message)
	}

	s.errors = append(s.errors, err)
}

// markAsRun marks the rule as having been run.
func (s *SignedIdentity) markAsRun() {
	s.hasRun = true
}

// Help returns a description of how to fix the rule violation.
func (s SignedIdentity) Help() string {
	if !s.HasErrors() {
		return "No errors to fix"
	}

	// First check for specific error codes
	errors := s.Errors()
	if len(errors) > 0 {
		switch errors[0].Code {
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
			keyType := errors[0].Context["key_type"]
			bits := errors[0].Context["key_bits"]
			required := errors[0].Context["required_bits"]

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

// SignedIdentityOption is a functional option for configuring SignedIdentity.
type SignedIdentityOption func(*SignedIdentity)

// WithAllowedSignatureTypes configures which signature types are allowed.
func WithAllowedSignatureTypes(allowedTypes ...string) SignedIdentityOption {
	return func(s *SignedIdentity) {
		if len(allowedTypes) > 0 {
			s.allowedTypes = allowedTypes
		}
	}
}

// WithKeyTypes configures which key types are allowed for each signature type.
func WithKeyTypes(keyTypes map[string][]string) SignedIdentityOption {
	return func(s *SignedIdentity) {
		if keyTypes != nil {
			s.allowedKeyTypes = keyTypes
		}
	}
}

// NewSignedIdentity creates a new SignedIdentity rule with the given key directory and allowed identity.
func NewSignedIdentity(keyDir string, allowedIdentity string, options ...SignedIdentityOption) *SignedIdentity {
	rule := &SignedIdentity{
		name:         "SignedIdentity",
		errors:       make([]appErrors.ValidationError, 0),
		KeyDir:       keyDir,
		Identity:     allowedIdentity, // Store the allowedIdentity in the rule's Identity field
		allowedTypes: []string{GPG, SSH},
		allowedKeyTypes: map[string][]string{
			GPG: {"rsa", "dsa", "ecdsa", "ed25519"},
			SSH: {"rsa", "dsa", "ecdsa", "ed25519"},
		},
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Validate implements the Rule interface for SignedIdentity.
func (s *SignedIdentity) Validate(commit *object.Commit) []appErrors.ValidationError {
	s.ClearErrors()

	if commit == nil {
		s.addError(appErrors.ErrCommitNil, "commit cannot be nil", nil)
		s.markAsRun()

		return s.errors
	}

	if s.KeyDir == "" {
		s.addError(appErrors.ErrNoKeyDir, "no key directory provided", nil)
		s.markAsRun()

		return s.errors
	}

	// Sanitize keyDir to prevent path traversal
	sanitizedKeyDir, err := SanitizePath(s.KeyDir)
	if err != nil {
		s.addError(
			appErrors.ErrInvalidKeyDir,
			fmt.Sprintf("invalid key directory: %s", err),
			map[string]string{
				"path": s.KeyDir,
			},
		)
		s.markAsRun()

		return s.errors
	}

	// Get the signature from the commit
	signature := commit.PGPSignature
	if signature == "" {
		s.addError(appErrors.ErrMissingSignature, "no signature provided", nil)
		s.markAsRun()

		return s.errors
	}

	// Verify the signature
	verificationResult := VerifySignatureIdentity(commit, signature, sanitizedKeyDir)
	s.errors = verificationResult.errors
	s.Identity = verificationResult.Identity
	s.SignatureType = verificationResult.SignatureType
	s.markAsRun()

	return s.errors
}

// ClearErrors clears all errors.
func (s *SignedIdentity) ClearErrors() {
	s.errors = make([]appErrors.ValidationError, 0)
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
		name:         "SignedIdentity",
		errors:       make([]appErrors.ValidationError, 0),
		KeyDir:       keyDir,
		allowedTypes: []string{GPG, SSH},
		allowedKeyTypes: map[string][]string{
			GPG: {"rsa", "dsa", "ecdsa", "ed25519"},
			SSH: {"rsa", "dsa", "ecdsa", "ed25519"},
		},
	}

	if commit == nil {
		rule.addError(
			appErrors.ErrCommitNil,
			"commit cannot be nil",
			nil,
		)

		return rule
	}

	if keyDir == "" {
		rule.addError(
			appErrors.ErrNoKeyDir,
			"no key directory provided",
			nil,
		)

		return rule
	}

	// Sanitize keyDir to prevent path traversal
	sanitizedKeyDir, err := SanitizePath(keyDir)
	if err != nil {
		rule.addError(
			appErrors.ErrInvalidKeyDir,
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
			appErrors.ErrMissingSignature,
			"no signature provided",
			nil,
		)

		return rule
	}

	// Get commit data
	commitBytes, err := getCommitBytes(commit)
	if err != nil {
		rule.addError(
			appErrors.ErrInvalidCommit,
			fmt.Sprintf("failed to prepare commit data: %s", err),
			map[string]string{
				"error": err.Error(),
			},
		)

		return rule
	}

	// Auto-detect signature type
	sigType := DetectSignatureType(signature)
	rule.SignatureType = sigType

	// Helper function to handle verification errors
	handleVerificationError := func(err error, sigType string) bool {
		if err == nil {
			return false
		}

		// Determine error type and add appropriate validation error
		if strings.Contains(err.Error(), "not verified with any trusted key") {
			rule.addError(
				appErrors.ErrKeyNotTrusted,
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
				appErrors.ErrWeakKey,
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
				appErrors.ErrVerificationFailed,
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
		identity, err := VerifyGPGSignature(commitBytes, signature, sanitizedKeyDir)
		if handleVerificationError(err, GPG) {
			return rule
		}

		rule.Identity = identity

	case SSH:
		// Parse SSH signature from string
		format, blob, err := ParseSSHSignature(signature)
		if err != nil {
			rule.addError(
				appErrors.ErrInvalidSignatureFormat,
				fmt.Sprintf("invalid SSH signature format: %s", err),
				map[string]string{
					"signature_type": SSH,
					"error":          err.Error(),
				},
			)

			return rule
		}

		identity, err := VerifySSHSignature(commitBytes, format, blob, sanitizedKeyDir)
		if handleVerificationError(err, SSH) {
			return rule
		}

		rule.Identity = identity

	default:
		rule.addError(
			appErrors.ErrUnknownSigFormat,
			"unknown signature type",
			map[string]string{
				"signature": signature,
			},
		)
	}

	rule.markAsRun()

	return rule
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

// The following functions are assumed to be defined elsewhere in the package:
// sanitizePath(path string) (string, error)
// getCommitBytes(commit *object.Commit) ([]byte, error)
// parseSSHSignature(signature string) (string, []byte, error)
// verifyGPGSignature(commitData []byte, signature string, keyDir string) (string, error)
// verifySSHSignature(commit []byte, format string, blob []byte, keyDir string) (string, error)
