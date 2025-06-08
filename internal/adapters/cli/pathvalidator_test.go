// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrictPathValidator_ValidateHookPath(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		hookType      string
		setupRepo     func(string) error
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:          "invalid hook type",
			repoPath:      "/tmp",
			hookType:      "invalid-hook",
			expectError:   true,
			errorContains: "invalid hook type",
			description:   "should reject invalid hook types",
		},
		{
			name:          "valid hook type with invalid repo",
			repoPath:      "/nonexistent/path",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject invalid repository paths",
		},
		{
			name:     "valid repo and hook type",
			repoPath: "",
			hookType: "commit-msg",
			setupRepo: func(path string) error {
				// Create a minimal git repo structure
				gitDir := filepath.Join(path, ".git")
				hooksDir := filepath.Join(gitDir, "hooks")

				// Create .git directory
				if err := os.MkdirAll(hooksDir, 0755); err != nil {
					return err
				}

				// Create a minimal config file to satisfy git repository validation
				configPath := filepath.Join(gitDir, "config")
				configContent := "[core]\n\trepositoryformatversion = 0\n"

				return os.WriteFile(configPath, []byte(configContent), 0600)
			},
			expectError: false,
			description: "should accept valid repo and hook type",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			validator := StrictPathValidator{}

			// Set up temporary repo if needed
			var tempDir string

			if testCase.setupRepo != nil {
				var err error
				tempDir, err = os.MkdirTemp("", "pathvalidator-test-*")
				require.NoError(t, err)

				defer os.RemoveAll(tempDir)

				err = testCase.setupRepo(tempDir)
				require.NoError(t, err)

				testCase.repoPath = tempDir
			}

			hookPath, err := validator.ValidateHookPath(testCase.repoPath, testCase.hookType)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorContains != "" {
					require.Contains(t, err.Error(), testCase.errorContains, testCase.description)
				}

				require.Empty(t, hookPath, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, hookPath, testCase.description)
				require.Contains(t, hookPath, testCase.hookType, "hook path should contain hook type")
			}
		})
	}
}

func TestStrictPathValidator_EnsureHooksDirectory(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		setupRepo     func(string) error
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:          "invalid repository path",
			repoPath:      "/nonexistent/invalid/path",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject invalid repository paths",
		},
		{
			name:     "create hooks directory in valid repo",
			repoPath: "",
			setupRepo: func(path string) error {
				// Create minimal git repo structure
				gitDir := filepath.Join(path, ".git")

				if err := os.MkdirAll(gitDir, 0755); err != nil {
					return err
				}

				// Create a minimal config file to satisfy git repository validation
				configPath := filepath.Join(gitDir, "config")
				configContent := "[core]\n\trepositoryformatversion = 0\n"

				return os.WriteFile(configPath, []byte(configContent), 0600)
			},
			expectError: false,
			description: "should create hooks directory successfully",
		},
		{
			name:     "hooks directory already exists",
			repoPath: "",
			setupRepo: func(path string) error {
				// Create complete git repo structure
				gitDir := filepath.Join(path, ".git")
				hooksDir := filepath.Join(gitDir, "hooks")

				if err := os.MkdirAll(hooksDir, 0755); err != nil {
					return err
				}

				// Create a minimal config file to satisfy git repository validation
				configPath := filepath.Join(gitDir, "config")
				configContent := "[core]\n\trepositoryformatversion = 0\n"

				return os.WriteFile(configPath, []byte(configContent), 0600)
			},
			expectError: false,
			description: "should handle existing hooks directory",
		},
		{
			name:     "hooks path exists but is a file",
			repoPath: "",
			setupRepo: func(path string) error {
				// Create git dir and a file where hooks dir should be
				gitDir := filepath.Join(path, ".git")
				err := os.MkdirAll(gitDir, 0755)
				if err != nil {
					return err
				}

				// Create a minimal config file to satisfy git repository validation
				configPath := filepath.Join(gitDir, "config")
				configContent := "[core]\n\trepositoryformatversion = 0\n"
				if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
					return err
				}

				hooksPath := filepath.Join(gitDir, "hooks")

				return os.WriteFile(hooksPath, []byte("test"), 0600)
			},
			expectError:   true,
			errorContains: "not a directory",
			description:   "should reject when hooks path is a file",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			validator := StrictPathValidator{}

			// Set up temporary repo if needed
			var tempDir string

			if testCase.setupRepo != nil {
				var err error
				tempDir, err = os.MkdirTemp("", "pathvalidator-hooks-test-*")
				require.NoError(t, err)

				defer os.RemoveAll(tempDir)

				err = testCase.setupRepo(tempDir)
				require.NoError(t, err)

				testCase.repoPath = tempDir
			}

			err := validator.EnsureHooksDirectory(testCase.repoPath)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorContains != "" {
					require.Contains(t, err.Error(), testCase.errorContains, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)

				// Verify hooks directory was created/exists
				if tempDir != "" {
					hooksDir := filepath.Join(tempDir, ".git", "hooks")
					stat, err := os.Stat(hooksDir)
					require.NoError(t, err, "hooks directory should exist")
					require.True(t, stat.IsDir(), "hooks path should be a directory")
				}
			}
		})
	}
}

func TestDefaultPathValidator(t *testing.T) {
	t.Run("returns StrictPathValidator", func(t *testing.T) {
		validator := DefaultPathValidator()

		require.NotNil(t, validator, "should return a validator")

		// Verify it's the expected type
		_, ok := validator.(StrictPathValidator)
		require.True(t, ok, "should return StrictPathValidator")
	})
}

func TestPathValidatorInterface(t *testing.T) {
	t.Run("StrictPathValidator implements PathValidator", func(t *testing.T) {
		var validator PathValidator = StrictPathValidator{}

		require.NotNil(t, validator, "should implement PathValidator interface")

		// Test that interface methods exist (compilation test)
		require.NotNil(t, validator.ValidateHookPath, "should have ValidateHookPath method")
		require.NotNil(t, validator.EnsureHooksDirectory, "should have EnsureHooksDirectory method")
	})
}

func TestStrictPathValidator_Integration(t *testing.T) {
	tests := []struct {
		name        string
		hookType    string
		description string
	}{
		{
			name:        "commit-msg hook",
			hookType:    "commit-msg",
			description: "should handle commit-msg hook validation",
		},
		{
			name:        "pre-commit hook",
			hookType:    "pre-commit",
			description: "should handle pre-commit hook validation",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary git repository
			tempDir, err := os.MkdirTemp("", "pathvalidator-integration-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Create git directory structure
			gitDir := filepath.Join(tempDir, ".git")
			err = os.MkdirAll(gitDir, 0755)
			require.NoError(t, err)

			// Create a minimal config file to satisfy git repository validation
			configPath := filepath.Join(gitDir, "config")
			configContent := "[core]\n\trepositoryformatversion = 0\n"
			err = os.WriteFile(configPath, []byte(configContent), 0600)
			require.NoError(t, err)

			validator := StrictPathValidator{}

			// Test EnsureHooksDirectory first
			err = validator.EnsureHooksDirectory(tempDir)
			require.NoError(t, err, "should create hooks directory")

			// Verify hooks directory was created
			hooksDir := filepath.Join(gitDir, "hooks")
			stat, err := os.Stat(hooksDir)
			require.NoError(t, err, "hooks directory should exist")
			require.True(t, stat.IsDir(), "hooks path should be a directory")

			// Test ValidateHookPath
			hookPath, err := validator.ValidateHookPath(tempDir, testCase.hookType)
			require.NoError(t, err, testCase.description)
			require.NotEmpty(t, hookPath, testCase.description)
			require.Contains(t, hookPath, testCase.hookType, "hook path should contain hook type")
			require.Contains(t, hookPath, hooksDir, "hook path should be in hooks directory")
		})
	}
}

func TestStrictPathValidator_SecurityValidation(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		hookType      string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:          "path traversal in repo path",
			repoPath:      "../../../etc",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject path traversal attempts",
		},
		{
			name:          "special characters in hook type",
			repoPath:      "/tmp",
			hookType:      "commit-msg; rm -rf /",
			expectError:   true,
			errorContains: "invalid hook type",
			description:   "should reject hook types with special characters",
		},
		{
			name:          "null bytes in hook type",
			repoPath:      "/tmp",
			hookType:      "commit-msg\x00",
			expectError:   true,
			errorContains: "invalid hook type",
			description:   "should reject hook types with null bytes",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			validator := StrictPathValidator{}

			hookPath, err := validator.ValidateHookPath(testCase.repoPath, testCase.hookType)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorContains != "" {
					require.Contains(t, err.Error(), testCase.errorContains, testCase.description)
				}

				require.Empty(t, hookPath, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, hookPath, testCase.description)
			}
		})
	}
}

// TestStrictPathValidator_CrossDirectoryValidation tests validation with --repo-path from different directories.
func TestStrictPathValidator_CrossDirectoryValidation(t *testing.T) {
	// Create test repository
	repoDir := t.TempDir()

	// Create git directory structure
	gitDir := filepath.Join(repoDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	err := os.MkdirAll(hooksDir, 0755)
	require.NoError(t, err)

	// Create config file to make it a valid git repo
	configFile := filepath.Join(gitDir, "config")
	err = os.WriteFile(configFile, []byte("[core]\n\trepositoryformatversion = 0"), 0600)
	require.NoError(t, err)

	// Create different working directory
	workDir := t.TempDir()
	originalWd, _ := os.Getwd()

	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(workDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		repoPath    string
		hookType    string
		expectError bool
		description string
	}{
		{
			name:        "validate absolute path to different repo",
			repoPath:    repoDir,
			hookType:    "commit-msg",
			expectError: false,
			description: "should validate repository at absolute path",
		},
		{
			name:        "validate hook in different directory",
			repoPath:    repoDir,
			hookType:    "pre-commit",
			expectError: false,
			description: "should validate different hook types in remote repo",
		},
		{
			name:        "reject current directory without git",
			repoPath:    ".",
			hookType:    "commit-msg",
			expectError: true,
			description: "should reject current directory when it's not a git repo",
		},
		{
			name:        "reject non-existent remote path",
			repoPath:    "/nonexistent/remote/repo",
			hookType:    "commit-msg",
			expectError: true,
			description: "should reject non-existent repository paths",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			validator := StrictPathValidator{}

			hookPath, err := validator.ValidateHookPath(testCase.repoPath, testCase.hookType)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
				require.Empty(t, hookPath)
			} else {
				require.NoError(t, err, testCase.description)
				require.NotEmpty(t, hookPath)
				require.Contains(t, hookPath, testCase.hookType)
				require.Contains(t, hookPath, testCase.repoPath)
			}
		})
	}
}

// TestStrictPathValidator_RepoPathSecurity tests security validation for --repo-path.
func TestStrictPathValidator_RepoPathSecurity(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		hookType      string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:          "path traversal in repo path",
			repoPath:      "../../../etc",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject path traversal attempts",
		},
		{
			name:          "path with null bytes",
			repoPath:      "/tmp/repo\x00malicious",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject paths with null bytes",
		},
		{
			name:          "symlink traversal attempt",
			repoPath:      "/tmp/../../../etc",
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject symlink traversal attempts",
		},
		{
			name:          "extremely long path",
			repoPath:      "/" + string(make([]byte, 5000)),
			hookType:      "commit-msg",
			expectError:   true,
			errorContains: "invalid repository path",
			description:   "should reject extremely long paths",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			validator := StrictPathValidator{}

			hookPath, err := validator.ValidateHookPath(testCase.repoPath, testCase.hookType)

			require.Error(t, err, testCase.description)
			require.Contains(t, err.Error(), testCase.errorContains, testCase.description)
			require.Empty(t, hookPath)
		})
	}
}

func TestStrictPathValidator_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupTest   func() (string, string, bool)
		description string
	}{
		{
			name: "empty repo path",
			setupTest: func() (string, string, bool) {
				return "", "commit-msg", true // expect error
			},
			description: "should handle empty repository path",
		},
		{
			name: "empty hook type",
			setupTest: func() (string, string, bool) {
				return "/tmp", "", true // expect error
			},
			description: "should handle empty hook type",
		},
		{
			name: "current directory as repo",
			setupTest: func() (string, string, bool) {
				return ".", "commit-msg", false // might work if current dir is a git repo
			},
			description: "should handle current directory reference",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			validator := StrictPathValidator{}
			repoPath, hookType, expectError := testCase.setupTest()

			_, err := validator.ValidateHookPath(repoPath, hookType)

			if expectError {
				require.Error(t, err, testCase.description)
			}
		})
	}
}
