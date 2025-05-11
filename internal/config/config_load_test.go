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

func TestLoadConfigFromYAML(t *testing.T) {
	// Create a temporary directory for the test config file
	tempDir, err := os.MkdirTemp("", "gommitlint-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file with JiraReference in enabled_rules
	configPath := filepath.Join(tempDir, ".gommitlint.yaml")
	configContent := `gommitlint:
  rules:
    enabled_rules:
      - JiraReference
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Create a provider and load the config
	provider, err := NewProvider()
	require.NoError(t, err)

	// Load the config from the test file
	err = provider.LoadFromPath(configPath)
	require.NoError(t, err)

	// Get the config
	config := provider.GetConfig()

	// Verify that JiraReference is in the enabled_rules
	found := false

	for _, rule := range config.Rules.EnabledRules {
		if rule == "JiraReference" {
			found = true

			break
		}
	}

	require.True(t, found, "JiraReference should be in enabled_rules after loading from YAML")

	// Verify that JiraReference is not in the disabled_rules (if it was initially added there)
	for _, rule := range config.Rules.DisabledRules {
		require.NotEqual(t, "JiraReference", rule, "JiraReference should not be in disabled_rules after loading from YAML")
	}
}

func TestJiraReferenceEnabledOverridesDisabled(t *testing.T) {
	// Create a temporary directory for the test config file
	tempDir, err := os.MkdirTemp("", "gommitlint-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file with JiraReference in enabled_rules
	configPath := filepath.Join(tempDir, ".gommitlint.yaml")
	configContent := `gommitlint:
  rules:
    enabled_rules:
      - JiraReference
    disabled_rules:
      - JiraReference
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Create a provider and load the config
	provider, err := NewProvider()
	require.NoError(t, err)

	// The provider should add JiraReference to disabled_rules by default
	disabledRules := provider.GetConfig().Rules.DisabledRules
	foundInDisabled := false

	for _, rule := range disabledRules {
		if rule == "JiraReference" {
			foundInDisabled = true

			break
		}
	}

	require.True(t, foundInDisabled, "JiraReference should be in disabled_rules by default")

	// Now load the config file where it's both enabled and disabled
	err = provider.LoadFromPath(configPath)
	require.NoError(t, err)

	// Get the config
	config := provider.GetConfig()

	// Print the full config for debugging
	t.Logf("Enabled rules: %v", config.Rules.EnabledRules)
	t.Logf("Disabled rules: %v", config.Rules.DisabledRules)

	// Verify that JiraReference is in the enabled_rules
	foundInEnabled := false

	for _, rule := range config.Rules.EnabledRules {
		if rule == "JiraReference" {
			foundInEnabled = true

			break
		}
	}

	require.True(t, foundInEnabled, "JiraReference should be in enabled_rules after loading from YAML")

	// The critical test: verify that JiraReference is not in the disabled_rules
	// Because being explicitly enabled should override being disabled
	for _, rule := range config.Rules.DisabledRules {
		require.NotEqual(t, "JiraReference", rule,
			"JiraReference should NOT be in disabled_rules when it's explicitly enabled in config")
	}
}
