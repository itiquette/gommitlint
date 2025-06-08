// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecurityValidator_ValidateRepoPath(t *testing.T) {
	validator := NewSecurityValidator()

	// Create a temporary git repository for testing
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)
	// Create minimal git config to make it a valid git repo
	configFile := filepath.Join(gitDir, "config")
	err = os.WriteFile(configFile, []byte("[core]\n"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		repoPath    string
		expectError bool
		description string
	}{
		{
			name:        "valid git repository",
			repoPath:    tempDir,
			expectError: false,
			description: "Valid git repository should be accepted",
		},
		{
			name:        "path traversal attack",
			repoPath:    "../../../etc",
			expectError: true,
			description: "Path traversal should be blocked",
		},
		{
			name:        "null byte injection",
			repoPath:    "test\x00malicious",
			expectError: true,
			description: "Null byte injection should be blocked",
		},
		{
			name:        "very long path",
			repoPath:    strings.Repeat("a", 2000),
			expectError: true,
			description: "Excessively long paths should be blocked",
		},
		{
			name:        "non-existent path",
			repoPath:    "/nonexistent/path",
			expectError: true,
			description: "Non-existent paths should be blocked",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := validator.ValidateRepoPath(testCase.repoPath)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestSecurityValidator_ValidateOutputFilePath(t *testing.T) {
	validator := NewSecurityValidator()

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "empty path",
			filePath:    "",
			expectError: true,
			description: "Empty path should be rejected",
		},
		{
			name:        "valid relative path",
			filePath:    "output.txt",
			expectError: false,
			description: "Valid relative path should be accepted",
		},
		{
			name:        "path traversal attack",
			filePath:    "../../../etc/passwd",
			expectError: true,
			description: "Path traversal should be blocked",
		},
		{
			name:        "null byte injection",
			filePath:    "output\x00.txt",
			expectError: true,
			description: "Null byte injection should be blocked",
		},
		{
			name:        "control characters",
			filePath:    "output\x01.txt",
			expectError: true,
			description: "Control characters should be blocked",
		},
		{
			name:        "URL encoded traversal",
			filePath:    "test%2e%2e/output.txt",
			expectError: true,
			description: "URL encoded path traversal should be blocked",
		},
		{
			name:        "very long path",
			filePath:    strings.Repeat("a", 1500) + ".txt",
			expectError: true,
			description: "Excessively long paths should be blocked",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := validator.ValidateOutputFilePath(testCase.filePath)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestSecurityValidator_ValidateMessageFilePath(t *testing.T) {
	validator := NewSecurityValidator()

	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-message.txt")
	err := os.WriteFile(testFile, []byte("test message"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "empty path",
			filePath:    "",
			expectError: true,
			description: "Empty path should be rejected",
		},
		{
			name:        "valid existing file",
			filePath:    testFile,
			expectError: false,
			description: "Valid existing file should be accepted",
		},
		{
			name:        "non-existent file",
			filePath:    "non-existent.txt",
			expectError: true,
			description: "Non-existent file should be rejected",
		},
		{
			name:        "path traversal attack",
			filePath:    "../../../etc/passwd",
			expectError: true,
			description: "Path traversal should be blocked",
		},
		{
			name:        "null byte injection",
			filePath:    "message\x00.txt",
			expectError: true,
			description: "Null byte injection should be blocked",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := validator.ValidateMessageFilePath(testCase.filePath)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestSecurityValidator_ValidateGitReference(t *testing.T) {
	validator := NewSecurityValidator()

	tests := []struct {
		name        string
		gitRef      string
		expectError bool
		description string
	}{
		{
			name:        "empty reference",
			gitRef:      "",
			expectError: true,
			description: "Empty reference should be rejected",
		},
		{
			name:        "valid HEAD reference",
			gitRef:      "HEAD",
			expectError: false,
			description: "HEAD reference should be valid",
		},
		{
			name:        "valid branch name",
			gitRef:      "main",
			expectError: false,
			description: "Valid branch name should be accepted",
		},
		{
			name:        "valid commit hash",
			gitRef:      "abc123def456",
			expectError: false,
			description: "Valid commit hash should be accepted",
		},
		{
			name:        "command injection attempt",
			gitRef:      "main; rm -rf /",
			expectError: true,
			description: "Command injection should be blocked",
		},
		{
			name:        "null byte injection",
			gitRef:      "main\x00malicious",
			expectError: true,
			description: "Null byte injection should be blocked",
		},
		{
			name:        "path traversal in ref",
			gitRef:      "../../../etc/passwd",
			expectError: true,
			description: "Path traversal should be blocked",
		},
		{
			name:        "control characters",
			gitRef:      "main\x01branch",
			expectError: true,
			description: "Control characters should be blocked",
		},
		{
			name:        "backtick injection",
			gitRef:      "main`rm -rf /`",
			expectError: true,
			description: "Backtick injection should be blocked",
		},
		{
			name:        "dollar injection",
			gitRef:      "main$(rm -rf /)",
			expectError: true,
			description: "Dollar injection should be blocked",
		},
		{
			name:        "very long reference",
			gitRef:      strings.Repeat("a", 300),
			expectError: true,
			description: "Excessively long references should be blocked",
		},
		{
			name:        "invalid git characters",
			gitRef:      "branch~name^with:invalid?chars*",
			expectError: true,
			description: "Invalid git characters should be blocked",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateGitReference(testCase.gitRef)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestSecurityValidator_ValidateCommitRange(t *testing.T) {
	validator := NewSecurityValidator()

	tests := []struct {
		name        string
		commitRange string
		expectError bool
		description string
	}{
		{
			name:        "empty range",
			commitRange: "",
			expectError: true,
			description: "Empty range should be rejected",
		},
		{
			name:        "valid double dot range",
			commitRange: "main..feature",
			expectError: false,
			description: "Valid double dot range should be accepted",
		},
		{
			name:        "valid triple dot range",
			commitRange: "main...feature",
			expectError: false,
			description: "Valid triple dot range should be accepted",
		},
		{
			name:        "invalid format no dots",
			commitRange: "main-feature",
			expectError: true,
			description: "Range without dots should be rejected",
		},
		{
			name:        "invalid three parts",
			commitRange: "main..feature..other",
			expectError: true,
			description: "Range with three parts should be rejected",
		},
		{
			name:        "command injection in range",
			commitRange: "main..feature; rm -rf /",
			expectError: true,
			description: "Command injection in range should be blocked",
		},
		{
			name:        "null byte in range",
			commitRange: "main..\x00feature",
			expectError: true,
			description: "Null byte in range should be blocked",
		},
		{
			name:        "path traversal in range",
			commitRange: "../../../etc/passwd..main",
			expectError: true,
			description: "Path traversal in range should be blocked",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateCommitRange(testCase.commitRange)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestSecurityValidator_validatePathSecurity(t *testing.T) {
	validator := NewSecurityValidator()

	tests := []struct {
		name        string
		path        string
		expectError bool
		description string
	}{
		{
			name:        "valid path",
			path:        "valid/path.txt",
			expectError: false,
			description: "Valid path should be accepted",
		},
		{
			name:        "null byte injection",
			path:        "path\x00malicious",
			expectError: true,
			description: "Null byte should be rejected",
		},
		{
			name:        "control character",
			path:        "path\x01file",
			expectError: true,
			description: "Control character should be rejected",
		},
		{
			name:        "tab character allowed",
			path:        "path\tfile",
			expectError: false,
			description: "Tab character should be allowed",
		},
		{
			name:        "path traversal",
			path:        "../secret/file",
			expectError: true,
			description: "Path traversal should be rejected",
		},
		{
			name:        "URL encoded traversal",
			path:        "path%2e%2esecret",
			expectError: true,
			description: "URL encoded traversal should be rejected",
		},
		{
			name:        "very long path",
			path:        strings.Repeat("a", 1001),
			expectError: true,
			description: "Excessively long path should be rejected",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.validatePathSecurity(testCase.path)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

// Benchmark tests to ensure security checks don't significantly impact performance.
func BenchmarkSecurityValidator_ValidateGitReference(b *testing.B) {
	validator := NewSecurityValidator()
	validRef := "main"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.ValidateGitReference(validRef)
	}
}

func BenchmarkSecurityValidator_ValidatePathSecurity(b *testing.B) {
	validator := NewSecurityValidator()
	validPath := "path/to/file.txt"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.validatePathSecurity(validPath)
	}
}
