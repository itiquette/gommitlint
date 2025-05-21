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

	// Create a nested directory
	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.Mkdir(nestedDir, 0755))

	// Create a file for negative test
	filePath := filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(filePath, []byte("test content"), 0600)
	require.NoError(t, err)

	// Create a symbolic link to the nested directory
	symlinkPath := filepath.Join(tempDir, "symlink")
	require.NoError(t, os.Symlink(nestedDir, symlinkPath))

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
			name:        "Nested directory",
			input:       nestedDir,
			expectError: false,
		},
		{
			name:        "Valid directory with trailing slash",
			input:       tempDir + "/",
			expectError: false,
		},
		{
			name:        "Valid directory with .. components",
			input:       filepath.Join(nestedDir, ".."),
			expectError: false,
		},
		{
			name:        "Symlink to directory",
			input:       symlinkPath,
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
				require.NotEmpty(t, result)

				// Verify the path has been canonicalized
				if testCase.input == filepath.Join(nestedDir, "..") {
					canonPath, err := filepath.EvalSymlinks(tempDir)
					require.NoError(t, err)
					require.Equal(t, canonPath, result)
				} else if testCase.input == symlinkPath {
					canonPath, err := filepath.EvalSymlinks(symlinkPath)
					require.NoError(t, err)
					require.Equal(t, canonPath, result)
				}
			}
		})
	}
}

func TestIsWithinDirectory(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "secpath-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create nested directories
	nestedDir := filepath.Join(tempDir, "nested")
	deepNestedDir := filepath.Join(nestedDir, "deep")
	require.NoError(t, os.MkdirAll(deepNestedDir, 0755))

	// Create a directory outside the temp structure
	outsideDir, err := os.MkdirTemp("", "secpath-outside")
	require.NoError(t, err)
	defer os.RemoveAll(outsideDir)

	// Create a symbolic link to the outside directory
	symlinkPath := filepath.Join(tempDir, "symlink-outside")
	require.NoError(t, os.Symlink(outsideDir, symlinkPath))

	tests := []struct {
		name    string
		path    string
		baseDir string
		want    bool
		wantErr bool
	}{
		{
			name:    "Path equals base directory",
			path:    tempDir,
			baseDir: tempDir,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Path within base directory",
			path:    nestedDir,
			baseDir: tempDir,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Deep nested path within base directory",
			path:    deepNestedDir,
			baseDir: tempDir,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Base directory within path (reversed)",
			path:    tempDir,
			baseDir: nestedDir,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Path outside base directory",
			path:    outsideDir,
			baseDir: tempDir,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Symlink to directory outside base",
			path:    symlinkPath,
			baseDir: tempDir,
			want:    false, // Our implementation resolves symlinks, so it should detect it points outside
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := fsutils.IsWithinDirectory(testCase.path, testCase.baseDir)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "secpath-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create directories for the test cases
	subDir := filepath.Join(tempDir, "foo")
	require.NoError(t, os.Mkdir(subDir, 0755))
	deepDir := filepath.Join(subDir, "bar")
	require.NoError(t, os.Mkdir(deepDir, 0755))

	tests := []struct {
		name     string
		baseDir  string
		elements []string
		want     string
		wantErr  bool
	}{
		{
			name:     "Join simple elements",
			baseDir:  tempDir,
			elements: []string{"foo", "bar"},
			want:     deepDir,
			wantErr:  false,
		},
		{
			name:     "Join with empty element",
			baseDir:  tempDir,
			elements: []string{"foo", "", "bar"},
			want:     deepDir,
			wantErr:  false,
		},
		{
			name:     "Join with dot element",
			baseDir:  tempDir,
			elements: []string{"foo", ".", "bar"},
			want:     deepDir,
			wantErr:  false,
		},
		{
			name:     "Attempt path traversal with ..",
			baseDir:  tempDir,
			elements: []string{"foo", "..", ".."},
			wantErr:  true,
		},
		{
			name:     "Attempt path traversal with leading ..",
			baseDir:  tempDir,
			elements: []string{"../etc/passwd"},
			wantErr:  true,
		},
		{
			name:     "Attempt path traversal with embedded ..",
			baseDir:  tempDir,
			elements: []string{"foo/../../etc/passwd"},
			wantErr:  true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := fsutils.SafeJoin(testCase.baseDir, testCase.elements...)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.want, got)

				// Verify the result is actually within the base directory
				isWithin, err := fsutils.IsWithinDirectory(got, testCase.baseDir)
				require.NoError(t, err)
				require.True(t, isWithin)
			}
		})
	}
}

func TestSafeReadDirectoryContents(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "secpath-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create some files in the directory
	for _, name := range []string{"file1.txt", "file2.txt", "file3.txt"} {
		filePath := filepath.Join(tempDir, name)
		require.NoError(t, os.WriteFile(filePath, []byte("test"), 0600))
	}

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	tests := []struct {
		name    string
		dirPath string
		wantLen int
		wantErr bool
	}{
		{
			name:    "Valid directory",
			dirPath: tempDir,
			wantLen: 4, // 3 files + 1 directory
			wantErr: false,
		},
		{
			name:    "Empty directory",
			dirPath: subDir,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "Non-existent directory",
			dirPath: filepath.Join(tempDir, "nonexistent"),
			wantErr: true,
		},
		{
			name:    "Not a directory",
			dirPath: filepath.Join(tempDir, "file1.txt"),
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			entries, err := fsutils.SafeReadDirectoryContents(testCase.dirPath)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, entries, testCase.wantLen)
			}
		})
	}
}
