// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package sigverify

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignedIdentity_Name(t *testing.T) {
	// Simple test for rule name
	rule := SignedIdentity{}
	require.Equal(t, "SignedIdentity", rule.Name())
}

// TestDetectSignatureType tests the signature type detection function.
func TestDetectSignatureType(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		expected  string
	}{
		{
			name:      "GPG signature with PGP header",
			signature: "-----BEGIN PGP SIGNATURE-----\nVersion: GnuPG v2\nData\n-----END PGP SIGNATURE-----",
			expected:  GPG,
		},
		{
			name:      "SSH RSA signature format",
			signature: "ssh-rsa:AAAAB3NzaC1yc2EAAA...",
			expected:  SSH,
		},
		{
			name:      "SSH ed25519 signature format",
			signature: "ssh-ed25519:AAAAC3NzaC1lZDI1NTE5AAAA...",
			expected:  SSH,
		},
		{
			name:      "ECDSA SSH signature format",
			signature: "ecdsa-sha2-nistp256:AAAAE2VjZHNhLXNoYTItbmlzdHA...",
			expected:  SSH,
		},
		{
			name:      "Unknown format defaulting to GPG",
			signature: "unknown-signature-format",
			expected:  GPG,
		},
	}

	for _, atest := range tests {
		t.Run(atest.name, func(t *testing.T) {
			result := DetectSignatureType(atest.signature)
			require.Equal(t, atest.expected, result)
		})
	}
}

// TestValidateSignature checks basic initialization of the SignedIdentity rule.
func TestValidateSignature(t *testing.T) {
	// Create a rule with default settings
	rule := NewSignedIdentity("", "")

	// Check rule name
	require.Equal(t, "SignedIdentity", rule.Name())
}
