// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package crypto provides core cryptographic verification implementations.

This package contains the core logic for verifying commit signatures,
organized by signature type:

  - gpg: GPG/OpenPGP signature verification
  - ssh: SSH signature verification
  - common: Shared cryptographic utilities

Each implementation handles the specific details of its signature format
while providing a consistent interface for signature verification. The
implementations validate:

  - Signature authenticity
  - Key strength and algorithm security
  - Key expiration and revocation status
  - Identity binding

All cryptographic operations follow security best practices with
appropriate key strength requirements and algorithm validation.
*/
package crypto
