// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestHookRemovalParameters(t *testing.T) {
	// Create a mock CLI command
	var buf bytes.Buffer
	cmd := &cli.Command{
		Writer: &buf,
		Reader: strings.NewReader("y\n"),
	}

	params := NewHookRemovalParameters(cmd, "/test/repo", false)

	require.Equal(t, "/test/repo", params.RepoPath)
	require.False(t, params.SkipConfirm)
	require.Equal(t, "commit-msg", params.HookType)
	require.Equal(t, &buf, params.Output)
	require.NotNil(t, params.Input)
	require.NotNil(t, params.PathValidator)
}

func TestHookRemovalParameters_IsGommitlintHook(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		hookContent string
		wantResult  bool
		wantErr     bool
	}{
		{
			name:        "gommitlint hook",
			hookContent: createDefaultHookScript(),
			wantResult:  true,
			wantErr:     false,
		},
		{
			name:        "custom hook with gommitlint mention",
			hookContent: "#!/bin/sh\n# Uses gommitlint\necho test",
			wantResult:  true,
			wantErr:     false,
		},
		{
			name:        "non-gommitlint hook",
			hookContent: "#!/bin/sh\necho 'custom validation'",
			wantResult:  false,
			wantErr:     false,
		},
		{
			name:        "large hook file",
			hookContent: strings.Repeat("x", 11*1024), // Over 10KB limit
			wantResult:  false,
			wantErr:     true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			hookPath := filepath.Join(tmpDir, testCase.name)
			require.NoError(t, os.WriteFile(hookPath, []byte(testCase.hookContent), 0600))

			params := HookRemovalParameters{
				PathValidator: &mockRemovalPathValidator{
					hookPath: hookPath,
				},
			}

			result, err := params.IsGommitlintHook()
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.wantResult, result)
			}
		})
	}
}

func TestHookRemovalParameters_ConfirmRemoval(t *testing.T) {
	tests := []struct {
		name        string
		skipConfirm bool
		userInput   string
		wantResult  bool
	}{
		{
			name:        "skip confirmation",
			skipConfirm: true,
			userInput:   "",
			wantResult:  true,
		},
		{
			name:        "user confirms with y",
			skipConfirm: false,
			userInput:   "y\n",
			wantResult:  true,
		},
		{
			name:        "user confirms with yes",
			skipConfirm: false,
			userInput:   "yes\n",
			wantResult:  true,
		},
		{
			name:        "user declines with n",
			skipConfirm: false,
			userInput:   "n\n",
			wantResult:  false,
		},
		{
			name:        "user declines with empty",
			skipConfirm: false,
			userInput:   "\n",
			wantResult:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			params := HookRemovalParameters{
				SkipConfirm: testCase.skipConfirm,
				Output:      &output,
				Input:       strings.NewReader(testCase.userInput),
			}

			result, err := params.ConfirmRemoval()
			// Empty input causes an expected error (unexpected newline), which is OK
			if testCase.userInput == "\n" && err != nil {
				require.False(t, result, "empty input should decline")
				require.Contains(t, err.Error(), "unexpected newline")

				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.wantResult, result)

			if !testCase.skipConfirm {
				outputStr := output.String()
				require.Contains(t, outputStr, "Warning")
				require.Contains(t, outputStr, "Do you still want to remove it?")
			}
		})
	}
}

func TestRemoveHookFile(t *testing.T) {
	// Test successful removal
	t.Run("successful removal", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "hook")
		require.NoError(t, os.WriteFile(tmpFile, []byte("test"), 0600))

		err := RemoveHookFile(tmpFile)
		require.NoError(t, err)

		// Verify file is gone
		_, err = os.Stat(tmpFile)
		require.True(t, os.IsNotExist(err))
	})

	// Test removal of non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		err := RemoveHookFile("/non/existent/file")
		require.Error(t, err)
		require.Contains(t, err.Error(), "could not remove hook file")
	})
}

// mockRemovalPathValidator is a test implementation of PathValidator for removal tests.
type mockRemovalPathValidator struct {
	hookPath    string
	hooksDir    string
	shouldError bool
}

func (m *mockRemovalPathValidator) ValidateHookPath(_, _ string) (string, error) {
	if m.shouldError {
		return "", os.ErrNotExist
	}

	return m.hookPath, nil
}

func (m *mockRemovalPathValidator) EnsureHooksDirectory(_ string) error {
	if m.shouldError {
		return os.ErrNotExist
	}

	if m.hooksDir != "" {
		return os.MkdirAll(m.hooksDir, 0755)
	}

	return nil
}

// Ensure mock implements the interface.
var _ cliAdapter.PathValidator = (*mockRemovalPathValidator)(nil)
