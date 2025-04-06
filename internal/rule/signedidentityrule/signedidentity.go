// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signedidentityrule

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
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
	errors        []error
	Identity      string // Email or name of the signer
	SignatureType string // "GPG" or "SSH"
}

// Name returns the rule identifier.
func (s SignedIdentity) Name() string {
	return "SignedIdentity"
}

// Result returns a string representation of the rule's status.
func (s SignedIdentity) Result() string {
	if len(s.errors) > 0 {
		return s.errors[0].Error()
	}

	return fmt.Sprintf("Signed by %q using %s", s.Identity, s.SignatureType)
}

// Errors returns any violations detected by the rule.
func (s SignedIdentity) Errors() []error {
	return s.errors
}

// Help returns a description of how to fix the rule violation.
func (s SignedIdentity) Help() string {
	if len(s.errors) == 0 {
		return "No errors to fix"
	}

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
func VerifySignatureIdentity(commit *object.Commit, signature string, keyDir string) SignedIdentity {
	rule := SignedIdentity{}

	if commit == nil {
		rule.errors = append(rule.errors, errors.New("commit cannot be nil"))

		return rule
	}

	if keyDir == "" {
		rule.errors = append(rule.errors, errors.New("no key directory provided"))

		return rule
	}

	// Sanitize keyDir to prevent path traversal
	sanitizedKeyDir, err := sanitizePath(keyDir)
	if err != nil {
		rule.errors = append(rule.errors, fmt.Errorf("invalid key directory: %w", err))

		return rule
	}

	if signature == "" {
		rule.errors = append(rule.errors, errors.New("no signature provided"))

		return rule
	}

	// Get commit data
	commitBytes, err := getCommitBytes(commit)
	if err != nil {
		rule.errors = append(rule.errors, fmt.Errorf("failed to prepare commit data: %w", err))

		return rule
	}

	// Auto-detect signature type
	sigType := detectSignatureType(signature)

	// Verify based on signature type
	switch sigType {
	case GPG:
		identity, err := verifyGPGSignature(commitBytes, signature, sanitizedKeyDir)
		if err != nil {
			rule.errors = append(rule.errors, err)

			return rule
		}

		rule.Identity = identity
		rule.SignatureType = GPG

	case SSH:
		// Parse SSH signature from string
		format, blob, err := parseSSHSignature(signature)
		if err != nil {
			rule.errors = append(rule.errors, fmt.Errorf("invalid SSH signature format: %w", err))

			return rule
		}

		identity, err := verifySSHSignature(commitBytes, format, blob, sanitizedKeyDir)
		if err != nil {
			rule.errors = append(rule.errors, err)

			return rule
		}

		rule.Identity = identity
		rule.SignatureType = SSH

	default:
		rule.errors = append(rule.errors, errors.New("unknown signature type"))
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
