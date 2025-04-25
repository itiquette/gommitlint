// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestValidationConfigAdapter(t *testing.T) {
	// Create a test config with non-default values
	config := Config{
		Subject: SubjectConfig{
			Case:            "upper",
			Imperative:      false,
			InvalidSuffixes: ".,:",
			MaxLength:       50,
			Jira: JiraConfig{
				Projects: []string{"PROJ"},
				Required: true,
				BodyRef:  true,
				Pattern:  "custom",
				Strict:   true,
			},
		},
		Body: BodyConfig{
			Required:         true,
			AllowSignOffOnly: true,
		},
		Conventional: ConventionalConfig{
			MaxDescriptionLength: 50,
			Scopes:               []string{"api", "ui"},
			Types:                []string{"feat", "fix"},
			Required:             false,
		},
		SpellCheck: SpellCheckConfig{
			Locale:      "US",
			Enabled:     false,
			IgnoreWords: []string{"foo", "bar"},
			CustomWords: map[string]string{"foo": "bar"},
			MaxErrors:   5,
		},
		Security: SecurityConfig{
			SignatureRequired:     false,
			AllowedSignatureTypes: []string{"gpg"},
			SignOffRequired:       false,
			AllowMultipleSignOffs: false,
			Identity: IdentityConfig{
				PublicKeyURI: "test",
			},
		},
		Repository: RepositoryConfig{
			Reference:          "develop",
			IgnoreMergeCommits: false,
			MaxCommitsAhead:    10,
			CheckCommitsAhead:  false,
		},
		Rules: RulesConfig{
			EnabledRules:  []string{"rule1"},
			DisabledRules: []string{"rule2"},
		},
	}

	// Create the adapter
	adapter := NewValidationConfigAdapter(config)

	// Verify adapter implements all the required interfaces
	var _ domain.SubjectConfigProvider = adapter

	var _ domain.JiraConfigProvider = adapter

	var _ domain.BodyConfigProvider = adapter

	var _ domain.ConventionalConfigProvider = adapter

	var _ domain.SecurityConfigProvider = adapter

	var _ domain.SpellCheckConfigProvider = adapter

	var _ domain.RepositoryConfigProvider = adapter

	var _ domain.RulesConfigProvider = adapter

	// Test subject methods
	t.Run("Subject methods", func(t *testing.T) {
		require.Equal(t, 50, adapter.SubjectMaxLength())
		require.Equal(t, "upper", adapter.SubjectCase())
		require.False(t, adapter.SubjectRequireImperative())
		require.Equal(t, ".,:", adapter.SubjectInvalidSuffixes())
	})

	// Test conventional commit methods
	t.Run("Conventional methods", func(t *testing.T) {
		require.Equal(t, []string{"feat", "fix"}, adapter.ConventionalTypes())
		require.Equal(t, []string{"api", "ui"}, adapter.ConventionalScopes())
		require.Equal(t, 50, adapter.ConventionalMaxDescriptionLength())
		require.False(t, adapter.ConventionalRequired())
	})

	// Test Jira methods
	t.Run("Jira methods", func(t *testing.T) {
		require.Equal(t, []string{"PROJ"}, adapter.JiraProjects())
		require.True(t, adapter.JiraBodyRef())
		require.True(t, adapter.JiraRequired())
		require.Equal(t, "custom", adapter.JiraPattern())
		require.True(t, adapter.JiraStrict())
	})

	// Test body methods
	t.Run("Body methods", func(t *testing.T) {
		require.True(t, adapter.BodyRequired())
		require.True(t, adapter.BodyAllowSignOffOnly())
	})

	// Test security methods
	t.Run("Security methods", func(t *testing.T) {
		require.False(t, adapter.SignatureRequired())
		require.Equal(t, []string{"gpg"}, adapter.AllowedSignatureTypes())
		require.False(t, adapter.SignOffRequired())
		require.False(t, adapter.AllowMultipleSignOffs())
		require.Equal(t, "test", adapter.IdentityPublicKeyURI())
	})

	// Test spell check methods
	t.Run("Spell check methods", func(t *testing.T) {
		require.Equal(t, "US", adapter.SpellLocale())
		require.False(t, adapter.SpellEnabled())
		require.Equal(t, []string{"foo", "bar"}, adapter.SpellIgnoreWords())
		require.Equal(t, map[string]string{"foo": "bar"}, adapter.SpellCustomWords())
		require.Equal(t, 5, adapter.SpellMaxErrors())
	})

	// Test repository methods
	t.Run("Repository methods", func(t *testing.T) {
		require.Equal(t, "develop", adapter.ReferenceBranch())
		require.False(t, adapter.IgnoreMergeCommits())
		require.Equal(t, 10, adapter.MaxCommitsAhead())
		require.False(t, adapter.CheckCommitsAhead())
	})

	// Test rules methods
	t.Run("Rules methods", func(t *testing.T) {
		require.Equal(t, []string{"rule1"}, adapter.EnabledRules())
		require.Equal(t, []string{"rule2"}, adapter.DisabledRules())
	})
}
