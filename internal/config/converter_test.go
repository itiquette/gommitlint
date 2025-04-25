// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromGommitlintConfig(t *testing.T) {
	t.Run("empty config", func(t *testing.T) {
		gommitCfg := GommitlintConfig{}
		config := FromGommitlintConfig(gommitCfg)

		// Should contain default values
		require.Equal(t, DefaultConfig(), config)
	})

	t.Run("full config", func(t *testing.T) {
		gommitCfg := GommitlintConfig{
			Subject: SubjectConfig{
				Case:            "upper",
				Imperative:      true,
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
			ConventionalCommit: ConventionalConfig{
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
			Signature: SignatureConfig{
				Required: false,
				Identity: IdentityConfig{
					PublicKeyURI: "test",
				},
			},
			SignOffRequired:    false,
			NCommitsAhead:      true,
			IgnoreMergeCommits: false,
			Reference:          "develop",
		}

		config := FromGommitlintConfig(gommitCfg)

		// Check subject config
		require.Equal(t, "upper", config.Subject.Case)
		require.True(t, config.Subject.Imperative)
		require.Equal(t, ".,:", config.Subject.InvalidSuffixes)
		require.Equal(t, 50, config.Subject.MaxLength)

		// Check Jira config
		require.Equal(t, []string{"PROJ"}, config.Subject.Jira.Projects)
		require.True(t, config.Subject.Jira.Required)
		require.True(t, config.Subject.Jira.BodyRef)
		require.Equal(t, "custom", config.Subject.Jira.Pattern)
		require.True(t, config.Subject.Jira.Strict)

		// Check body config
		require.True(t, config.Body.Required)
		require.True(t, config.Body.AllowSignOffOnly)

		// Check conventional config
		require.Equal(t, 50, config.Conventional.MaxDescriptionLength)
		require.Equal(t, []string{"api", "ui"}, config.Conventional.Scopes)
		require.Equal(t, []string{"feat", "fix"}, config.Conventional.Types)
		require.False(t, config.Conventional.Required)

		// Check spell check config
		require.Equal(t, "US", config.SpellCheck.Locale)
		require.False(t, config.SpellCheck.Enabled)
		require.Equal(t, []string{"foo", "bar"}, config.SpellCheck.IgnoreWords)
		require.Equal(t, map[string]string{"foo": "bar"}, config.SpellCheck.CustomWords)
		require.Equal(t, 5, config.SpellCheck.MaxErrors)

		// Check security config
		require.False(t, config.Security.SignatureRequired)
		require.Equal(t, "test", config.Security.Identity.PublicKeyURI)
		require.False(t, config.Security.SignOffRequired)

		// Check repository config
		require.True(t, config.Repository.CheckCommitsAhead)
		require.False(t, config.Repository.IgnoreMergeCommits)
		require.Equal(t, "develop", config.Repository.Reference)
	})
}

func TestToGommitlintConfig(t *testing.T) {
	t.Run("convert internal config to gommitlint config", func(t *testing.T) {
		// Create internal config
		config := Config{
			Subject: SubjectConfig{
				Case:            "upper",
				Imperative:      true,
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
				CheckCommitsAhead:  true,
			},
		}

		// Convert to gommitlint config
		gommitCfg := ToGommitlintConfig(config)

		// Verify conversion
		require.Equal(t, "upper", gommitCfg.Subject.Case)
		require.True(t, gommitCfg.Subject.Imperative)
		require.Equal(t, 50, gommitCfg.Subject.MaxLength)
		require.Equal(t, ".,:", gommitCfg.Subject.InvalidSuffixes)

		require.Equal(t, []string{"PROJ"}, gommitCfg.Subject.Jira.Projects)
		require.True(t, gommitCfg.Subject.Jira.Required)

		require.True(t, gommitCfg.Body.Required)
		require.True(t, gommitCfg.Body.AllowSignOffOnly)

		require.Equal(t, 50, gommitCfg.ConventionalCommit.MaxDescriptionLength)
		require.Equal(t, []string{"api", "ui"}, gommitCfg.ConventionalCommit.Scopes)
		require.Equal(t, []string{"feat", "fix"}, gommitCfg.ConventionalCommit.Types)
		require.False(t, gommitCfg.ConventionalCommit.Required)

		require.Equal(t, "US", gommitCfg.SpellCheck.Locale)
		require.False(t, gommitCfg.SpellCheck.Enabled)

		require.False(t, gommitCfg.Signature.Required)
		require.Equal(t, "test", gommitCfg.Signature.Identity.PublicKeyURI)
		require.False(t, gommitCfg.SignOffRequired)

		require.Equal(t, "develop", gommitCfg.Reference)
		require.False(t, gommitCfg.IgnoreMergeCommits)
		require.True(t, gommitCfg.NCommitsAhead)
	})
}
