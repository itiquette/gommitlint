// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test default subject config
	require.Equal(t, "lower", config.Subject.Case)
	require.True(t, config.Subject.Imperative)
	require.Equal(t, ".", config.Subject.InvalidSuffixes)
	require.Equal(t, 100, config.Subject.MaxLength)

	// Test default body config
	require.False(t, config.Body.Required)
	require.False(t, config.Body.AllowSignOffOnly)

	// Test default conventional config
	require.Equal(t, 72, config.Conventional.MaxDescriptionLength)
	require.Empty(t, config.Conventional.Scopes)
	require.Contains(t, config.Conventional.Types, "feat")
	require.Contains(t, config.Conventional.Types, "fix")
	require.True(t, config.Conventional.Required)

	// Test default spell check config
	require.Equal(t, "UK", config.SpellCheck.Locale)
	require.True(t, config.SpellCheck.Enabled)
	require.Empty(t, config.SpellCheck.IgnoreWords)
	require.Empty(t, config.SpellCheck.CustomWords)
	require.Equal(t, 0, config.SpellCheck.MaxErrors)

	// Test default security config
	require.True(t, config.Security.SignatureRequired)
	require.Contains(t, config.Security.AllowedSignatureTypes, "gpg")
	require.Contains(t, config.Security.AllowedSignatureTypes, "ssh")
	require.True(t, config.Security.SignOffRequired)
	require.True(t, config.Security.AllowMultipleSignOffs)
	require.Equal(t, "", config.Security.Identity.PublicKeyURI)

	// Test default repository config
	require.Equal(t, "main", config.Repository.Reference)
	require.True(t, config.Repository.IgnoreMergeCommits)
	require.Equal(t, 5, config.Repository.MaxCommitsAhead)
	require.True(t, config.Repository.CheckCommitsAhead)

	// Test default rules config
	require.Empty(t, config.Rules.EnabledRules)
	require.Empty(t, config.Rules.DisabledRules)

	// Test default jira config
	require.Empty(t, config.Subject.Jira.Projects)
	require.False(t, config.Subject.Jira.Required)
	require.False(t, config.Subject.Jira.BodyRef)
	require.Equal(t, "[A-Z]+-\\d+", config.Subject.Jira.Pattern)
	require.False(t, config.Subject.Jira.Strict)
}

func TestDefaultSubjectConfig(t *testing.T) {
	config := DefaultSubjectConfig()
	require.Equal(t, "lower", config.Case)
	require.True(t, config.Imperative)
	require.Equal(t, ".", config.InvalidSuffixes)
	require.Equal(t, 100, config.MaxLength)
}

func TestDefaultBodyConfig(t *testing.T) {
	config := DefaultBodyConfig()
	require.False(t, config.Required)
	require.False(t, config.AllowSignOffOnly)
}

func TestDefaultConventionalConfig(t *testing.T) {
	config := DefaultConventionalConfig()
	require.Equal(t, 72, config.MaxDescriptionLength)
	require.Empty(t, config.Scopes)
	require.Len(t, config.Types, 11)
	require.True(t, config.Required)
}

func TestDefaultSpellCheckConfig(t *testing.T) {
	config := DefaultSpellCheckConfig()
	require.Equal(t, "UK", config.Locale)
	require.True(t, config.Enabled)
	require.Empty(t, config.IgnoreWords)
	require.Empty(t, config.CustomWords)
	require.Equal(t, 0, config.MaxErrors)
}

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()
	require.True(t, config.SignatureRequired)
	require.Len(t, config.AllowedSignatureTypes, 2)
	require.True(t, config.SignOffRequired)
	require.True(t, config.AllowMultipleSignOffs)
}

func TestDefaultIdentityConfig(t *testing.T) {
	config := DefaultIdentityConfig()
	require.Equal(t, "", config.PublicKeyURI)
}

func TestDefaultRepositoryConfig(t *testing.T) {
	config := DefaultRepositoryConfig()
	require.Equal(t, "main", config.Reference)
	require.True(t, config.IgnoreMergeCommits)
	require.Equal(t, 5, config.MaxCommitsAhead)
	require.True(t, config.CheckCommitsAhead)
}

func TestDefaultRulesConfig(t *testing.T) {
	config := DefaultRulesConfig()
	require.Empty(t, config.EnabledRules)
	require.Empty(t, config.DisabledRules)
}

func TestDefaultJiraConfig(t *testing.T) {
	config := DefaultJiraConfig()
	require.Empty(t, config.Projects)
	require.False(t, config.Required)
	require.False(t, config.BodyRef)
	require.Equal(t, "[A-Z]+-\\d+", config.Pattern)
	require.False(t, config.Strict)
}
