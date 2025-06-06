// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/adapters/signing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemKeyRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "key-repo-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create some test files with different extensions
	files := []string{
		"key1.gpg",
		"key2.pub",
		"key3.txt", // Should not match
		"key4.ssh",
	}

	for _, file := range files {
		path := filepath.Join(tempDir, file)
		err := os.WriteFile(path, []byte("test"), 0600)
		require.NoError(t, err)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file in the subdirectory (should be ignored)
	subFile := filepath.Join(subDir, "subkey.gpg")
	err = os.WriteFile(subFile, []byte("test"), 0600)
	require.NoError(t, err)

	// Create the repository
	repo := signing.NewFileSystemKeyRepository(tempDir)

	// Test GetKeyDirectory
	require.Equal(t, tempDir, repo.GetKeyDirectory())

	// Test FindKeyFiles with specific extensions
	extensions := []string{".gpg", ".pub"}
	foundFiles, err := repo.FindKeyFiles(extensions)
	require.NoError(t, err)
	require.Len(t, foundFiles, 2)

	// Check that we found the right files
	expectedPaths := []string{
		filepath.Join(tempDir, "key1.gpg"),
		filepath.Join(tempDir, "key2.pub"),
	}

	// Check if all expected paths are in the found files
	for _, expected := range expectedPaths {
		found := false

		for _, actual := range foundFiles {
			if expected == actual {
				found = true

				break
			}
		}

		require.True(t, found, "Expected file %s not found", expected)
	}

	// Test with different extensions
	sshExtensions := []string{".ssh"}
	sshFiles, err := repo.FindKeyFiles(sshExtensions)
	require.NoError(t, err)
	require.Len(t, sshFiles, 1)
	require.Equal(t, filepath.Join(tempDir, "key4.ssh"), sshFiles[0])

	// Test with non-existent extensions
	noFiles, err := repo.FindKeyFiles([]string{".nonexistent"})
	require.NoError(t, err)
	require.Empty(t, noFiles)
}

func TestReadKeyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "key-repo-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create the repository
	repo := signing.NewFileSystemKeyRepository(tempDir)

	// Create test files with different permissions and types
	tests := []struct {
		name        string
		filename    string
		permissions os.FileMode
		content     string
		shouldPass  bool
	}{
		{
			name:        "Private key with secure permissions",
			filename:    "private.key",
			permissions: 0600,
			content:     "secure private key content",
			shouldPass:  true,
		},
		{
			name:        "Private key with insecure permissions",
			filename:    "insecure-private.key",
			permissions: 0644,
			content:     "insecure private key content",
			shouldPass:  false, // Private keys need 0600 permissions
		},
		{
			name:        "Public key with reasonable permissions",
			filename:    "public.pub",
			permissions: 0640,
			content:     "public key content",
			shouldPass:  true,
		},
		{
			name:        "Public key with world-writable permissions",
			filename:    "world-writable.pub",
			permissions: 0666,
			content:     "world-writable public key content",
			shouldPass:  false, // World-writable is insecure
		},
		{
			name:        "Regular file with normal permissions",
			filename:    "regular.txt",
			permissions: 0644,
			content:     "regular file content",
			shouldPass:  true,
		},
	}

	// Create all test files
	files := make([]string, 0, len(tests))

	for _, testCase := range tests {
		filePath := filepath.Join(tempDir, testCase.filename)
		err := os.WriteFile(filePath, []byte(testCase.content), testCase.permissions)
		require.NoError(t, err)

		files = append(files, filePath)
	}

	// Test each file
	for i, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			filePath := files[i]
			content, err := repo.ReadKeyFile(filePath)

			if testCase.shouldPass {
				require.NoError(t, err)
				assert.Equal(t, testCase.content, string(content))
			} else {
				assert.Error(t, err)
			}
		})
	}
}
