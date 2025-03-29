// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
// Package signedidentity provides commit signature verification functionality
package signedidentityrule

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

var MinimumRSABits uint16 = 2048
var MinimumECBits uint16 = 256

const SSH = "SSH"
const GPG = "GPG"

// SignedIdentity validates that a commit is properly signed with either GPG or SSH.
type SignedIdentity struct {
	errors        []error
	Identity      string
	SignatureType string // "GPG" or "SSH"
}

// Name returns the rule identifier.
func (rule SignedIdentity) Name() string {
	return "SignedIdentityRule"
}

// Result returns a string representation of the validation result.
func (rule SignedIdentity) Result() string {
	if len(rule.errors) > 0 {
		return rule.errors[0].Error()
	}

	return fmt.Sprintf("Signed by %q using %s", rule.Identity, rule.SignatureType)
}

// Errors returns any validation errors.
func (rule SignedIdentity) Errors() []error {
	return rule.errors
}

// VerifyCommitSignature checks if a commit is signed with a trusted key.
// It automatically detects whether the signature is GPG or SSH based on its format.
func VerifyCommitSignature(commit *object.Commit, signature string, keyDir string) SignedIdentity {
	rule := SignedIdentity{}

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
