// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRuleEnabled(t *testing.T) {
	t.Run("Rule is enabled when in enabled_rules", func(t *testing.T) {
		enabledRules := []string{"Rule1", "Rule2", "JiraReference"}
		disabledRules := []string{"Rule3"}

		result := IsRuleEnabled("Rule1", enabledRules, disabledRules)
		require.True(t, result, "Rule should be enabled when in enabled_rules")

		result = IsRuleEnabled("JiraReference", enabledRules, disabledRules)
		require.True(t, result, "JiraReference should be enabled when in enabled_rules")
	})

	t.Run("Rule is disabled when in disabled_rules and not in enabled_rules", func(t *testing.T) {
		enabledRules := []string{"Rule1", "Rule2"}
		disabledRules := []string{"Rule3", "JiraReference"}

		result := IsRuleEnabled("Rule3", enabledRules, disabledRules)
		require.False(t, result, "Rule should be disabled when in disabled_rules")

		result = IsRuleEnabled("JiraReference", enabledRules, disabledRules)
		require.False(t, result, "JiraReference should be disabled when in disabled_rules")
	})

	t.Run("disabled_rules takes precedence over enabled_rules", func(t *testing.T) {
		enabledRules := []string{"Rule1", "JiraReference"}
		disabledRules := []string{"Rule1", "JiraReference"}

		result := IsRuleEnabled("Rule1", enabledRules, disabledRules)
		require.False(t, result, "Rule should be disabled when in both lists")

		result = IsRuleEnabled("JiraReference", enabledRules, disabledRules)
		require.False(t, result, "JiraReference should be disabled when in both lists")
	})

	t.Run("Rule is enabled when not in any list (default behavior)", func(t *testing.T) {
		enabledRules := []string{"Rule1"}
		disabledRules := []string{"Rule3"}

		result := IsRuleEnabled("Rule2", enabledRules, disabledRules)
		require.True(t, result, "Rule should be enabled by default when not in any list")
	})
}

// func TestIsJiraRuleEnabled(t *testing.T) {
// 	t.Run("JiraReference is enabled when explicitly in enabled_rules", func(t *testing.T) {
// 		cfg := NewDefaultConfig()
// 		cfg = cfg.WithRules(cfg.Rules.WithEnabledRules([]string{"JiraReference"}))

// 		result := IsJiraRuleEnabled(cfg)
// 		require.True(t, result, "JiraReference should be enabled when in enabled_rules")
// 	})

// 	t.Run("JiraReference is disabled when in disabled_rules", func(t *testing.T) {
// 		cfg := NewDefaultConfig()
// 		cfg = cfg.WithRules(cfg.Rules.WithDisabledRules([]string{"JiraReference"}))

// 		// Make sure it's not in enabled_rules
// 		cfg = cfg.WithRules(cfg.Rules.WithEnabledRules([]string{"Rule1"}))

// 		result := IsJiraRuleEnabled(cfg)
// 		require.False(t, result, "JiraReference should be disabled when in disabled_rules")
// 	})

// 	t.Run("JiraReference is enabled when Jira.Required is true", func(t *testing.T) {
// 		cfg := NewDefaultConfig()
// 		cfg = cfg.WithJira(cfg.Jira.WithRequired(true))

// 		result := IsJiraRuleEnabled(cfg)
// 		require.True(t, result, "JiraReference should be enabled when Jira.Required is true")
// 	})

// 	t.Run("JiraReference is disabled by default", func(t *testing.T) {
// 		cfg := NewDefaultConfig()

// 		// Make sure it's not in enabled_rules and Jira.Required is false
// 		cfg = cfg.WithRules(cfg.Rules.WithEnabledRules([]string{"Rule1"}))
// 		cfg = cfg.WithJira(cfg.Jira.WithRequired(false))

// 		result := IsJiraRuleEnabled(cfg)
// 		require.False(t, result, "JiraReference should be disabled by default")
// 	})

// 	t.Run("JiraReference is enabled when in config file", func(t *testing.T) {
// 		// This test requires writing a temp config file
// 		tempDir, err := os.MkdirTemp("", "gommitlint-jira-test")
// 		require.NoError(t, err)
// 		defer os.RemoveAll(tempDir)

// 		// Create a config file that mentions JiraReference
// 		configPath := filepath.Join(tempDir, ".gommitlint.yaml")
// 		configContent := `gommitlint:
//   rules:
//     enabled_rules:
//       - JiraReference
// `
// 		err = os.WriteFile(configPath, []byte(configContent), 0600)
// 		require.NoError(t, err)

// 		// Change to the temp directory to make the config file discoverable
// 		oldDir, err := os.Getwd()
// 		require.NoError(t, err)

// 		err = os.Chdir(tempDir)
// 		require.NoError(t, err)

// 		defer func() {
// 			err := os.Chdir(oldDir)
// 			if err != nil {
// 				t.Logf("Failed to change back to original directory: %v", err)
// 			}
// 		}()

// 		// Now test with a default config
// 		cfg := NewDefaultConfig()

// 		// Make sure JiraReference is in disabled_rules by default
// 		cfg = cfg.WithRules(cfg.Rules.WithDisabledRules([]string{"JiraReference"}))

// 		// The function should find it in the config file
// 		result := IsJiraRuleEnabled(cfg)
// 		require.True(t, result, "JiraReference should be enabled when it's in the config file")
// 	})
// }
