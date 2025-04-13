// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package signedidentityrule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSSHSignature(t *testing.T) {
	tests := []struct {
		name        string
		signature   string
		expectError bool
		wantFormat  string
	}{
		{
			name:        "valid SSH signature",
			signature:   "ssh-rsa:AAAAB3NzaC1yc2EAAAAD",
			expectError: false,
			wantFormat:  "ssh-rsa",
		},
		{
			name:        "valid ed25519 signature",
			signature:   "ssh-ed25519:AAAAC3NzaC1lZDI1NTE5",
			expectError: false,
			wantFormat:  "ssh-ed25519",
		},
		{
			name:        "invalid format without separator",
			signature:   "ssh-rsaAAAAB3NzaC1yc2EAAAAD",
			expectError: true,
		},
		{
			name:        "invalid base64 blob",
			signature:   "ssh-rsa:not-base64-data",
			expectError: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			format, blob, err := parseSSHSignature(tabletest.signature)

			if tabletest.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tabletest.wantFormat, format)
				require.NotEmpty(t, blob)
			}
		})
	}
}

// Add more SSH-specific tests.
func TestSSHKeyHasMinimumStrength(t *testing.T) {
	// This would test the SSH key strength validation
	t.Run("strong enough key", func(t *testing.T) {
		// Skip test for now since we need proper SSH key mocks
		// Example of what a real test might look like:
		// keyData := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADA...")  // Mock RSA key data
		// pubKey, err := ssh.ParsePublicKey(keyData)
		// require.NoError(t, err)
		// require.True(t, sshKeyHasMinimumStrength(pubKey), "SSH key should meet minimum strength requirements")
		t.Skip("Requires SSH key mocks with known strengths")
	})
}

func TestVerifySSHSignature(t *testing.T) {
	// This would test SSH signature verification
	t.Run("valid SSH signature verification", func(t *testing.T) {
		// Skip test for now as it requires more complex test setup
		// Example structure for a real test:
		// 1. Create test commit data
		// 2. Create valid SSH signature for that data
		// 3. Set up mock/test SSH key in a temp directory
		// 4. Call verifySSHSignature
		// 5. Verify correct identity returned and no error
		t.Skip("Requires SSH signature test data")
	})
}

func TestFindSSHKeyFiles(t *testing.T) {
	// Currently commented out in the original test file
	// This test would verify the SSH key finding functionality
	// Example of what a complete test might look like:
	// tempDir := t.TempDir()
	//
	//	files := []struct {
	//		name    string
	//		content string
	//	}{
	//
	//		{name: "id_rsa.pub", content: "ssh-rsa AAAAB3NzaC1yc2E test-key"},
	//		{name: "id_ed25519.pub", content: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5 test-key"},
	//		{name: "gpg_key.pub", content: "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: 1\nContent\n-----END PGP PUBLIC KEY BLOCK-----"},
	//		{name: "custom.ssh", content: "ssh-rsa AAAAB3NzaC1yc2E custom-key"},
	//	}
	//
	//	for _, file := range files {
	//		err := os.WriteFile(filepath.Join(tempDir, file.name), []byte(file.content), 0600)
	//		require.NoError(t, err)
	//	}
	//
	// sshKeys, err := findSSHKeyFiles(tempDir)
	// require.NoError(t, err)
	// require.Len(t, sshKeys, 3)  // Should find id_rsa.pub, id_ed25519.pub, and custom.ssh
	t.Skip("SSH key finding test needs to be implemented")
}
