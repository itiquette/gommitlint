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

func TestInstallHook(t *testing.T) {
	// Create a temporary directory using Go's testing utility - automatically cleaned up
	tmpDir := t.TempDir()

	// Set up a real Git repository
	err := setupGitRepo(tmpDir)
	require.NoError(t, err, "Failed to set up Git repository")

	// Define table-driven test cases
	tests := []struct {
		name          string
		repoPath      string
		force         bool
		setupFunc     func(repoDir string) error
		expectedError bool
		checkFunc     func(t *testing.T, repoDir string)
	}{
		{
			name:     "successful_commit_msg_hook",
			repoPath: tmpDir,
			force:    false,
			setupFunc: func(_ string) error {
				return nil // No special setup needed
			},
			expectedError: false,
			checkFunc: func(t *testing.T, repoDir string) {
				t.Helper()
				hookPath := filepath.Join(repoDir, ".git", "hooks", "commit-msg")
				content, err := os.ReadFile(hookPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "gommitlint commit-msg hook")

				// Check executable permissions (0755)
				info, err := os.Stat(hookPath)
				require.NoError(t, err)
				require.Equal(t, os.FileMode(0600), info.Mode()&0777)
			},
		},
		{
			name:     "fails_with_existing_hook_no_force",
			repoPath: tmpDir,
			force:    false,
			setupFunc: func(repoDir string) error {
				t.Helper()
				// Create a pre-existing hook
				hooksDir := filepath.Join(repoDir, ".git", "hooks")
				err := os.MkdirAll(hooksDir, 0755)
				if err != nil {
					return err
				}

				return os.WriteFile(
					filepath.Join(hooksDir, "commit-msg"),
					[]byte("#!/bin/sh\necho 'existing hook'\n"),
					0600,
				)
			},
			expectedError: true,
			checkFunc: func(t *testing.T, repoDir string) {
				t.Helper()
				// The existing hook should remain unchanged
				hookPath := filepath.Join(repoDir, ".git", "hooks", "commit-msg")
				content, err := os.ReadFile(hookPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "existing hook")
			},
		},
		{
			name:     "overwrites_existing_hook_with_force",
			repoPath: tmpDir,
			force:    true,
			setupFunc: func(repoDir string) error {
				t.Helper()
				// Create a pre-existing hook
				hooksDir := filepath.Join(repoDir, ".git", "hooks")
				err := os.MkdirAll(hooksDir, 0755)
				if err != nil {
					return err
				}

				return os.WriteFile(
					filepath.Join(hooksDir, "commit-msg"),
					[]byte("#!/bin/sh\necho 'existing hook'\n"),
					0600,
				)
			},
			expectedError: false,
			checkFunc: func(t *testing.T, repoDir string) {
				t.Helper()
				// The hook should be overwritten with our version
				hookPath := filepath.Join(repoDir, ".git", "hooks", "commit-msg")
				content, err := os.ReadFile(hookPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "gommitlint commit-msg hook")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Set up the test case
			err := testCase.setupFunc(tmpDir)
			require.NoError(t, err)

			// Execute the function being tested
			err = installHook(testCase.force, testCase.repoPath)

			// Check the result
			if testCase.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Additional verification
			testCase.checkFunc(t, tmpDir)
		})
	}
}
