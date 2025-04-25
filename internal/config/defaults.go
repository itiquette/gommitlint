// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

// DefaultConfig creates a new configuration with default values.
// All configuration values are set to their default values.
func DefaultConfig() Config {
	return Config{
		Subject:      DefaultSubjectConfig(),
		Body:         DefaultBodyConfig(),
		Conventional: DefaultConventionalConfig(),
		SpellCheck:   DefaultSpellCheckConfig(),
		Security:     DefaultSecurityConfig(),
		Repository:   DefaultRepositoryConfig(),
		Rules:        DefaultRulesConfig(),
	}
}

// DefaultSubjectConfig creates a new subject configuration with default values.
func DefaultSubjectConfig() SubjectConfig {
	return SubjectConfig{
		Case:            "lower",
		Imperative:      true,
		InvalidSuffixes: ".",
		MaxLength:       100,
		Jira:            DefaultJiraConfig(),
	}
}

// DefaultBodyConfig creates a new body configuration with default values.
func DefaultBodyConfig() BodyConfig {
	return BodyConfig{
		Required:         false,
		AllowSignOffOnly: false,
	}
}

// DefaultConventionalConfig creates a new conventional commit configuration with default values.
func DefaultConventionalConfig() ConventionalConfig {
	return ConventionalConfig{
		MaxDescriptionLength: 72,
		Scopes:               []string{},
		Types: []string{
			"feat", "fix", "docs", "style", "refactor",
			"perf", "test", "build", "ci", "chore", "revert",
		},
		Required: true,
	}
}

// DefaultSpellCheckConfig creates a new spell check configuration with default values.
func DefaultSpellCheckConfig() SpellCheckConfig {
	return SpellCheckConfig{
		Locale:      "UK",
		Enabled:     true,
		IgnoreWords: []string{},
		CustomWords: map[string]string{},
		MaxErrors:   0,
	}
}

// DefaultSecurityConfig creates a new security configuration with default values.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		SignatureRequired:     true,
		AllowedSignatureTypes: []string{"gpg", "ssh"},
		SignOffRequired:       true,
		AllowMultipleSignOffs: true,
		Identity:              DefaultIdentityConfig(),
	}
}

// DefaultIdentityConfig creates a new identity configuration with default values.
func DefaultIdentityConfig() IdentityConfig {
	return IdentityConfig{
		PublicKeyURI: "",
	}
}

// DefaultRepositoryConfig creates a new repository configuration with default values.
func DefaultRepositoryConfig() RepositoryConfig {
	return RepositoryConfig{
		Reference:          "main",
		IgnoreMergeCommits: true,
		MaxCommitsAhead:    5,
		CheckCommitsAhead:  true,
	}
}

// DefaultRulesConfig creates a new rules configuration with default values.
func DefaultRulesConfig() RulesConfig {
	return RulesConfig{
		EnabledRules:  []string{},
		DisabledRules: []string{},
	}
}

// DefaultJiraConfig creates a new Jira configuration with default values.
func DefaultJiraConfig() JiraConfig {
	return JiraConfig{
		Projects: []string{},
		Required: false,
		BodyRef:  false,
		Pattern:  "[A-Z]+-\\d+",
		Strict:   false,
	}
}
