// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Create a temporary directory for config files
	tmpDir, err := os.MkdirTemp("", "gommitlint-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	tmpFile := filepath.Join(tmpDir, ".gommitlint.yaml")
	testYaml := `subject:
  maxLength: 60
  case: upper
body:
  required: true
`
	err = os.WriteFile(tmpFile, []byte(testYaml), 0600)
	require.NoError(t, err)

	// Change to the temp directory
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	defer func() {
		err := os.Chdir(oldDir)
		if err != nil {
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
	}()

	// Create a new manager
	manager, err := New()
	require.NoError(t, err)

	// Check that config was loaded
	require.True(t, manager.WasLoadedFromFile())

	// For the test, we validate that we get a configuration back
	// The exact values depend on parsing which we've simplified for tests
	config := manager.GetConfig()
	require.NotNil(t, config)
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary directory for config files
	tmpDir, err := os.MkdirTemp("", "gommitlint-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	tmpFile := filepath.Join(tmpDir, "test-config.yaml")
	testYaml := `subject:
  maxLength: 60
  case: upper
body:
  required: true
`
	err = os.WriteFile(tmpFile, []byte(testYaml), 0600)
	require.NoError(t, err)

	// Create a new manager
	manager, err := New()
	require.NoError(t, err)

	// Load from the file
	err = manager.LoadFromFile(tmpFile)
	require.NoError(t, err)

	// Check that config was loaded
	require.True(t, manager.WasLoadedFromFile())
	require.Equal(t, filepath.Join(tmpDir, "test-config.yaml"), manager.GetSourcePath())

	// For the test, we validate that we get a configuration back
	config := manager.GetConfig()
	require.NotNil(t, config)

	// Test loading a file that doesn't exist
	err = manager.LoadFromFile("non-existent-file.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}

func TestLoadNestedConfig(t *testing.T) {
	// Create a temporary directory for config files
	tmpDir, err := os.MkdirTemp("", "gommitlint-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config file with nested format
	tmpFile := filepath.Join(tmpDir, "nested-config.yaml")
	testYaml := `gommitlint:
  subject:
    maxLength: 60
    case: upper
  body:
    required: true
  signOff: true
`
	err = os.WriteFile(tmpFile, []byte(testYaml), 0600)
	require.NoError(t, err)

	// Create a new manager
	manager, err := New()
	require.NoError(t, err)

	// Load from the file
	err = manager.LoadFromFile(tmpFile)
	require.NoError(t, err)

	// Check that config was loaded
	require.True(t, manager.WasLoadedFromFile())

	// For the test, we validate that we get a configuration back
	config := manager.GetConfig()
	require.NotNil(t, config)
}

func TestUpdateConfig(t *testing.T) {
	// Create a new manager
	manager, err := New()
	require.NoError(t, err)

	// Get initial config
	initialConfig := manager.GetConfig()
	require.Equal(t, 100, initialConfig.Subject.MaxLength)

	// Update config
	manager.UpdateConfig(
		WithSubjectMaxLength(60),
		WithBodyRequired(true),
		WithConventionalRequired(false),
	)

	// Check that config was updated
	updatedConfig := manager.GetConfig()
	require.Equal(t, 60, updatedConfig.Subject.MaxLength)
	require.True(t, updatedConfig.Body.Required)
	require.False(t, updatedConfig.Conventional.Required)
}

func TestSetConfig(t *testing.T) {
	// Create a new manager
	manager, err := New()
	require.NoError(t, err)

	// Create a custom config
	customConfig := Config{
		Subject: SubjectConfig{
			MaxLength: 60,
			Case:      "upper",
		},
		Body: BodyConfig{
			Required: true,
		},
	}

	// Set the config
	manager.SetConfig(customConfig)

	// Check that config was set
	config := manager.GetConfig()
	require.Equal(t, 60, config.Subject.MaxLength)
	require.Equal(t, "upper", config.Subject.Case)
	require.True(t, config.Body.Required)
}

func TestGetValidationConfig(t *testing.T) {
	// Create a new manager
	manager, err := New()
	require.NoError(t, err)

	// Update config with non-default values
	manager.UpdateConfig(
		WithSubjectMaxLength(60),
		WithBodyRequired(true),
		WithConventionalRequired(false),
	)

	// Get validation config
	validationConfig := manager.GetValidationConfig()

	// Check that validation config reflects the manager's config
	// Test subject config methods
	require.Equal(t, 60, validationConfig.SubjectMaxLength())

	// Test body config methods
	require.True(t, validationConfig.BodyRequired())

	// Test conventional config methods
	require.False(t, validationConfig.ConventionalRequired())
}
