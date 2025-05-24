// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package fsutils_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/common/fsutils"
	"github.com/stretchr/testify/require"
)

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

func TestSafeWriteFile(t *testing.T) {
	tests := []struct {
		name        string
		permissions os.FileMode
		content     []byte
		setupDir    func(dir string) (string, error)
		expectError bool
	}{
		{
			name:        "Write to new file",
			permissions: 0600,
			content:     []byte("test content"),
			setupDir: func(dir string) (string, error) {
				return filepath.Join(dir, "new.txt"), nil
			},
			expectError: false,
		},
		{
			name:        "Write to existing file",
			permissions: 0600,
			content:     []byte("new content"),
			setupDir: func(dir string) (string, error) {
				filePath := filepath.Join(dir, "existing.txt")
				err := os.WriteFile(filePath, []byte("old content"), 0600)

				return filePath, err
			},
			expectError: false,
		},
		{
			name:        "Write with executable permissions",
			permissions: 0700,
			content:     []byte("executable content"),
			setupDir: func(dir string) (string, error) {
				return filepath.Join(dir, "executable.txt"), nil
			},
			expectError: false,
		},
		{
			name:        "Write to nested directory",
			permissions: 0600,
			content:     []byte("nested content"),
			setupDir: func(dir string) (string, error) {
				return filepath.Join(dir, "nested", "dir", "nested.txt"), nil
			},
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a temporary directory for each test
			tempDir := t.TempDir()

			// Setup the test directory and get the target file path
			testFilePath, err := testCase.setupDir(tempDir)
			require.NoError(t, err, "Failed to set up test directory")

			// Write the file using SafeWriteFile
			err = fsutils.SafeWriteFile(testFilePath, testCase.content, testCase.permissions)
			if testCase.expectError {
				require.Error(t, err, "Expected an error but got none")

				return
			}

			require.NoError(t, err, "SafeWriteFile should not return an error")

			// Verify file content
			content, err := os.ReadFile(testFilePath)
			require.NoError(t, err, "Failed to read test file")
			require.Equal(t, testCase.content, content, "File content should match what was written")

			// Verify file permissions
			info, err := os.Stat(testFilePath)
			require.NoError(t, err, "Failed to stat test file")
			require.Equal(t, testCase.permissions, info.Mode().Perm()&0777, "File permissions should match what was set")
		})
	}
}

// TestSafeWriteFileRace tests that the file is written atomically
// by simulating a potential race condition with file access.
func TestSafeWriteFileRace(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "race.txt")

	// First, create the file with initial content
	initialContent := []byte("initial content")
	err := os.WriteFile(testFilePath, initialContent, 0600)
	require.NoError(t, err, "Failed to create initial test file")

	// Create channels for communication with reader goroutine
	done := make(chan struct{})
	errChan := make(chan string, 10) // Buffer to avoid blocking

	// Start a goroutine that reads the file repeatedly
	go func() {
		defer close(done)

		// Read the file in a loop for a short period
		// This simulates concurrent access that might cause race conditions
		timeout := time.After(500 * time.Millisecond)

		for {
			select {
			case <-timeout:
				return
			default:
				content, err := os.ReadFile(testFilePath)
				if err == nil {
					// File should either contain the initial content or the new content
					// but never be in an inconsistent state
					validContent := string(content) == "initial content" ||
						string(content) == "new content"
					if !validContent {
						// Instead of using require in goroutine, send error to channel
						errChan <- "File contained invalid content: " + string(content)
					}
				}

				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Write new content with SafeWriteFile while the reader goroutine is running
	newContent := []byte("new content")
	err = fsutils.SafeWriteFile(testFilePath, newContent, 0600)
	require.NoError(t, err, "SafeWriteFile should not return an error")

	// Wait for reader goroutine to finish
	<-done
	close(errChan)

	// Check if any errors were reported by reader goroutine
	for errMsg := range errChan {
		require.Fail(t, errMsg)
	}

	// Verify the file eventually contains the new content
	content, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read test file")
	require.Equal(t, newContent, content, "File should contain new content")
}
