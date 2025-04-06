// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package signedidentityrule provides cryptographic signature verification for Git
commits to enhance security and establish authorship.

This package implements a comprehensive validation system for both GPG and SSH
signatures on Git commits. It verifies that commits are cryptographically signed
by trusted keys, establishing a secure chain of authorship and preventing
unauthorized code modifications.

The package offers:

  - Automatic detection of signature types (GPG or SSH)
  - Validation against trusted public keys stored in a specified directory
  - Security checks for key strength, expiration, and revocation status
  - Support for multiple key formats and encodings

Key components include:

  - SignedIdentity: The main rule structure that validates commit signatures
    against a set of trusted keys.

  - VerifySignatureIdentity: The main validation function that detects signature
    type and dispatches to the appropriate verification method.

  - Helper functions for GPG and SSH signature verification, key loading,
    and security validation.

The rule enforces the following security policies:

  - RSA keys must meet minimum bit length requirements (default: 2048 bits)
  - EC keys must meet minimum security requirements (default: 256 bits)
  - Expired or revoked keys are rejected
  - Only recognized signature formats are accepted

Note: This package is being gradually migrated to the main "rule" package.
New code should use the equivalent functionality in package "rule" instead.

Example Usage:

	repo := openRepository(".")
	commit, _ := repo.HeadCommit()
	signature := commit.PGPSignature

	// Verify the signature against trusted keys
	rule := signedidentityrule.VerifySignatureIdentity(commit, signature, "/path/to/trusted/keys")
	if len(rule.Errors()) > 0 {
	    fmt.Println(rule.Help())
	} else {
	    fmt.Printf("Commit verified: %s\n", rule.Result())
	}
*/
package signedidentityrule
