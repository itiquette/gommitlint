// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package signing provides cryptographic signature verification adapters.

This package implements domain interfaces for verifying commit signatures,
isolating cryptographic complexity from the core domain logic.

Key components:

  - verification.go: Main verification logic and domain interface implementation
  - gpg.go: GPG/OpenPGP signature verification
  - ssh.go: SSH signature verification
  - files.go: Secure file operations for key management
  - repository.go: Key repository abstraction

The package validates:

  - Signature authenticity and integrity
  - Key strength and algorithm security
  - Key expiration and revocation status
  - Identity binding between signatures and committers

All cryptographic operations follow security best practices with appropriate
key strength requirements and algorithm validation.
*/
package signing
