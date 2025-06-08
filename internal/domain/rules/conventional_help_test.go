// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
)

// TestConventionalCommitRule_HelpContextAware tests context-aware help generation.
func TestConventionalCommitRule_HelpContextAware(t *testing.T) {
	tests := []struct {
		name             string
		config           config.Config
		expectedContains []string
		description      string
	}{
		{
			name: "default help",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					AllowBreaking:        true,
					MaxDescriptionLength: 72,
				},
			},
			expectedContains: []string{
				"Use conventional commit format",
				"Standard types (any case): feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert",
				"Multi-scope format",
				"Breaking changes",
				"Max description length: 72",
				"Spacing: exactly one space after colon",
			},
			description: "Default configuration help",
		},
		{
			name: "custom types only",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Types: []string{"feature", "bugfix", "docs"},
				},
			},
			expectedContains: []string{
				"Use conventional commit format",
				"Valid types: feature, bugfix, docs",
				"Multi-scope format",
				"Spacing: exactly one space after colon",
			},
			description: "Custom types configuration help",
		},
		{
			name: "custom scopes",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Scopes: []string{"frontend", "backend", "docs"},
				},
			},
			expectedContains: []string{
				"Use conventional commit format",
				"Valid scopes: frontend, backend, docs",
				"Multi-scope format",
				"Standard types (any case)",
			},
			description: "Custom scopes configuration help",
		},
		{
			name: "custom description length",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					MaxDescriptionLength: 50,
				},
			},
			expectedContains: []string{
				"Use conventional commit format",
				"Max description length: 50",
				"Standard types (any case)",
			},
			description: "Custom description length help",
		},
		{
			name: "comprehensive config",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Types:                []string{"feat", "fix"},
					Scopes:               []string{"ui", "api"},
					MaxDescriptionLength: 100,
					AllowBreaking:        true,
				},
			},
			expectedContains: []string{
				"Use conventional commit format",
				"Valid types: feat, fix",
				"Valid scopes: ui, api",
				"Multi-scope format",
				"Breaking changes",
				"Max description length: 100",
				"Spacing: exactly one space after colon",
			},
			description: "Comprehensive configuration help",
		},
		{
			name: "no breaking changes allowed",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					AllowBreaking: false,
				},
			},
			expectedContains: []string{
				"Use conventional commit format",
				"Standard types (any case)",
				"Multi-scope format",
				"Spacing: exactly one space after colon",
			},
			description: "Configuration with breaking changes disabled",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(testCase.config)
			help := rule.Help()

			for _, expected := range testCase.expectedContains {
				require.Contains(t, help, expected,
					"Help text should contain '%s' for: %s\nFull help: %s",
					expected, testCase.description, help)
			}

			// Verify basic structure
			require.Contains(t, help, "type(scope): description",
				"Help should contain basic format description")
		})
	}
}

// TestConventionalCommitRule_HelpFormatting tests the structure and formatting of help text.
func TestConventionalCommitRule_HelpFormatting(t *testing.T) {
	rule := rules.NewConventionalCommitRule(config.Config{})
	help := rule.Help()

	// Should be multiple lines
	lines := strings.Split(help, "\n")
	require.Greater(t, len(lines), 3, "Help should contain multiple lines")

	// First line should be the main format
	require.Contains(t, lines[0], "Use conventional commit format")

	// Should mention spacing requirement
	require.Contains(t, help, "exactly one space after colon")

	// Should not be too verbose (reasonable length)
	require.Less(t, len(help), 1000, "Help should be reasonably concise")
}

// TestConventionalCommitRule_HelpWithExplicitConfig tests help text with various explicit configurations.
func TestConventionalCommitRule_HelpWithExplicitConfig(t *testing.T) {
	tests := []struct {
		name             string
		config           config.Config
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "only custom types configured",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Types: []string{"custom", "special"},
				},
			},
			shouldContain: []string{
				"Valid types: custom, special",
			},
			shouldNotContain: []string{
				"Standard types",
				"feat, fix, docs", // Should not show standard types
			},
		},
		{
			name: "scopes but no types",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					Scopes: []string{"api", "ui"},
				},
			},
			shouldContain: []string{
				"Valid scopes: api, ui",
				"Standard types (any case)",
			},
			shouldNotContain: []string{
				"Valid types:",
			},
		},
		{
			name: "breaking changes disabled",
			config: config.Config{
				Conventional: config.ConventionalConfig{
					AllowBreaking: false,
				},
			},
			shouldContain: []string{
				"Multi-scope format",
				"Standard types",
			},
			shouldNotContain: []string{
				"Breaking changes:",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := rules.NewConventionalCommitRule(testCase.config)
			help := rule.Help()

			for _, should := range testCase.shouldContain {
				require.Contains(t, help, should,
					"Help should contain '%s'\nFull help: %s", should, help)
			}

			for _, shouldNot := range testCase.shouldNotContain {
				require.NotContains(t, help, shouldNot,
					"Help should not contain '%s'\nFull help: %s", shouldNot, help)
			}
		})
	}
}
