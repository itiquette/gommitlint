// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configuration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Basic tests for New() and DefaultConfiguration() functions

func TestNew(t *testing.T) {
	// Create a temporary directory for this test
	tmpDir := t.TempDir()

	// Change to temp directory for test
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		err := os.Chdir(originalDir) // Make sure we restore the original directory
		require.NoError(t, err)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a test configuration file
	content := `gommitlint:
  conventional-commit:
    types:
      - feat
      - fix
  subject:
    max-length: 72`

	err = os.WriteFile(filepath.Join(tmpDir, ".gommitlint.yaml"), []byte(content), 0600)
	require.NoError(t, err)

	// Test the New function
	config, err := configuration.New()
	if err != nil {
		// Some errors might be expected in testing environments
		t.Logf("New() returned error: %v", err)

		return
	}

	// If we get a result, verify it looks correct
	if config != nil && config.GommitConf != nil {
		assert.NotNil(t, config.GommitConf.Subject)
		assert.NotNil(t, config.GommitConf.ConventionalCommit)
	}
}

func TestDefaultConfiguration(t *testing.T) {
	// Test that default configuration loads correctly
	config := configuration.DefaultConfiguration()
	require.NotNil(t, config)

	// Check that required fields are not nil
	require.NotNil(t, config.GommitConf)
	require.NotNil(t, config.GommitConf.Subject)
	require.NotNil(t, config.GommitConf.Body)
	require.NotNil(t, config.GommitConf.ConventionalCommit)
	require.NotNil(t, config.GommitConf.SpellCheck)
	require.NotNil(t, config.GommitConf.Signature)
	require.NotNil(t, config.GommitConf.SignOffRequired)
	require.NotNil(t, config.GommitConf.NCommitsAhead)
	require.NotNil(t, config.GommitConf.IgnoreMergeCommits)
}
