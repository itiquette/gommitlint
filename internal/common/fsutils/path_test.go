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

func TestSanitizePath(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fsutils-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file for negative test
	filePath := filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(filePath, []byte("test content"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Valid directory path",
			input:       tempDir,
			expectError: false,
		},
		{
			name:        "Non-existent path",
			input:       filepath.Join(tempDir, "non-existent"),
			expectError: true,
		},
		{
			name:        "File instead of directory",
			input:       filePath,
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := fsutils.SanitizePath(testCase.input)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				absPath, _ := filepath.Abs(testCase.input)
				require.Equal(t, absPath, result)
			}
		})
	}
}
