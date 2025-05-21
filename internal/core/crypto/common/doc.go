// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package common provides utility functions for cryptographic operations
// used by both GPG and SSH signature verification modules.
//
// The package focuses on secure path handling for keys and signatures,
// as well as encoding utilities for signature data processing. All
// functions follow functional programming principles with value semantics,
// ensuring immutability and predictable behavior.
//
// Key features include:
// - Secure path handling for cryptographic key files
// - Path traversal attack prevention
// - Base64 encoding/decoding utilities
// - Armored text detection and processing
//
// This package is designed to be used only by the crypto implementation
// packages and should not be used directly by rules or application code.
package common
