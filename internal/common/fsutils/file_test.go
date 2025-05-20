// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/fsutils"
	"github.com/stretchr/testify/require"
)

func TestSafeReadFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutils-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testContent := []byte("test content")
	filePath := filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(filePath, testContent, 0600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectError bool
		expected    []byte
	}{
		{
			name:        "Valid file",
			path:        filePath,
			expectError: false,
			expected:    testContent,
		},
		{
			name:        "Non-existent file",
			path:        filepath.Join(tempDir, "non-existent.txt"),
			expectError: true,
			expected:    nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			content, err := fsutils.SafeReadFile(testCase.path)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expected, content)
			}
		})
	}
}

func TestFindFilesWithExtensions(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutils-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files with different extensions
	files := map[string]string{
		"file1.txt":  "text content",
		"file2.json": "{\"key\": \"value\"}",
		"file3.xml":  "<root></root>",
		"file4.txt":  "more text",
		"file5":      "no extension",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0600)
		require.NoError(t, err)
	}

	// Create a subdirectory with a file (should be ignored)
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0700)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "file6.txt"), []byte("subdirectory file"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name           string
		dir            string
		extensions     []string
		expectedCount  int
		expectedExists map[string]bool
		expectError    bool
	}{
		{
			name:          "Find .txt files",
			dir:           tempDir,
			extensions:    []string{".txt"},
			expectedCount: 2,
			expectedExists: map[string]bool{
				"file1.txt": true,
				"file4.txt": true,
			},
			expectError: false,
		},
		{
			name:          "Find multiple extensions",
			dir:           tempDir,
			extensions:    []string{".json", ".xml"},
			expectedCount: 2,
			expectedExists: map[string]bool{
				"file2.json": true,
				"file3.xml":  true,
			},
			expectError: false,
		},
		{
			name:           "Find non-existent extension",
			dir:            tempDir,
			extensions:     []string{".doc"},
			expectedCount:  0,
			expectedExists: map[string]bool{
				// Empty
			},
			expectError: false,
		},
		{
			name:        "Non-existent directory",
			dir:         filepath.Join(tempDir, "non-existent"),
			extensions:  []string{".txt"},
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			files, err := fsutils.FindFilesWithExtensions(testCase.dir, testCase.extensions)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, files, testCase.expectedCount)

				for _, file := range files {
					baseName := filepath.Base(file)
					_, expected := testCase.expectedExists[baseName]
					require.True(t, expected, "Unexpected file found: %s", baseName)
				}
			}
		})
	}
}
