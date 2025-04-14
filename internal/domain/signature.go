// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"fmt"
	"strings"
)

// SignatureType represents the type of cryptographic signature.
type SignatureType string

// Known signature types.
const (
	SignatureTypeGPG     SignatureType = "gpg"
	SignatureTypeSSH     SignatureType = "ssh"
	SignatureTypeUnknown SignatureType = "unknown"
)

// Signature represents a cryptographic signature with its format and data.
type Signature struct {
	sigType SignatureType
	data    string
}

// NewSignature creates a new signature from the given data.
// It automatically detects the signature type.
func NewSignature(data string) Signature {
	return Signature{
		sigType: detectSignatureType(data),
		data:    data,
	}
}

// NewSignatureWithType creates a new signature with the specified type.
func NewSignatureWithType(data string, sigType SignatureType) Signature {
	return Signature{
		sigType: sigType,
		data:    data,
	}
}

// Type returns the type of the signature.
func (s Signature) Type() SignatureType {
	return s.sigType
}

// Data returns the raw signature data.
func (s Signature) Data() string {
	return s.data
}

// IsEmpty returns true if the signature data is empty.
func (s Signature) IsEmpty() bool {
	return s.data == ""
}

// IsValid returns true if the signature has a known type and non-empty data.
func (s Signature) IsValid() bool {
	return s.sigType != SignatureTypeUnknown && !s.IsEmpty()
}

// String returns a string representation of the signature.
func (s Signature) String() string {
	return fmt.Sprintf("%s signature: %s...", s.sigType, truncate(s.data, 20))
}

// detectSignatureType determines the type of signature from its content.
func detectSignatureType(signature string) SignatureType {
	signature = strings.TrimSpace(signature)

	// Empty signature check
	if signature == "" {
		return SignatureTypeUnknown
	}

	// Check for GPG signature with headers
	if strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") {
		return SignatureTypeGPG
	}

	// Check for SSH signature with headers
	if strings.Contains(signature, "-----BEGIN SSH SIGNATURE-----") ||
		strings.HasPrefix(signature, "-----BEGIN SSH SIGN") ||
		strings.Contains(signature, "SSH SIGNATURE") ||
		strings.Contains(signature, "ssh-rsa") ||
		strings.Contains(signature, "ssh-ed25519") {
		return SignatureTypeSSH
	}

	// Check for SSH signature format (format:blob)
	if strings.Contains(signature, ":") {
		parts := strings.SplitN(signature, ":", 2)
		if len(parts) == 2 {
			prefix := strings.ToLower(parts[0])
			if strings.HasPrefix(prefix, "ssh-") ||
				strings.HasPrefix(prefix, "ecdsa-") ||
				strings.HasPrefix(prefix, "sk-") {
				return SignatureTypeSSH
			}
		}
	}

	// Check for common GPG signature patterns (base64-encoded data)
	// GPG signatures typically start with certain prefixes in their base64 form
	if strings.HasPrefix(signature, "iQEcBAA") || // Common GPG signature prefix
		strings.HasPrefix(signature, "iQIcBAA") ||
		strings.HasPrefix(signature, "iQA") ||
		strings.HasPrefix(signature, "iQI") ||
		strings.HasPrefix(signature, "iQEz") {
		return SignatureTypeGPG
	}

	// If non-empty but unrecognized, return unknown
	return SignatureTypeUnknown
}

// truncate shortens a string to the specified length, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "..."
}
