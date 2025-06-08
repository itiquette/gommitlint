// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"path/filepath"
	"testing"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/stretchr/testify/require"
)

func TestCreateDefaultHookScript(t *testing.T) {
	script := createDefaultHookScript()

	// Verify script structure
	require.Contains(t, script, "#!/bin/sh", "script should have shebang")
	require.Contains(t, script, "gommitlint commit-msg hook", "script should have description")
	require.Contains(t, script, "git commit --no-verify", "script should mention bypass option")

	// Verify environment variables are documented
	require.Contains(t, script, "GOMMITLINT_DEBUG=1", "script should document debug env var")
	require.Contains(t, script, "NO_COLOR=1", "script should document no-color env var")

	// Verify error handling
	require.Contains(t, script, "set -e", "script should exit on error")
	require.Contains(t, script, "command -v gommitlint", "script should check for gommitlint")
	require.Contains(t, script, "gommitlint not found in PATH", "script should have helpful error message")

	// Verify command building
	require.Contains(t, script, `FLAGS=""`, "script should initialize flags")
	require.Contains(t, script, "--color=never", "script should handle no-color")
	require.Contains(t, script, "--debug", "script should handle debug mode")

	// Verify validation command
	require.Contains(t, script, "gommitlint validate --message-file=", "script should run validation")
	require.Contains(t, script, "$FLAGS", "script should use built flags")

	// Verify error messages
	require.Contains(t, script, "Commit rejected due to message validation errors", "script should have rejection message")
	require.Contains(t, script, "gommitlint config show", "script should suggest showing config")
}

func TestGenerateCommitMsgHook(t *testing.T) {
	hook := generateCommitMsgHook()

	// Should return the default script
	require.Equal(t, createDefaultHookScript(), hook)
}

func TestHookInstallationParameters(t *testing.T) {
	tests := []struct {
		name      string
		force     bool
		repoPath  string
		wantForce bool
		wantType  string
	}{
		{
			name:      "default parameters",
			force:     false,
			repoPath:  "/test/repo",
			wantForce: false,
			wantType:  "commit-msg",
		},
		{
			name:      "force enabled",
			force:     true,
			repoPath:  "/test/repo",
			wantForce: true,
			wantType:  "commit-msg",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			params := NewHookInstallationParameters(testCase.force, testCase.repoPath)

			require.Equal(t, testCase.wantForce, params.Force)
			require.Equal(t, testCase.repoPath, params.RepoPath)
			require.Equal(t, testCase.wantType, params.HookType)
			require.NotNil(t, params.PathValidator)
		})
	}
}

func TestHookInstallationParameters_GetHookContent(t *testing.T) {
	params := NewHookInstallationParameters(false, "/test/repo")
	content := params.GetHookContent()

	// Should return the default hook script
	require.Equal(t, createDefaultHookScript(), content)
	require.Contains(t, content, "#!/bin/sh")
	require.Contains(t, content, "gommitlint validate")
}

func TestHookInstallationParameters_CanInstallHook(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0755))

	tests := []struct {
		name         string
		force        bool
		existingHook bool
		wantErr      bool
		errContains  string
	}{
		{
			name:         "no existing hook",
			force:        false,
			existingHook: false,
			wantErr:      false,
		},
		{
			name:         "existing hook without force",
			force:        false,
			existingHook: true,
			wantErr:      true,
			errContains:  "use --force to overwrite",
		},
		{
			name:         "existing hook with force",
			force:        true,
			existingHook: true,
			wantErr:      false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			hookPath := filepath.Join(hooksDir, "commit-msg")

			// Setup existing hook if needed
			if testCase.existingHook {
				require.NoError(t, os.WriteFile(hookPath, []byte("existing hook"), 0600))
			} else {
				os.Remove(hookPath) // Ensure it doesn't exist
			}

			// Create params with mock path validator
			params := HookInstallationParameters{
				Force:    testCase.force,
				RepoPath: tmpDir,
				HookType: "commit-msg",
				PathValidator: &mockPathValidator{
					hookPath: hookPath,
				},
			}

			err := params.CanInstallHook()
			if testCase.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// mockPathValidator is a test implementation of PathValidator.
type mockPathValidator struct {
	hookPath    string
	hooksDir    string
	shouldError bool
}

func (m *mockPathValidator) ValidateHookPath(_, _ string) (string, error) {
	if m.shouldError {
		return "", os.ErrNotExist
	}

	return m.hookPath, nil
}

func (m *mockPathValidator) EnsureHooksDirectory(_ string) error {
	if m.shouldError {
		return os.ErrNotExist
	}

	if m.hooksDir != "" {
		return os.MkdirAll(m.hooksDir, 0755)
	}

	return nil
}

// Ensure mock implements the interface.
var _ cliAdapter.PathValidator = (*mockPathValidator)(nil)
