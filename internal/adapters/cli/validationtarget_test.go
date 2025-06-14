// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewValidationTarget(t *testing.T) {
	tests := []struct {
		name         string
		messageFile  string
		gitRef       string
		commitRange  string
		baseBranch   string
		commitCount  int
		expectedType string
		expectedSrc  string
		expectedTgt  string
		expectError  bool
		description  string
	}{
		{
			name:         "message file - highest priority",
			messageFile:  "/path/to/message.txt",
			gitRef:       "HEAD",
			commitRange:  "main..feature",
			baseBranch:   "main",
			commitCount:  5,
			expectedType: "message",
			expectedSrc:  "/path/to/message.txt",
			expectedTgt:  "",
			expectError:  false,
			description:  "should prioritize message file over all other options",
		},
		{
			name:         "base branch - second priority",
			messageFile:  "",
			gitRef:       "HEAD",
			commitRange:  "main..feature",
			baseBranch:   "main",
			commitCount:  5,
			expectedType: "range",
			expectedSrc:  "main",
			expectedTgt:  "HEAD",
			expectError:  false,
			description:  "should use base branch when no message file",
		},
		{
			name:         "commit range - third priority",
			messageFile:  "",
			gitRef:       "HEAD",
			commitRange:  "main..feature",
			baseBranch:   "",
			commitCount:  5,
			expectedType: "range",
			expectedSrc:  "main",
			expectedTgt:  "feature",
			expectError:  false,
			description:  "should use commit range when no message file or base branch",
		},
		{
			name:         "git reference - fourth priority",
			messageFile:  "",
			gitRef:       "HEAD~1",
			commitRange:  "",
			baseBranch:   "",
			commitCount:  5,
			expectedType: "commit",
			expectedSrc:  "HEAD~1",
			expectedTgt:  "",
			expectError:  false,
			description:  "should use git reference when no higher priority options",
		},
		{
			name:         "commit count - fifth priority",
			messageFile:  "",
			gitRef:       "",
			commitRange:  "",
			baseBranch:   "",
			commitCount:  5,
			expectedType: "count",
			expectedSrc:  "5",
			expectedTgt:  "",
			expectError:  false,
			description:  "should use commit count when no other options",
		},
		{
			name:         "default to HEAD",
			messageFile:  "",
			gitRef:       "",
			commitRange:  "",
			baseBranch:   "",
			commitCount:  1,
			expectedType: "commit",
			expectedSrc:  "HEAD",
			expectedTgt:  "",
			expectError:  false,
			description:  "should default to HEAD when count is 1",
		},
		{
			name:         "default with no options",
			messageFile:  "",
			gitRef:       "",
			commitRange:  "",
			baseBranch:   "",
			commitCount:  0,
			expectedType: "commit",
			expectedSrc:  "HEAD",
			expectedTgt:  "",
			expectError:  false,
			description:  "should default to HEAD when no options provided",
		},
		{
			name:        "invalid file path - too long",
			messageFile: strings.Repeat("a", 5000), // Exceeds MaxPathLength
			expectError: true,
			description: "should reject paths that are too long",
		},
		{
			name:        "invalid file path - null bytes",
			messageFile: "/path/with\x00null",
			expectError: true,
			description: "should reject paths with null bytes",
		},
		{
			name:        "invalid file path - path traversal",
			messageFile: "/path/../../../etc/passwd",
			expectError: true,
			description: "should reject paths with traversal attempts",
		},
		{
			name:        "invalid git reference - too long",
			gitRef:      strings.Repeat("a", 300), // Exceeds MaxRefLength
			expectError: true,
			description: "should reject git references that are too long",
		},
		{
			name:        "invalid git reference - dangerous chars",
			gitRef:      "HEAD; rm -rf /",
			expectError: true,
			description: "should reject git references with dangerous characters",
		},
		{
			name:        "invalid commit count - negative",
			commitCount: -1,
			expectError: true,
			description: "should reject negative commit counts",
		},
		{
			name:        "invalid commit count - too large",
			commitCount: 2000, // Exceeds MaxCommitCount
			expectError: true,
			description: "should reject commit counts that are too large",
		},
		{
			name:        "invalid commit range format",
			commitRange: "invalid-range-format",
			expectError: true,
			description: "should reject invalid range formats",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			target, err := NewValidationTarget(
				testCase.messageFile,
				testCase.gitRef,
				testCase.commitRange,
				testCase.baseBranch,
				testCase.commitCount,
			)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.Equal(t, testCase.expectedType, target.Type, testCase.description)
				require.Equal(t, testCase.expectedSrc, target.Source, testCase.description)
				require.Equal(t, testCase.expectedTgt, target.Target, testCase.description)
			}
		})
	}
}

func TestValidationTarget_TypeCheckers(t *testing.T) {
	tests := []struct {
		name              string
		target            ValidationTarget
		expectedIsMessage bool
		expectedIsCommit  bool
		expectedIsRange   bool
		expectedIsCount   bool
		description       string
	}{
		{
			name: "message target",
			target: ValidationTarget{
				Type:   "message",
				Source: "/path/to/file",
			},
			expectedIsMessage: true,
			expectedIsCommit:  false,
			expectedIsRange:   false,
			expectedIsCount:   false,
			description:       "should correctly identify message target",
		},
		{
			name: "commit target",
			target: ValidationTarget{
				Type:   "commit",
				Source: "HEAD",
			},
			expectedIsMessage: false,
			expectedIsCommit:  true,
			expectedIsRange:   false,
			expectedIsCount:   false,
			description:       "should correctly identify commit target",
		},
		{
			name: "range target",
			target: ValidationTarget{
				Type:   "range",
				Source: "main",
				Target: "feature",
			},
			expectedIsMessage: false,
			expectedIsCommit:  false,
			expectedIsRange:   true,
			expectedIsCount:   false,
			description:       "should correctly identify range target",
		},
		{
			name: "count target",
			target: ValidationTarget{
				Type:   "count",
				Source: "5",
			},
			expectedIsMessage: false,
			expectedIsCommit:  false,
			expectedIsRange:   false,
			expectedIsCount:   true,
			description:       "should correctly identify count target",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expectedIsMessage, testCase.target.IsMessageFile(), testCase.description)
			require.Equal(t, testCase.expectedIsCommit, testCase.target.IsCommit(), testCase.description)
			require.Equal(t, testCase.expectedIsRange, testCase.target.IsRange(), testCase.description)
			require.Equal(t, testCase.expectedIsCount, testCase.target.IsCount(), testCase.description)
		})
	}
}

func TestValidateInputs(t *testing.T) {
	tests := []struct {
		name        string
		messageFile string
		gitRef      string
		commitRange string
		baseBranch  string
		commitCount int
		expectError bool
		description string
	}{
		{
			name:        "all valid inputs",
			messageFile: "/valid/path.txt",
			gitRef:      "HEAD",
			commitRange: "main..feature",
			baseBranch:  "main",
			commitCount: 5,
			expectError: false,
			description: "should accept all valid inputs",
		},
		{
			name:        "empty inputs are valid",
			messageFile: "",
			gitRef:      "",
			commitRange: "",
			baseBranch:  "",
			commitCount: 0,
			expectError: false,
			description: "should accept empty inputs",
		},
		{
			name:        "stdin message file",
			messageFile: "-",
			expectError: false,
			description: "should accept stdin as message file",
		},
		{
			name:        "invalid message file path",
			messageFile: "/path/with\x00null",
			expectError: true,
			description: "should reject invalid file paths",
		},
		{
			name:        "invalid git reference",
			gitRef:      "..invalid",
			expectError: true,
			description: "should reject invalid git references",
		},
		{
			name:        "invalid base branch",
			baseBranch:  "..invalid",
			expectError: true,
			description: "should reject invalid base branches",
		},
		{
			name:        "invalid commit count",
			commitCount: -1,
			expectError: true,
			description: "should reject invalid commit counts",
		},
		{
			name:        "invalid commit range - bad start",
			commitRange: "..invalid..HEAD",
			expectError: true,
			description: "should reject ranges with invalid start ref",
		},
		{
			name:        "invalid commit range - bad end",
			commitRange: "main....invalid",
			expectError: true,
			description: "should reject ranges with invalid end ref",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateInputs(
				testCase.messageFile,
				testCase.gitRef,
				testCase.commitRange,
				testCase.baseBranch,
				testCase.commitCount,
			)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		description string
	}{
		{
			name:        "empty path is valid",
			path:        "",
			expectError: false,
			description: "should accept empty paths",
		},
		{
			name:        "stdin path is valid",
			path:        "-",
			expectError: false,
			description: "should accept stdin path",
		},
		{
			name:        "normal path is valid",
			path:        "/path/to/file.txt",
			expectError: false,
			description: "should accept normal file paths",
		},
		{
			name:        "relative path is valid",
			path:        "./relative/path.txt",
			expectError: false,
			description: "should accept relative paths",
		},
		{
			name:        "path too long",
			path:        strings.Repeat("a", 5000),
			expectError: true,
			description: "should reject paths that are too long",
		},
		{
			name:        "path with null bytes",
			path:        "/path/with\x00null",
			expectError: true,
			description: "should reject paths with null bytes",
		},
		{
			name:        "path traversal attempt",
			path:        "/path/../../../etc/passwd",
			expectError: true,
			description: "should reject path traversal attempts",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateFilePath(testCase.path)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestValidateGitReference(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		expectError bool
		description string
	}{
		{
			name:        "empty reference is valid",
			ref:         "",
			expectError: false,
			description: "should accept empty references",
		},
		{
			name:        "HEAD reference",
			ref:         "HEAD",
			expectError: false,
			description: "should accept HEAD reference",
		},
		{
			name:        "HEAD with tilde",
			ref:         "HEAD~1",
			expectError: false,
			description: "should accept HEAD with tilde notation",
		},
		{
			name:        "branch name",
			ref:         "main",
			expectError: false,
			description: "should accept branch names",
		},
		{
			name:        "feature branch",
			ref:         "feature/new-thing",
			expectError: false,
			description: "should accept feature branch names",
		},
		{
			name:        "commit hash",
			ref:         "abc123def456",
			expectError: false,
			description: "should accept commit hashes",
		},
		{
			name:        "reference too long",
			ref:         strings.Repeat("a", 300),
			expectError: true,
			description: "should reject references that are too long",
		},
		{
			name:        "reference with null bytes",
			ref:         "HEAD\x00test",
			expectError: true,
			description: "should reject references with null bytes",
		},
		{
			name:        "reference starting with dot",
			ref:         ".invalid",
			expectError: true,
			description: "should reject references starting with dot",
		},
		{
			name:        "reference with double dots",
			ref:         "branch..invalid",
			expectError: true,
			description: "should reject references with double dots",
		},
		{
			name:        "reference with shell metacharacters",
			ref:         "HEAD; rm -rf /",
			expectError: true,
			description: "should reject references with shell metacharacters",
		},
		{
			name:        "reference with pipe",
			ref:         "HEAD | cat",
			expectError: true,
			description: "should reject references with pipe characters",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateGitReference(testCase.ref)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestValidateCommitCount(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		expectError bool
		description string
	}{
		{
			name:        "zero count is valid",
			count:       0,
			expectError: false,
			description: "should accept zero count",
		},
		{
			name:        "positive count is valid",
			count:       5,
			expectError: false,
			description: "should accept positive counts",
		},
		{
			name:        "large valid count",
			count:       100,
			expectError: false,
			description: "should accept large valid counts",
		},
		{
			name:        "maximum allowed count",
			count:       1000,
			expectError: false,
			description: "should accept maximum allowed count",
		},
		{
			name:        "negative count",
			count:       -1,
			expectError: true,
			description: "should reject negative counts",
		},
		{
			name:        "count exceeds maximum",
			count:       2000,
			expectError: true,
			description: "should reject counts that exceed maximum",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateCommitCount(testCase.count)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestParseRevisionRange(t *testing.T) {
	tests := []struct {
		name          string
		revRange      string
		expectedParts []string
		description   string
	}{
		{
			name:          "standard range with double dots",
			revRange:      "main..feature",
			expectedParts: []string{"main", "feature"},
			description:   "should parse standard range format",
		},
		{
			name:          "symmetric difference with triple dots",
			revRange:      "main...feature",
			expectedParts: []string{"main", "feature"},
			description:   "should parse symmetric difference format",
		},
		{
			name:          "single reference",
			revRange:      "HEAD",
			expectedParts: []string{"HEAD"},
			description:   "should return single reference as-is",
		},
		{
			name:          "range with spaces",
			revRange:      "  main  ..  feature  ",
			expectedParts: []string{"main", "feature"},
			description:   "should trim whitespace from parts",
		},
		{
			name:          "empty range",
			revRange:      "",
			expectedParts: []string{""},
			description:   "should handle empty range",
		},
		{
			name:          "range with commit hashes",
			revRange:      "abc123..def456",
			expectedParts: []string{"abc123", "def456"},
			description:   "should parse commit hash ranges",
		},
		{
			name:          "complex reference names",
			revRange:      "origin/main..feature/new-thing",
			expectedParts: []string{"origin/main", "feature/new-thing"},
			description:   "should parse complex reference names",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			parts := parseRevisionRange(testCase.revRange)

			require.Equal(t, testCase.expectedParts, parts, testCase.description)
		})
	}
}

func TestValidateParameterLength(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		value       string
		maxLength   int
		expectError bool
		description string
	}{
		{
			name:        "valid length",
			paramName:   "test",
			value:       "short",
			maxLength:   10,
			expectError: false,
			description: "should accept values within length limit",
		},
		{
			name:        "exactly at limit",
			paramName:   "test",
			value:       "exactly10c",
			maxLength:   10,
			expectError: false,
			description: "should accept values exactly at limit",
		},
		{
			name:        "exceeds limit",
			paramName:   "test",
			value:       "this is too long",
			maxLength:   10,
			expectError: true,
			description: "should reject values that exceed limit",
		},
		{
			name:        "empty value",
			paramName:   "test",
			value:       "",
			maxLength:   10,
			expectError: false,
			description: "should accept empty values",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateParameterLength(testCase.paramName, testCase.value, testCase.maxLength)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
				require.Contains(t, err.Error(), testCase.paramName, "error should mention parameter name")
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestValidationTarget_Constants(t *testing.T) {
	t.Run("constants have reasonable values", func(t *testing.T) {
		require.Equal(t, 4096, MaxPathLength, "MaxPathLength should be 4096")
		require.Equal(t, 255, MaxRefLength, "MaxRefLength should be 255")
		require.Equal(t, 1000, MaxCommitCount, "MaxCommitCount should be 1000")
	})
}

func TestValidationTarget_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupTarget func() (string, string, string, string, int)
		expectError bool
		description string
	}{
		{
			name: "all parameters at maximum valid length",
			setupTarget: func() (string, string, string, string, int) {
				return strings.Repeat("a", MaxPathLength),
					strings.Repeat("b", MaxRefLength),
					strings.Repeat("c", 50) + ".." + strings.Repeat("d", 50),
					strings.Repeat("e", MaxRefLength),
					MaxCommitCount
			},
			expectError: false,
			description: "should handle maximum valid lengths",
		},
		{
			name: "unicode characters in paths and refs",
			setupTarget: func() (string, string, string, string, int) {
				return "/path/with/ünïcödé.txt",
					"brañch-ñämé",
					"ünicöde..braɲch",
					"basé-braɲch",
					5
			},
			expectError: false,
			description: "should handle unicode characters",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			messageFile, gitRef, commitRange, baseBranch, commitCount := testCase.setupTarget()

			target, err := NewValidationTarget(messageFile, gitRef, commitRange, baseBranch, commitCount)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, target.Type, "should have a valid type")
			}
		})
	}
}
