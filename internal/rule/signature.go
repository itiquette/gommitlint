// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

// Package rule provides validation rules for commit message linting.
// This file contains the Signature rule, which enforces cryptographic
// signing of commits to enhance security and verify author identity.
package rule

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// Signature enforces that commits are cryptographically signed using SSH or GPG keys.
// This rule helps ensure security and verifiability in the git history by requiring
// all commits to be properly signed by their authors.
//
// A valid signature confirms that the commit was created by someone with access to
// the private key corresponding to a trusted public key, providing authentication
// and non-repudiation for the commit history.
//
// For example, commits created with 'git commit -S' (GPG) or 'git commit -s' (SSH)
// would pass validation, while unsigned commits would fail.
//
// IMPORTANT: This rule only checks for the presence and basic format of signatures.
// It does NOT verify the cryptographic validity of signatures or check if they were
// created with trusted keys. For full security, additional verification mechanisms
// should be used. You can use the signedidentityrule for that.
type Signature struct {
	errors []error
}

// Name returns the name of the rule.
func (Signature) Name() string {
	return "Signature"
}

// Result returns validation results as a human-readable string.
// If no errors are found, it returns a success message indicating
// that a valid signature was found.
func (rule Signature) Result() string {
	if len(rule.errors) != 0 {
		return rule.errors[0].Error()
	}

	return "SSH/GPG signature found (format verified only, not cryptographically validated)"
}

// Errors returns any violations of the rule.
func (rule Signature) Errors() []error {
	return rule.errors
}

// Help returns a description of how to fix the rule violation.
func (rule Signature) Help() string {
	if len(rule.errors) == 0 {
		return `No errors to fix

Note: This rule only checks that a signature exists and has valid formatting.
It does NOT verify the cryptographic validity of the signature or that it was 
created by a trusted key. For full security, additional verification is required.`
	}

	return `Sign your commit using SSH or GPG to verify your identity.

You can do this by:
1. Setting up GPG signing:
   - Generate a GPG key if you don't have one: 'gpg --gen-key'
   - Configure Git to use your key: 'git config --global user.signingkey YOUR_KEY_ID'
   - Sign commits with: 'git commit -S' or set automatic signing: 'git config --global commit.gpgsign true'

2. Setting up SSH signing:
   - Configure Git to use your SSH key: 'git config --global gpg.format ssh'
   - Set your signing key: 'git config --global user.signingkey ~/.ssh/id_ed25519.pub'
   - Sign commits with: 'git commit -S'

Note: This rule only checks for the existence and basic format of a signature.
For full security, Git or your Git platform should verify signatures cryptographically.

For more information, see:
- GPG signing: https://git-scm.com/book/en/v2/Git-Tools-Signing-Your-Work
- SSH signing: https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification`
}

// addErrorf adds an error to the rule's errors slice.
func (rule *Signature) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}

// ValidateSignature checks if the commit has a valid cryptographic signature.
// It returns a Signature rule with any validation errors.
//
// The function accepts a signature string from the Git commit metadata and validates:
// 1. That a signature exists (not empty)
// 2. That it follows either GPG or SSH signature format using library validation
// 3. That the signature has the correct structure and encoding
//
// This validation only checks the format, not cryptographic validity against a public key.
// It uses the ProtonMail Go OpenPGP library for GPG signature validation.
func ValidateSignature(signature string) *Signature {
	rule := &Signature{}

	// Check for empty signature
	if signature == "" || len(strings.TrimSpace(signature)) == 0 {
		rule.addErrorf("commit does not have a SSH/GPG signature")

		return rule
	}

	// Trim whitespace for validation
	signature = strings.TrimSpace(signature)

	// Check for GPG signature
	if strings.HasPrefix(signature, "-----BEGIN PGP SIGNATURE-----") {
		if !strings.Contains(signature, "-----END PGP SIGNATURE-----") {
			rule.addErrorf("incomplete GPG signature (missing end marker)")

			return rule
		}

		// Use ProtonMail's openpgp library to validate the format
		block, err := armor.Decode(strings.NewReader(signature))
		if err != nil {
			rule.addErrorf("invalid GPG signature format: %v", err)

			return rule
		}

		if block.Type != "PGP SIGNATURE" {
			rule.addErrorf("invalid GPG armor block type: %s", block.Type)

			return rule
		}

		// We don't verify the signature cryptographically, just check it can be read
		reader := packet.NewReader(block.Body)

		_, err = reader.Next()
		if err != nil {
			rule.addErrorf("malformed GPG signature content: %v", err)

			return rule
		}

		return rule // Valid GPG signature format
	}

	// Check for SSH signature
	if strings.HasPrefix(signature, "-----BEGIN SSH SIGNATURE-----") {
		// Use SSH-specific validation
		err := verifySSHSignatureFormat(signature)
		if err != nil {
			rule.addErrorf("%v", err)

			return rule
		}

		return rule // Valid SSH signature format
	}

	// Not a recognized signature format
	rule.addErrorf("invalid signature format (must be GPG or SSH)")

	return rule
}

// verifySSHSignatureFormat performs basic validation of SSH signature format
// without attempting to parse the detailed binary structure.
func verifySSHSignatureFormat(signature string) error {
	if !strings.HasPrefix(signature, "-----BEGIN SSH SIGNATURE-----") {
		return errors.New("missing SSH signature begin marker")
	}

	if !strings.Contains(signature, "-----END SSH SIGNATURE-----") {
		return errors.New("incomplete SSH signature (missing end marker)")
	}

	// Extract the base64-encoded content
	content := signature
	content = strings.TrimPrefix(content, "-----BEGIN SSH SIGNATURE-----")
	content = strings.TrimSuffix(content, "-----END SSH SIGNATURE-----")
	content = strings.TrimSpace(content)

	// Validate the base64 encoding
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("invalid SSH signature encoding: %w", err)
	}

	// Basic structure checks

	// Check for the SSHSIG prefix
	if len(data) < 6 || !bytes.Equal(data[:6], []byte("SSHSIG")) {
		return errors.New("malformed SSH signature: missing SSHSIG prefix")
	}

	// A valid SSH signature should have reasonable length
	if len(data) < 50 {
		return errors.New("malformed SSH signature: too short")
	}

	// We'll avoid detailed binary parsing which can be fragile
	// and just check the overall structure is reasonable

	return nil
}
