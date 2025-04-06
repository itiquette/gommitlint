// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package signedidentityrule

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindKeyFiles(t *testing.T) {
	// Create a temp directory with test key files
	tempDir := t.TempDir()

	// Create some test files
	files := []struct {
		name    string
		content string
	}{
		{name: "key1.pub", content: "ssh-rsa AAAAB3NzaC1yc2E test-key"},
		{name: "key2.gpg", content: "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: 1\nContent\n-----END PGP PUBLIC KEY BLOCK-----"},
		{name: "key3.asc", content: "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: 1\nContent\n-----END PGP PUBLIC KEY BLOCK-----"},
		{name: "not-a-key.txt", content: "This is not a key file"},
	}

	for _, file := range files {
		err := os.WriteFile(filepath.Join(tempDir, file.name), []byte(file.content), 0600)
		require.NoError(t, err)
	}

	// Test finding GPG key files
	gpgKeyFiles, err := findKeyFiles(tempDir, []string{".pub", ".gpg", ".asc"}, GPG)
	require.NoError(t, err)
	require.Len(t, gpgKeyFiles, 2) // Should find key2.gpg and key3.asc

	// Test finding SSH key files
	sshKeyFiles, err := findKeyFiles(tempDir, []string{".pub"}, SSH)
	require.NoError(t, err)
	require.Len(t, sshKeyFiles, 1) // Should find key1.pub

	// Test with non-existent directory
	_, err = findKeyFiles("/non-existent-dir", []string{".pub"}, SSH)
	require.Error(t, err)
}

func TestSanitizePath(t *testing.T) {
	// Test valid directory
	tempDir := t.TempDir()
	sanitized, err := sanitizePath(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, sanitized)

	// Test non-existent directory
	_, err = sanitizePath("/this/path/does/not/exist")
	require.Error(t, err)

	// Test file (not a directory)
	tempFile := filepath.Join(tempDir, "testfile")
	err = os.WriteFile(tempFile, []byte("test"), 0600)
	require.NoError(t, err)
	_, err = sanitizePath(tempFile)
	require.Error(t, err)
	require.Contains(t, err.Error(), "is not a directory")
}

func TestIsValidKeyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create test key files
	files := []struct {
		name    string
		content string
		isValid bool
		keyType string
	}{
		{
			name:    "valid_gpg.gpg",
			content: "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: 1\nContent\n-----END PGP PUBLIC KEY BLOCK-----",
			isValid: true,
			keyType: GPG,
		},
		{
			name:    "invalid_gpg.gpg",
			content: "Not a GPG key file",
			isValid: false,
			keyType: GPG,
		},
		{
			name:    "valid_ssh.pub",
			content: "ssh-rsa AAAAB3NzaC1yc2E test-key",
			isValid: true,
			keyType: SSH,
		},
		{
			name:    "invalid_ssh.pub",
			content: "Not an SSH key file",
			isValid: false,
			keyType: SSH,
		},
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file.name)
		err := os.WriteFile(filePath, []byte(file.content), 0600)
		require.NoError(t, err)

		t.Run(file.name, func(t *testing.T) {
			result := isValidKeyFile(filePath, file.keyType)
			require.Equal(t, file.isValid, result)
		})
	}
}

func TestGetCommitBytes(t *testing.T) {
	// This test depends on setupTestRepo helper
	t.Run("get commit bytes", func(t *testing.T) {
		// Create a test repo and commit
		opts := setupRepoOptions{
			authorName:  "Test User",
			authorEmail: "test@example.com",
			message:     "Test commit",
		}

		_, commit := setupTestRepo(t, opts)

		// Get bytes from commit
		bytes, err := getCommitBytes(commit)
		require.NoError(t, err)
		require.NotEmpty(t, bytes)
	})
}
