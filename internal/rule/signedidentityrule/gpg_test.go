// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package signedidentityrule

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyGPGSignature(t *testing.T) {
	testDataDir, _ := filepath.Abs("testdata")
	tests := []struct {
		name        string
		setupOpts   setupRepoOptions
		keyDir      string
		expectError bool
		wantErrMsg  string
		wantID      string
	}{
		{
			name: "valid signed commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t),
			},
			keyDir:      testDataDir,
			expectError: false,
			wantID:      "Test User <test@example.com>",
		},
		{
			name: "unsigned commit",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Unsigned commit",
			},
			keyDir:      testDataDir,
			expectError: true,
			wantErrMsg:  "no signature provided",
		},
		{
			name: "missing key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t),
			},
			keyDir:      "",
			expectError: true,
			wantErrMsg:  "no key directory provided",
		},
		{
			name: "empty key directory",
			setupOpts: setupRepoOptions{
				authorName:  "Test User",
				authorEmail: "test@example.com",
				message:     "Signed commit",
				signKey:     loadTestKey(t),
			},
			keyDir:      t.TempDir(), // Empty directory
			expectError: true,
			wantErrMsg:  "no GPG key files found",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			_, commit := setupTestRepo(t, tabletest.setupOpts)

			// Test the convenience wrapper function
			result := VerifySignatureIdentity(commit, commit.PGPSignature, tabletest.keyDir)

			if tabletest.expectError {
				require.NotEmpty(t, result.Errors(), "Expected errors but got none")

				if tabletest.wantErrMsg != "" {
					require.Contains(t, result.Errors()[0].Error(), tabletest.wantErrMsg,
						"Error message doesn't contain expected text")
				}
			} else {
				require.Empty(t, result.Errors(), "Expected no errors but got: %v", result.Errors())
				require.Equal(t, tabletest.wantID, result.Identity, "Identity doesn't match expected value")
				require.Equal(t, "GPG", result.SignatureType, "Signature type should be GPG")
			}
		})
	}
}

// Add specific tests for GPG-related functions.
func TestIsKeyRevoked(t *testing.T) {
	// This test would require mock entity setup with revocation information
	// Implementing a basic structure here that you can expand with actual test data
	t.Run("not revoked key", func(t *testing.T) {
		// Skip test for now since we need proper test keys with revocation data
		// Example of what a real test would look like:
		// entity := loadTestKey(t, "valid.priv")
		// require.False(t, isKeyRevoked(entity), "Key should not be revoked")
		t.Skip("Requires proper GPG key revocation mocks")
	})
}

func TestIsKeyExpired(t *testing.T) {
	// This would test the key expiration logic with different time values
	t.Run("not expired key", func(t *testing.T) {
		// Skip test for now since we need proper test keys with expiration data
		// Example of what a real test might look like:
		// entity := loadTestKey(t, "valid.priv")
		// pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		// require.False(t, isKeyExpired(entity, pastTime), "Key should not be expired in the past")
		t.Skip("Requires proper GPG key expiration mocks")
	})
}

func TestHasMinimumKeyStrength(t *testing.T) {
	// This would test the key strength validation logic
	t.Run("strong enough key", func(t *testing.T) {
		// Skip test for now since we need proper test keys with known strengths
		// Example of what a real test might look like:
		// entity := loadTestKey(t, "valid.priv")
		// require.True(t, hasMinimumKeyStrength(entity), "Key should meet minimum strength requirements")
		t.Skip("Requires proper GPG key strength mocks")
	})
}
