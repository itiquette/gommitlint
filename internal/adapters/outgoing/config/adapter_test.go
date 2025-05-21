// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration handling for gommitlint.
package config_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	configTypes "github.com/itiquette/gommitlint/internal/config/types"
)

func TestAdapter_InterfaceCompliance(t *testing.T) {
	// Create a test config
	cfg := configTypes.Config{
		Message: configTypes.MessageConfig{
			Subject: configTypes.SubjectConfig{
				MaxLength: 72,
			},
			Body: configTypes.BodyConfig{
				MinLength:        10,
				AllowSignoffOnly: true,
			},
		},
		Rules: configTypes.RulesConfig{
			Enabled:  []string{"subject.max_length", "body.required"},
			Disabled: []string{"spell"},
		},
		Signing: configTypes.SigningConfig{
			RequireSignature:      true,
			AllowMultipleSignoffs: false,
		},
		Jira: configTypes.JiraConfig{
			Projects: []string{"PROJ1", "PROJ2"},
		},
		Conventional: configTypes.ConventionalConfig{
			Types: []string{"feat", "fix", "docs"},
		},
		Repo: configTypes.RepoConfig{
			Path: "/test/repo",
		},
	}

	// Create the adapter
	adapter := config.NewAdapter(cfg)

	// Test common config methods
	require.Equal(t, 72, adapter.GetInt("message.subject.max_length"))
	require.True(t, adapter.GetBool("message.body.allow_signoff_only"))
	require.Equal(t, []string{"subject.max_length", "body.required"}, adapter.GetStringSlice("rules.enabled"))

	// Test rule enabled/disabled methods
	// Create a test context
	ctx := context.Background()
	require.True(t, adapter.IsRuleEnabled(ctx, "subject.max_length"))
	require.False(t, adapter.IsRuleEnabled(ctx, "spell"))
	require.True(t, adapter.IsRuleDisabled(ctx, "spell"))
}

func TestAdapter_SpecificMethods(t *testing.T) {
	// Create a test config
	cfg := configTypes.Config{
		Message: configTypes.MessageConfig{
			Subject: configTypes.SubjectConfig{
				MaxLength: 50,
			},
			Body: configTypes.BodyConfig{
				MinLength:        20,
				AllowSignoffOnly: false,
			},
		},
		Jira: configTypes.JiraConfig{
			Pattern:  "JIRA-[0-9]+",
			Projects: []string{"PROJ"},
		},
		Conventional: configTypes.ConventionalConfig{
			Types: []string{"feat", "fix"},
		},
	}

	// Create the adapter
	adapter := config.NewAdapter(cfg)

	// Test specific getters
	require.Equal(t, 50, adapter.GetInt("message.subject.max_length"))
	require.Equal(t, 20, adapter.GetInt("message.body.min_length"))
	require.Equal(t, "JIRA-[0-9]+", adapter.GetString("jira.pattern"))
	require.Equal(t, []string{"PROJ"}, adapter.GetStringSlice("jira.projects"))
	require.Equal(t, []string{"feat", "fix"}, adapter.GetStringSlice("conventional.types"))
}

func TestAdapter_NestedConfig(t *testing.T) {
	// Create a complex nested config
	cfg := configTypes.Config{
		Message: configTypes.MessageConfig{
			Subject: configTypes.SubjectConfig{
				MaxLength:         100,
				RequireImperative: true,
				Case:              "sentence",
			},
			Body: configTypes.BodyConfig{
				MinLength: 50,
				MinLines:  3,
			},
		},
		Jira: configTypes.JiraConfig{
			Pattern:  "([A-Z]+-[0-9]+)",
			Projects: []string{"PROJ1", "PROJ2", "PROJ3"},
		},
	}

	// Create the adapter
	adapter := config.NewAdapter(cfg)

	// Test nested access
	require.Equal(t, 100, adapter.GetInt("message.subject.max_length"))
	require.True(t, adapter.GetBool("message.subject.require_imperative"))
	require.Equal(t, "sentence", adapter.GetString("message.subject.case"))
	require.Equal(t, "([A-Z]+-[0-9]+)", adapter.GetString("jira.pattern"))
	require.Len(t, adapter.GetStringSlice("jira.projects"), 3)
}

func TestAdapter_EmptyValues(t *testing.T) {
	// Create a config with empty values
	cfg := configTypes.Config{}

	// Create the adapter
	adapter := config.NewAdapter(cfg)

	// Test that empty values return defaults
	require.Equal(t, 0, adapter.GetInt("message.subject.max_length"))
	require.False(t, adapter.GetBool("message.body.allow_signoff_only"))
	require.Equal(t, "", adapter.GetString("jira.pattern"))
	require.Empty(t, adapter.GetStringSlice("conventional.types"))
}

func TestAdapter_InvalidPaths(t *testing.T) {
	// Create a config
	cfg := configTypes.Config{
		Message: configTypes.MessageConfig{
			Subject: configTypes.SubjectConfig{
				MaxLength: 72,
			},
		},
	}

	// Create the adapter
	adapter := config.NewAdapter(cfg)

	// Test invalid paths
	require.Equal(t, 0, adapter.GetInt("invalid.path"))
	require.False(t, adapter.GetBool("another.invalid.path"))
	require.Equal(t, "", adapter.GetString("not.exists"))
	require.Empty(t, adapter.GetStringSlice("invalid.array"))
}

func TestAdapter_DeepNestedPaths(t *testing.T) {
	// Create a config with deep nesting
	cfg := configTypes.Config{
		Jira: configTypes.JiraConfig{
			Projects: []string{"P1", "P2", "P3", "P4", "P5"},
		},
	}

	// Create the adapter
	adapter := config.NewAdapter(cfg)

	// Test array access
	projects := adapter.GetStringSlice("jira.projects")
	require.Len(t, projects, 5)
	require.Equal(t, "P1", projects[0])
	require.Equal(t, "P5", projects[4])

	// Test individual array element access
	require.Equal(t, "P1", adapter.GetString("jira.projects.0"))
	require.Equal(t, "P5", adapter.GetString("jira.projects.4"))
}
