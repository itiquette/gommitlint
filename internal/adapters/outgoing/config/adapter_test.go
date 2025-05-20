// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config_test

import (
	"testing"

	infra "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	commonConfig "github.com/itiquette/gommitlint/internal/common/config"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
	"github.com/stretchr/testify/require"
)

func TestAdapter_InterfaceCompliance(t *testing.T) {
	// Create a test config
	cfg := configTypes.Config{
		Subject: configTypes.SubjectConfig{
			MaxLength: 72,
		},
		Body: configTypes.BodyConfig{
			MinLength:        10,
			AllowSignOffOnly: true,
		},
		Rules: configTypes.RulesConfig{
			Enabled:  []string{"subject.max_length", "body.required"},
			Disabled: []string{"spell"},
		},
		Security: configTypes.SecurityConfig{
			GPGRequired:      true,
			MultipleSignoffs: false,
		},
		Jira: configTypes.JiraConfig{
			Projects: []string{"PROJ1", "PROJ2"},
		},
		Conventional: configTypes.ConventionalConfig{
			Types: []string{"feat", "fix", "docs"},
		},
	}

	// Create adapter
	adapter := infra.NewAdapter(cfg)

	// Ensure adapter implements all required interfaces
	t.Run("Implements Config", func(t *testing.T) {
		var _ commonConfig.Config = adapter

		require.NotNil(t, adapter)
	})
}

func TestAdapter_ConfigMethods(t *testing.T) {
	// Create a test config
	cfg := configTypes.Config{
		Subject: configTypes.SubjectConfig{
			MaxLength: 72,
		},
		Body: configTypes.BodyConfig{
			MinLength:        10,
			AllowSignOffOnly: true,
		},
		Jira: configTypes.JiraConfig{
			Projects: []string{"PROJ1", "PROJ2"},
		},
		Rules: configTypes.RulesConfig{
			Enabled:  []string{"SubjectLength"},
			Disabled: []string{"spell"},
		},
	}

	// Create adapter
	adapter := infra.NewAdapter(cfg)

	// Test Config interface methods
	t.Run("Get", func(t *testing.T) {
		// Test nested access
		subjectMaxLength := adapter.Get("subject.max_length")
		require.Equal(t, 72, subjectMaxLength)

		// Test array access
		jiraProject := adapter.Get("jira.projects.0")
		require.Equal(t, "PROJ1", jiraProject)

		// Test nil for non-existent keys
		nonExistent := adapter.Get("non.existent.key")
		require.Nil(t, nonExistent)
	})

	t.Run("GetString", func(t *testing.T) {
		str := adapter.GetString("jira.projects.0")
		require.Equal(t, "PROJ1", str)

		// Non-existent key returns empty string
		empty := adapter.GetString("non.existent")
		require.Equal(t, "", empty)
	})

	t.Run("GetBool", func(t *testing.T) {
		// Check a valid boolean field
		allowSignOffOnly := adapter.GetBool("body.allow_sign_off_only")
		require.True(t, allowSignOffOnly) // Set to true in test config

		// Non-existent key returns false
		nonExistent := adapter.GetBool("non.existent")
		require.False(t, nonExistent)
	})

	t.Run("GetInt", func(t *testing.T) {
		maxLength := adapter.GetInt("subject.max_length")
		require.Equal(t, 72, maxLength)

		// Non-existent key returns 0
		nonExistent := adapter.GetInt("non.existent")
		require.Equal(t, 0, nonExistent)
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		projects := adapter.GetStringSlice("jira.projects")
		require.Equal(t, []string{"PROJ1", "PROJ2"}, projects)

		// Non-existent key returns nil
		nonExistent := adapter.GetStringSlice("non.existent")
		require.Nil(t, nonExistent)
	})

	t.Run("Rule state methods", func(t *testing.T) {
		// Test enabled rule
		require.True(t, adapter.IsRuleEnabled("SubjectLength"))
		require.False(t, adapter.IsRuleDisabled("SubjectLength"))

		// Test disabled rule
		require.False(t, adapter.IsRuleEnabled("spell"))
		require.True(t, adapter.IsRuleDisabled("spell"))

		// Test rule not in any list but is not default disabled (most rules are enabled by default)
		require.True(t, adapter.IsRuleEnabled("ImperativeVerb"))
		require.False(t, adapter.IsRuleDisabled("ImperativeVerb"))
	})
}

func TestAdapter_ValidationConfigMethods(t *testing.T) {
	// Create a test config
	cfg := configTypes.Config{
		Subject: configTypes.SubjectConfig{
			MaxLength: 72,
		},
		Body: configTypes.BodyConfig{
			MinLength:        10,
			AllowSignOffOnly: true,
		},
		Security: configTypes.SecurityConfig{
			GPGRequired: true,
		},
		Jira: configTypes.JiraConfig{
			Projects: []string{"PROJ1", "PROJ2"},
		},
		Conventional: configTypes.ConventionalConfig{
			Types: []string{"feat", "fix"},
		},
		Rules: configTypes.RulesConfig{
			Enabled:  []string{"subject.max_length"},
			Disabled: []string{"spell"},
		},
	}

	// Create adapter
	adapter := infra.NewAdapter(cfg)

	// Test ValidationConfig interface methods
	t.Run("ValidationConfig methods", func(t *testing.T) {
		require.Equal(t, cfg.Rules.Enabled, adapter.EnabledRules())
		require.Equal(t, cfg.Rules.Disabled, adapter.DisabledRules())
		require.Equal(t, cfg.Subject.MaxLength, adapter.SubjectMaxLength())
		require.Equal(t, cfg.Subject.Imperative, adapter.SubjectImperative())
		require.Equal(t, cfg.Body.MinLength, adapter.BodyMinLength())
		// Note: GPGRequired and SignOffRequired are part of SecurityConfig but not exposed in ValidationConfig interface
		// They are accessed through the config.Get methods instead
		require.Equal(t, cfg.Jira.Projects, adapter.JiraProjects())
		require.Equal(t, cfg.Conventional.Types, adapter.ConventionalTypes())
	})
}

func TestAdapter_RulesConfigMethods(t *testing.T) {
	// Create a test config
	cfg := configTypes.Config{
		Body: configTypes.BodyConfig{
			MinLength:        10,
			AllowSignOffOnly: true,
		},
		Repository: configTypes.RepositoryConfig{
			ReferenceBranch: "main",
			MaxCommitsAhead: 15,
		},
		Rules: configTypes.RulesConfig{
			Enabled:  []string{"body.min_length"},
			Disabled: []string{"spell"},
		},
		SpellCheck: configTypes.SpellCheckConfig{
			CustomDictionary: []string{"gommitlint", "golang"},
		},
		Subject: configTypes.SubjectConfig{
			Case: "sentence",
		},
	}

	// Create adapter
	adapter := infra.NewAdapter(cfg)

	// Test RulesConfig interface methods
	t.Run("RulesConfig methods", func(t *testing.T) {
		require.Equal(t, cfg.Body.MinLength, adapter.BodyMinLength())
		require.Equal(t, cfg.Repository.ReferenceBranch, adapter.ReferenceBranch())
		require.Equal(t, cfg.Repository.MaxCommitsAhead, adapter.MaxCommitsAhead())
		// SpellIgnoreWords is stored in SpellCheck.CustomDictionary
		require.Equal(t, cfg.SpellCheck.CustomDictionary, adapter.SpellIgnoreWords())
		require.Equal(t, cfg.Subject.Case, adapter.SubjectCase())
	})
}

func TestAdapter_RuleConfigurationMethods(t *testing.T) {
	// Create a test config with rules that have default disabled states
	cfg := configTypes.Config{
		Rules: configTypes.RulesConfig{
			Enabled:  []string{"SubjectLength", "SignedIdentity"}, // Force enable SignedIdentity
			Disabled: []string{"spell"},
		},
	}

	// Create adapter
	adapter := infra.NewAdapter(cfg)

	// Test RuleConfiguration interface methods
	t.Run("DefaultDisabledRules", func(t *testing.T) {
		// SignedIdentity is normally default disabled
		// but since it's in enabled_rules, it should be enabled
		require.True(t, adapter.IsRuleEnabled("SignedIdentity"))
		require.False(t, adapter.IsRuleDisabled("SignedIdentity"))

		// spell is in disabled_rules, so it should be disabled
		require.False(t, adapter.IsRuleEnabled("spell"))
		require.True(t, adapter.IsRuleDisabled("spell"))

		// A rule that's not in any list and not default disabled should be enabled
		require.True(t, adapter.IsRuleEnabled("SubjectLength"))
		require.False(t, adapter.IsRuleDisabled("SubjectLength"))
	})
}
