// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/fsutils"
	"github.com/stretchr/testify/require"
)

func TestSanitizeKeyDirectory(t *testing.T) {
	// Note: This test now uses fsutils.SanitizePath directly
	// as the crypto/common.SanitizeKeyDirectory function has been moved
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gommitlint-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test with a valid directory
	sanitized, err := fsutils.SanitizePath(tempDir)
	require.NoError(t, err)
	require.NotEmpty(t, sanitized)

	// Test with nonexistent directory
	_, err = fsutils.SanitizePath(filepath.Join(tempDir, "nonexistent"))
	require.Error(t, err)

	// Create a file to test with a non-directory path
	filePath := filepath.Join(tempDir, "testfile")
	err = os.WriteFile(filePath, []byte("test"), 0600)
	require.NoError(t, err)

	// Test with a file path instead of a directory
	_, err = fsutils.SanitizePath(filePath)
	require.Error(t, err)
}

func TestFindKeyFiles(t *testing.T) {
	// Note: This test now uses fsutils.FindFilesWithExtensions directly
	// as the crypto/common.FindKeyFiles function has been moved
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gommitlint-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files with different extensions
	err = os.WriteFile(filepath.Join(tempDir, "key1.pub"), []byte("test"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "key2.gpg"), []byte("test"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "key3.txt"), []byte("test"), 0600)
	require.NoError(t, err)

	// Test finding files with specific extensions
	files, err := fsutils.FindFilesWithExtensions(tempDir, []string{".pub", ".gpg"})
	require.NoError(t, err)
	require.Len(t, files, 2)

	// Verify the files found
	foundPub := false
	foundGpg := false

	for _, file := range files {
		if filepath.Ext(file) == ".pub" {
			foundPub = true
		}

		if filepath.Ext(file) == ".gpg" {
			foundGpg = true
		}
	}

	require.True(t, foundPub, "Expected to find .pub file")
	require.True(t, foundGpg, "Expected to find .gpg file")

	// Test with nonexistent directory
	_, err = fsutils.FindFilesWithExtensions(filepath.Join(tempDir, "nonexistent"), []string{".pub"})
	require.Error(t, err)
}

func TestIsValidKeyPath(t *testing.T) {
	// Note: This test now uses fsutils.IsWithinDirectory directly
	// as the crypto/common.IsValidKeyPath function has been moved
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gommitlint-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file in the directory
	validPath := filepath.Join(tempDir, "key.pub")
	err = os.WriteFile(validPath, []byte("test"), 0600)
	require.NoError(t, err)

	// Test with a valid path inside the directory
	isValid, err := fsutils.IsWithinDirectory(validPath, tempDir)
	require.NoError(t, err)
	require.True(t, isValid)

	// Test with a path outside the directory
	outsideDir, err := os.MkdirTemp("", "gommitlint-outside-*")
	require.NoError(t, err)
	defer os.RemoveAll(outsideDir)

	outsidePath := filepath.Join(outsideDir, "key.pub")
	err = os.WriteFile(outsidePath, []byte("test"), 0600)
	require.NoError(t, err)

	isValid, err = fsutils.IsWithinDirectory(outsidePath, tempDir)
	require.NoError(t, err)
	require.False(t, isValid)
}

func TestSafeJoinKeyPath(t *testing.T) {
	// Note: This test now uses fsutils.SafeJoin directly
	// as the crypto/common.SafeJoinKeyPath function has been moved
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gommitlint-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test joining a simple filename
	joined, err := fsutils.SafeJoin(tempDir, "key.pub")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "key.pub"), joined)

	// Test with a path traversal attempt
	_, err = fsutils.SafeJoin(tempDir, "../key.pub")
	require.Error(t, err)

	// Test with a nonexistent directory
	_, err = fsutils.SafeJoin(filepath.Join(tempDir, "nonexistent"), "key.pub")
	require.Error(t, err)
}
