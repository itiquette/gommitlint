// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	// Test with no options
	config := NewConfig()
	require.Equal(t, DefaultConfig(), config)

	// Test with a single option
	config = NewConfig(WithSubjectMaxLength(50))
	require.Equal(t, 50, config.Subject.MaxLength)

	// Test with multiple options
	config = NewConfig(
		WithSubjectMaxLength(50),
		WithBodyRequired(true),
		WithConventionalRequired(false),
	)
	require.Equal(t, 50, config.Subject.MaxLength)
	require.True(t, config.Body.Required)
	require.False(t, config.Conventional.Required)
}

func TestSubjectOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithSubject",
			option: WithSubject(SubjectConfig{MaxLength: 50}),
			expected: func(config Config) bool {
				return config.Subject.MaxLength == 50
			},
		},
		{
			name:   "WithSubjectCase",
			option: WithSubjectCase("upper"),
			expected: func(config Config) bool {
				return config.Subject.Case == "upper"
			},
		},
		{
			name:   "WithSubjectMaxLength",
			option: WithSubjectMaxLength(50),
			expected: func(config Config) bool {
				return config.Subject.MaxLength == 50
			},
		},
		{
			name:   "WithSubjectImperative",
			option: WithSubjectImperative(false),
			expected: func(config Config) bool {
				return !config.Subject.Imperative
			},
		},
		{
			name:   "WithSubjectInvalidSuffixes",
			option: WithSubjectInvalidSuffixes(".,:"),
			expected: func(config Config) bool {
				return config.Subject.InvalidSuffixes == ".,:"
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestBodyOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithBody",
			option: WithBody(BodyConfig{Required: true}),
			expected: func(config Config) bool {
				return config.Body.Required
			},
		},
		{
			name:   "WithBodyRequired",
			option: WithBodyRequired(true),
			expected: func(config Config) bool {
				return config.Body.Required
			},
		},
		{
			name:   "WithBodyAllowSignOffOnly",
			option: WithBodyAllowSignOffOnly(true),
			expected: func(config Config) bool {
				return config.Body.AllowSignOffOnly
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestConventionalOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithConventional",
			option: WithConventional(ConventionalConfig{Required: false}),
			expected: func(config Config) bool {
				return !config.Conventional.Required
			},
		},
		{
			name:   "WithConventionalRequired",
			option: WithConventionalRequired(false),
			expected: func(config Config) bool {
				return !config.Conventional.Required
			},
		},
		{
			name:   "WithConventionalTypes",
			option: WithConventionalTypes([]string{"custom"}),
			expected: func(config Config) bool {
				return len(config.Conventional.Types) == 1 && config.Conventional.Types[0] == "custom"
			},
		},
		{
			name:   "WithConventionalScopes",
			option: WithConventionalScopes([]string{"api", "ui"}),
			expected: func(config Config) bool {
				return len(config.Conventional.Scopes) == 2 && config.Conventional.Scopes[0] == "api"
			},
		},
		{
			name:   "WithConventionalMaxDescriptionLength",
			option: WithConventionalMaxDescriptionLength(50),
			expected: func(config Config) bool {
				return config.Conventional.MaxDescriptionLength == 50
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestSpellCheckOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithSpellCheck",
			option: WithSpellCheck(SpellCheckConfig{Enabled: false}),
			expected: func(config Config) bool {
				return !config.SpellCheck.Enabled
			},
		},
		{
			name:   "WithSpellCheckEnabled",
			option: WithSpellCheckEnabled(false),
			expected: func(config Config) bool {
				return !config.SpellCheck.Enabled
			},
		},
		{
			name:   "WithSpellCheckLocale",
			option: WithSpellCheckLocale("US"),
			expected: func(config Config) bool {
				return config.SpellCheck.Locale == "US"
			},
		},
		{
			name:   "WithSpellCheckIgnoreWords",
			option: WithSpellCheckIgnoreWords([]string{"foo", "bar"}),
			expected: func(config Config) bool {
				return len(config.SpellCheck.IgnoreWords) == 2 && config.SpellCheck.IgnoreWords[0] == "foo"
			},
		},
		{
			name:   "WithSpellCheckCustomWords",
			option: WithSpellCheckCustomWords(map[string]string{"foo": "bar"}),
			expected: func(config Config) bool {
				return len(config.SpellCheck.CustomWords) == 1 && config.SpellCheck.CustomWords["foo"] == "bar"
			},
		},
		{
			name:   "WithSpellCheckMaxErrors",
			option: WithSpellCheckMaxErrors(5),
			expected: func(config Config) bool {
				return config.SpellCheck.MaxErrors == 5
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestSecurityOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithSecurity",
			option: WithSecurity(SecurityConfig{SignatureRequired: false}),
			expected: func(config Config) bool {
				return !config.Security.SignatureRequired
			},
		},
		{
			name:   "WithSignatureRequired",
			option: WithSignatureRequired(false),
			expected: func(config Config) bool {
				return !config.Security.SignatureRequired
			},
		},
		{
			name:   "WithSignOffRequired",
			option: WithSignOffRequired(false),
			expected: func(config Config) bool {
				return !config.Security.SignOffRequired
			},
		},
		{
			name:   "WithAllowedSignatureTypes",
			option: WithAllowedSignatureTypes([]string{"gpg"}),
			expected: func(config Config) bool {
				return len(config.Security.AllowedSignatureTypes) == 1 && config.Security.AllowedSignatureTypes[0] == "gpg"
			},
		},
		{
			name:   "WithAllowMultipleSignOffs",
			option: WithAllowMultipleSignOffs(false),
			expected: func(config Config) bool {
				return !config.Security.AllowMultipleSignOffs
			},
		},
		{
			name:   "WithIdentity",
			option: WithIdentity(IdentityConfig{PublicKeyURI: "test"}),
			expected: func(config Config) bool {
				return config.Security.Identity.PublicKeyURI == "test"
			},
		},
		{
			name:   "WithPublicKeyURI",
			option: WithPublicKeyURI("test"),
			expected: func(config Config) bool {
				return config.Security.Identity.PublicKeyURI == "test"
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestRepositoryOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithRepository",
			option: WithRepository(RepositoryConfig{Reference: "develop"}),
			expected: func(config Config) bool {
				return config.Repository.Reference == "develop"
			},
		},
		{
			name:   "WithReference",
			option: WithReference("develop"),
			expected: func(config Config) bool {
				return config.Repository.Reference == "develop"
			},
		},
		{
			name:   "WithIgnoreMergeCommits",
			option: WithIgnoreMergeCommits(false),
			expected: func(config Config) bool {
				return !config.Repository.IgnoreMergeCommits
			},
		},
		{
			name:   "WithMaxCommitsAhead",
			option: WithMaxCommitsAhead(10),
			expected: func(config Config) bool {
				return config.Repository.MaxCommitsAhead == 10
			},
		},
		{
			name:   "WithCheckCommitsAhead",
			option: WithCheckCommitsAhead(false),
			expected: func(config Config) bool {
				return !config.Repository.CheckCommitsAhead
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestRulesOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithRules",
			option: WithRules(RulesConfig{EnabledRules: []string{"rule1"}}),
			expected: func(config Config) bool {
				return len(config.Rules.EnabledRules) == 1 && config.Rules.EnabledRules[0] == "rule1"
			},
		},
		{
			name:   "WithEnabledRules",
			option: WithEnabledRules([]string{"rule1", "rule2"}),
			expected: func(config Config) bool {
				return len(config.Rules.EnabledRules) == 2 && config.Rules.EnabledRules[0] == "rule1"
			},
		},
		{
			name:   "WithDisabledRules",
			option: WithDisabledRules([]string{"rule1"}),
			expected: func(config Config) bool {
				return len(config.Rules.DisabledRules) == 1 && config.Rules.DisabledRules[0] == "rule1"
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestJiraOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(config Config) bool
	}{
		{
			name:   "WithJira",
			option: WithJira(JiraConfig{Required: true}),
			expected: func(config Config) bool {
				return config.Subject.Jira.Required
			},
		},
		{
			name:   "WithJiraRequired",
			option: WithJiraRequired(true),
			expected: func(config Config) bool {
				return config.Subject.Jira.Required
			},
		},
		{
			name:   "WithJiraProjects",
			option: WithJiraProjects([]string{"PROJ"}),
			expected: func(config Config) bool {
				return len(config.Subject.Jira.Projects) == 1 && config.Subject.Jira.Projects[0] == "PROJ"
			},
		},
		{
			name:   "WithJiraPattern",
			option: WithJiraPattern("custom"),
			expected: func(config Config) bool {
				return config.Subject.Jira.Pattern == "custom"
			},
		},
		{
			name:   "WithJiraBodyRef",
			option: WithJiraBodyRef(true),
			expected: func(config Config) bool {
				return config.Subject.Jira.BodyRef
			},
		},
		{
			name:   "WithJiraStrict",
			option: WithJiraStrict(true),
			expected: func(config Config) bool {
				return config.Subject.Jira.Strict
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config := NewConfig(testCase.option)
			require.True(t, testCase.expected(config), "Option did not set the expected value")
		})
	}
}

func TestOptionChaining(t *testing.T) {
	// Test with a chain of options
	config := NewConfig(
		WithSubjectMaxLength(50),
		WithBodyRequired(true),
		WithConventionalRequired(false),
		WithSpellCheckEnabled(false),
		WithSignatureRequired(false),
		WithReference("develop"),
		WithJiraRequired(true),
	)

	// Verify all options were applied
	require.Equal(t, 50, config.Subject.MaxLength)
	require.True(t, config.Body.Required)
	require.False(t, config.Conventional.Required)
	require.False(t, config.SpellCheck.Enabled)
	require.False(t, config.Security.SignatureRequired)
	require.Equal(t, "develop", config.Repository.Reference)
	require.True(t, config.Subject.Jira.Required)
}
