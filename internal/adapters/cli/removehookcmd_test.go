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

	gitTestdata "github.com/itiquette/gommitlint/internal/adapters/git/testdata"
)

// TestHookRemovalParameters tests the HookRemovalParameters type.
func TestHookRemovalParameters(t *testing.T) {
	// Create a test Git repository
	tmpDir, cleanup := gitTestdata.GitRepo(t, "Initial commit")
	defer cleanup()

	// Create a mock cobra command
	mockCmd := &cobra.Command{}

	// Test the basic parameters
	params := NewHookRemovalParameters(mockCmd, tmpDir, false)
	require.False(t, params.SkipConfirm)
	require.Equal(t, tmpDir, params.RepoPath)
	require.Equal(t, "commit-msg", params.HookType)
	require.NotNil(t, params.Output)
	require.NotNil(t, params.Input)

	// Test immutability with methods
	skipParams := params.WithSkipConfirm()
	require.False(t, params.SkipConfirm, "Original should be unchanged")
	require.True(t, skipParams.SkipConfirm, "New instance should have updated value")

	confirmParams := skipParams.WithoutSkipConfirm()
	require.True(t, skipParams.SkipConfirm, "Original should be unchanged")
	require.False(t, confirmParams.SkipConfirm, "New instance should have updated value")

	customHookParams := params.Hook("custom-hook")
	require.Equal(t, "commit-msg", params.HookType, "Original should be unchanged")
	require.Equal(t, "custom-hook", customHookParams.HookType, "New instance should have updated value")

	// Test the methods - first create a hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	err := os.MkdirAll(hooksDir, 0755)
	require.NoError(t, err)

	hookPath := filepath.Join(hooksDir, "commit-msg")
	err = os.WriteFile(hookPath, []byte("#!/bin/sh\n# gommitlint commit-msg hook\n"), 0600)
	require.NoError(t, err)

	// Test FindHookPath
	foundPath, err := params.FindHookPath()
	require.NoError(t, err)
	require.Equal(t, hookPath, foundPath)

	// Test VerifyHookExists
	err = params.VerifyHookExists()
	require.NoError(t, err)

	// Test IsGommitlintHook
	isOurs, err := params.IsGommitlintHook()
	require.NoError(t, err)
	require.True(t, isOurs)

	// Test with a non-gommitlint hook
	err = os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'regular hook'\n"), 0600)
	require.NoError(t, err)

	isOurs, err = params.IsGommitlintHook()
	require.NoError(t, err)
	require.False(t, isOurs)

	// Test RemoveHookFile (separated side effect)
	hookPathToRemove, err := params.GetHookPath()
	require.NoError(t, err)
	require.Equal(t, hookPath, hookPathToRemove)

	err = RemoveHookFile(hookPathToRemove)
	require.NoError(t, err)

	_, err = os.Stat(hookPath)
	require.True(t, os.IsNotExist(err))
}

// TestRemoveHook tests the removeHook function.
func TestRemoveHook(t *testing.T) {
	// Create a test Git repository
	tmpDir, cleanup := gitTestdata.GitRepo(t, "Initial commit")
	defer cleanup()

	// Create a mock cobra command
	mockCmd := &cobra.Command{}

	// Create hook directories
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	err := os.MkdirAll(hooksDir, 0755)
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
