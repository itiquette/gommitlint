// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package sigverify

import (
	"errors"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// TestVerificationResult tests the basic properties of a verification result.
func TestVerificationResult(t *testing.T) {
	// Create a new empty verification result
	result := NewVerificationResult()

	// Check default values
	require.Empty(t, result.Errors, "New result should have no errors")
	require.Empty(t, result.Identity, "New result should have empty identity")
	require.Empty(t, result.SignatureType, "New result should have empty signature type")
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := DetectSignatureType(testCase.signature)
			require.Equal(t, testCase.expected, result)
		})
	}
}

// TestVerifySignature tests the main signature verification function.
func TestVerifySignature(t *testing.T) {
	tests := []struct {
		name          string
		commit        *object.Commit
		keyDir        string
		expectErrors  bool
		expectedType  string
		expectedIdent string
	}{
		{
			name:         "Nil commit",
			commit:       nil,
			keyDir:       "/tmp/keys",
			expectErrors: true,
		},
		{
			name:         "Empty key directory",
			commit:       &object.Commit{},
			keyDir:       "",
			expectErrors: true,
		},
		{
			name: "Missing signature",
			commit: &object.Commit{
				PGPSignature: "",
			},
			keyDir:       "/tmp/keys",
			expectErrors: true,
		},
		// We can't easily test successful verification without mocking the underlying functions
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := VerifySignature(testCase.commit, testCase.keyDir)

			if testCase.expectErrors {
				require.NotEmpty(t, result.Errors, "Should have validation errors")
			} else {
				require.Empty(t, result.Errors, "Should not have validation errors")
				require.Equal(t, testCase.expectedType, result.SignatureType)
				require.Equal(t, testCase.expectedIdent, result.Identity)
			}
		})
	}
}

// TestHandleVerificationError tests the error handling function.
func TestHandleVerificationError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		sigType   string
		errorCode string
	}{
		{
			name:      "Key not trusted error",
			err:       errors.New("signature not verified with any trusted key"),
			sigType:   GPG,
			errorCode: string(appErrors.ErrKeyNotTrusted),
		},
		{
			name:      "Weak key error",
			err:       errors.New("insufficient key strength: 1024 bits (required: 2048 bits)"),
			sigType:   GPG,
			errorCode: string(appErrors.ErrWeakKey),
		},
		{
			name:      "Generic verification error",
			err:       errors.New("verification failed"),
			sigType:   GPG,
			errorCode: string(appErrors.ErrVerificationFailed),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			valErr := handleVerificationError(testCase.err, testCase.sigType)
			require.Equal(t, testCase.errorCode, valErr.Code)
			require.Contains(t, valErr.Context, "signature_type")
			require.Equal(t, testCase.sigType, valErr.Context["signature_type"])
		})
	}
}
