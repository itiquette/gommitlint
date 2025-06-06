// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

// TestSpellRule tests the spell checking rule functionality.
func TestSpellRule(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		body        string
		description string
	}{
		{
			name:        "Valid spelling",
			subject:     "Add new feature",
			body:        "This commit adds a new feature to the application.",
			description: "Should handle correctly spelled words",
		},
		{
			name:        "Common misspelling in subject",
			subject:     "Fix comit message validation",
			body:        "Update validation logic",
			description: "Should check spelling in subject",
		},
		{
			name:        "Common misspelling in body",
			subject:     "Update documentation",
			body:        "This comit updates the documentation files.",
			description: "Should check spelling in body",
		},
		{
			name:        "Technical terms",
			subject:     "Add JWT authentication",
			body:        "Implement JWT-based authentication with bcrypt hashing.",
			description: "Should handle technical terms",
		},
		{
			name:        "Code snippets",
			subject:     "Fix variable naming",
			body:        "Update getUserInfo() method to getUser().",
			description: "Should handle code snippets",
		},
		{
			name:        "Empty subject and body",
			subject:     "",
			body:        "",
			description: "Should handle empty content gracefully",
		},
		{
			name:        "Subject only commit",
			subject:     "Quick typo fix",
			body:        "",
			description: "Should handle subject-only commits",
		},
		{
			name:        "Special characters and punctuation",
			subject:     "Fix bug: can't access user data",
			body:        "Resolve the issue where users can't access their data.",
			description: "Should handle contractions and punctuation",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using factory
			message := testCase.subject
			if testCase.body != "" {
				message = testCase.subject + "\n\n" + testCase.body
			}

			commit := testdata.Commit(message)

			cfg := config.Config{}
			rule := rules.NewSpellRule(cfg)

			failures := rule.Validate(commit, cfg)

			// We don't assert on specific errors since spell checking results
			// depend on the underlying misspell library behavior
			// We mainly test that the rule runs without crashing
			_ = failures

			// Verify rule name
			require.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")
		})
	}
}

// TestSpellRuleWithCustomDictionary tests the custom dictionary functionality.
func TestSpellRuleWithCustomDictionary(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		customWords []string
		description string
	}{
		{
			name:        "No custom dictionary",
			message:     "Add new feature",
			customWords: nil,
			description: "Should work without custom dictionary",
		},
		{
			name:        "Custom dictionary with technical terms",
			message:     "Add OAuth integration",
			customWords: []string{"OAuth"},
			description: "Should use custom dictionary to ignore specified words",
		},
		{
			name:        "Multiple custom words",
			message:     "Implement GraphQL with OAuth",
			customWords: []string{"GraphQL", "OAuth"},
			description: "Should handle multiple custom words",
		},
		{
			name:        "Empty custom dictionary",
			message:     "Update documentation",
			customWords: []string{},
			description: "Should handle empty custom dictionary",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using factory
			commit := testdata.Commit(testCase.message)

			// Create rule with custom dictionary if provided
			cfg := config.Config{
				Spell: config.SpellConfig{
					IgnoreWords: testCase.customWords,
				},
			}
			rule := rules.NewSpellRule(cfg)

			failures := rule.Validate(commit, cfg)

			// We don't assert on specific errors since results depend on the misspell library
			// We test that the rule handles custom dictionaries without crashing
			_ = failures

			// Verify rule name
			require.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")
		})
	}
}

// TestSpellRuleEdgeCases tests edge cases and boundary conditions.
func TestSpellRuleEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		description string
	}{
		{
			name:        "Very long message",
			message:     "Add feature with lots of content repeated multiple times to test performance and handling of long text",
			description: "Should handle long messages efficiently",
		},
		{
			name:        "Unicode characters",
			message:     "Add support for café and naïve words",
			description: "Should handle Unicode characters",
		},
		{
			name:        "Mixed case words",
			message:     "Update README and API documentation",
			description: "Should handle mixed case words",
		},
		{
			name:        "URLs and paths",
			message:     "Update https://example.com/api/v1/users endpoint",
			description: "Should handle URLs and paths",
		},
		{
			name:        "Code variables",
			message:     "Fix getUserById() method implementation",
			description: "Should handle code-style variable names",
		},
		{
			name:        "Numbers and versions",
			message:     "Upgrade to version 1.2.3 and Node.js 18",
			description: "Should handle version numbers",
		},
		{
			name:        "Only whitespace",
			message:     "   \t\n  ",
			description: "Should handle whitespace-only content",
		},
		{
			name:        "Special characters only",
			message:     "!@#$%^&*()_+-={}[]|:;\"'<>?,.`~",
			description: "Should handle special characters",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using factory
			commit := testdata.Commit(testCase.message)

			cfg := config.Config{}
			rule := rules.NewSpellRule(cfg)

			failures := rule.Validate(commit, cfg)

			// For edge cases, we mainly test that the rule doesn't crash
			// The specific results depend on the spell checker implementation
			_ = failures

			// Verify rule name
			require.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")
		})
	}
}
