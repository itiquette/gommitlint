// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	// Test creating a default config
	config := NewConfig()
	require.Equal(t, 100, config.Subject.MaxLength)
	require.Equal(t, "lower", config.Subject.Case)
	require.True(t, config.Subject.Imperative)

	// Test modifying the config with method chaining
	modifiedConfig := config.
		WithSubjectMaxLength(50).
		WithSubjectCase("upper").
		WithSubjectImperative(false).
		WithBodyRequired(true).
		WithConventionalRequired(true).
		WithSpellEnabled(false)

	// Original config should remain unchanged
	require.Equal(t, 100, config.Subject.MaxLength)
	require.Equal(t, "lower", config.Subject.Case)
	require.True(t, config.Subject.Imperative)

	// Modified config should have the new values
	require.Equal(t, 50, modifiedConfig.Subject.MaxLength)
	require.Equal(t, "upper", modifiedConfig.Subject.Case)
	require.False(t, modifiedConfig.Subject.Imperative)
	require.True(t, modifiedConfig.Body.Required)
	require.True(t, modifiedConfig.Conventional.Required)
	require.False(t, modifiedConfig.SpellCheck.Enabled)
}

func TestConfigMethodChaining(t *testing.T) {
	// Test with method chaining
	config := NewConfig().
		WithSubjectMaxLength(50).
		WithBodyRequired(true).
		WithConventionalRequired(false).
		WithSpellEnabled(false).
		WithSignatureRequired(false).
		WithReferenceBranch("develop").
		WithJiraRequired(true)

	// Verify all methods were applied
	require.Equal(t, 50, config.Subject.MaxLength)
	require.True(t, config.Body.Required)
	require.False(t, config.Conventional.Required)
	require.False(t, config.SpellCheck.Enabled)
	require.False(t, config.Security.SignatureRequired)
	require.Equal(t, "develop", config.Repository.Reference)
	require.True(t, config.Subject.Jira.Required)
}

func TestConfigValidationMethods(t *testing.T) {
	config := NewConfig()

	// Test subject config methods
	require.Equal(t, config.Subject.MaxLength, config.SubjectMaxLength())
	require.Equal(t, config.Subject.Case, config.SubjectCase())
	require.Equal(t, config.Subject.Imperative, config.SubjectRequireImperative())

	// Test body config methods
	require.Equal(t, config.Body.Required, config.BodyRequired())
	require.Equal(t, config.Body.AllowSignOffOnly, config.BodyAllowSignOffOnly())

	// Test conventional config methods
	require.Equal(t, config.Conventional.Required, config.ConventionalRequired())
	require.Equal(t, config.Conventional.Types, config.ConventionalTypes())
	require.Equal(t, config.Conventional.Scopes, config.ConventionalScopes())

	// Test jira config methods
	require.Equal(t, config.Subject.Jira.Required, config.JiraRequired())
	require.Equal(t, config.Subject.Jira.Projects, config.JiraProjects())

	// Test security config methods
	require.Equal(t, config.Security.SignatureRequired, config.SignatureRequired())
	require.Equal(t, config.Security.SignOffRequired, config.SignOffRequired())

	// Test repository config methods
	require.Equal(t, config.Repository.Reference, config.ReferenceBranch())
	require.Equal(t, config.Repository.IgnoreMergeCommits, config.IgnoreMergeCommits())
}

func TestConfigIsValid(t *testing.T) {
	// Valid config
	validConfig := NewConfig().
		WithSubjectMaxLength(50).
		WithBodyRequired(true)

	require.True(t, validConfig.IsValid())

	// Invalid config with negative subject max length
	invalidConfig := NewConfig().
		WithSubjectMaxLength(-10)

	require.False(t, invalidConfig.IsValid())

	// Validate should return errors for invalid configs
	errors := invalidConfig.Validate()
	require.NotEmpty(t, errors)
}
