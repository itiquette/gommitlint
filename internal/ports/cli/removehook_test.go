// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// TestRemoveHook tests the removeHook function.
func TestRemoveHook(t *testing.T) {
	// Create temp directory with git repo
	tmpDir := t.TempDir()
	err := setupGitRepo(tmpDir)
	require.NoError(t, err)

	// Create a mock cobra command
	mockCmd := &cobra.Command{}

	// Create hook directories
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	err = os.MkdirAll(hooksDir, 0600)
	require.NoError(t, err)

	// Create a gommitlint hook
	hookPath := filepath.Join(hooksDir, "commit-msg")
	err = os.WriteFile(hookPath, []byte("#!/bin/sh\n# gommitlint commit-msg hook\n"), 0600)
	require.NoError(t, err)

	// Test removing the gommitlint hook
	err = removeHook(mockCmd, tmpDir, false)
	require.NoError(t, err)

	// Hook should be removed
	_, err = os.Stat(hookPath)
	require.True(t, os.IsNotExist(err))

	// Create a non-gommitlint hook
	err = os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'regular hook'\n"), 0600)
	require.NoError(t, err)

	// Test with --yes flag, should remove the hook
	err = removeHook(mockCmd, tmpDir, true)
	require.NoError(t, err)

	// Hook should be removed
	_, err = os.Stat(hookPath)
	require.True(t, os.IsNotExist(err))
}
