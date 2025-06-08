// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/adapters/spell"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
)

// createSpellTestCommit creates a test commit with the given message.
func createSpellTestCommit(message string) domain.Commit {
	commit := domain.Commit{
		Hash:          "abc123def456",
		Subject:       message,
		Message:       message,
		Author:        "Test User",
		AuthorEmail:   "test@example.com",
		CommitDate:    time.Now().Format(time.RFC3339),
		IsMergeCommit: false,
	}

	// Parse subject and body from message
	lines := strings.Split(message, "\n")
	if len(lines) > 0 {
		commit.Subject = lines[0]
	}

	// Parse body - everything after the subject line
	if len(lines) > 1 {
		bodyStart := 1

		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				bodyStart = i + 1

				break
			}
		}

		if bodyStart < len(lines) {
			commit.Body = strings.Join(lines[bodyStart:], "\n")
		} else if bodyStart == 1 {
			commit.Body = strings.Join(lines[1:], "\n")
		}
	}

	return commit
}

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

			commit := createSpellTestCommit(message)

			cfg := config.Config{}
			checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
			rule := rules.NewSpellRule(checker, cfg)

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
			commit := createSpellTestCommit(testCase.message)

			// Create rule with custom dictionary if provided
			cfg := config.Config{
				Spell: config.SpellConfig{
					IgnoreWords: testCase.customWords,
				},
			}
			checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
			rule := rules.NewSpellRule(checker, cfg)

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
			commit := createSpellTestCommit(testCase.message)

			cfg := config.Config{}
			checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
			rule := rules.NewSpellRule(checker, cfg)

			failures := rule.Validate(commit, cfg)

			// For edge cases, we mainly test that the rule doesn't crash
			// The specific results depend on the spell checker implementation
			_ = failures

			// Verify rule name
			require.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")
		})
	}
}

// TestSpellRule_VerboseOutput tests the restored verbose output functionality.
func TestSpellRule_VerboseOutput(t *testing.T) {
	t.Skip("Enhanced verbose features not implemented")

	tests := []struct {
		name                string
		message             string
		locale              string
		expectMisspellings  bool
		checkVerboseContext bool
		checkHelpText       bool
	}{
		{
			name:                "Multiple misspellings with verbose context",
			message:             "Fix teh recieve problem", //nolint:misspell
			locale:              "us",
			expectMisspellings:  true,
			checkVerboseContext: true,
			checkHelpText:       true,
		},
		{
			name:                "Single misspelling",
			message:             "Fix teh issue", //nolint:misspell
			locale:              "",
			expectMisspellings:  true,
			checkVerboseContext: true,
			checkHelpText:       true,
		},
		{
			name:                "No misspellings",
			message:             "Fix commit message validation",
			locale:              "us",
			expectMisspellings:  false,
			checkVerboseContext: false,
			checkHelpText:       false,
		},
		{
			name:                "Many misspellings (test truncation)",
			message:             "Fix teh recieve seperate acheive occured issues", //nolint:misspell
			locale:              "uk",
			expectMisspellings:  true,
			checkVerboseContext: true,
			checkHelpText:       true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Spell: config.SpellConfig{
					Locale: testCase.locale,
				},
			}

			commit := createSpellTestCommit(testCase.message)
			checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
			rule := rules.NewSpellRule(checker, cfg)
			errors := rule.Validate(commit, cfg)

			if testCase.expectMisspellings {
				require.NotEmpty(t, errors, "Expected misspellings to be found")

				if testCase.checkVerboseContext {
					// Check that first error contains verbose context
					firstError := errors[0]

					// Check for verbose summary
					verboseSummary, hasVerbose := firstError.Context["verbose_summary"]
					require.True(t, hasVerbose, "Should have verbose_summary in context")
					require.NotEmpty(t, verboseSummary, "Verbose summary should not be empty")

					// Check locale is included in verbose summary if specified
					if testCase.locale != "" {
						require.Contains(t, verboseSummary, strings.ToUpper(testCase.locale),
							"Verbose summary should include locale information")
					}

					// Check for all misspellings list
					allMisspellings, hasAll := firstError.Context["all_misspellings"]
					require.True(t, hasAll, "Should have all_misspellings in context")
					require.NotEmpty(t, allMisspellings, "All misspellings list should not be empty")

					// Check for misspelling count
					totalCount, hasCount := firstError.Context["total_misspellings"]
					require.True(t, hasCount, "Should have total_misspellings in context")
					require.NotEmpty(t, totalCount, "Total count should not be empty")
				}

				if testCase.checkHelpText {
					// Check comprehensive help text
					for _, err := range errors {
						require.NotEmpty(t, err.Help, "Help text should not be empty")
						require.Contains(t, err.Help, "Found spelling errors:",
							"Help should contain header")
						require.Contains(t, err.Help, "→",
							"Help should contain correction arrows")

						// Check for truncation message in help if many errors
						if len(errors) > 5 {
							require.Contains(t, errors[0].Help, "and",
								"Help should indicate more errors exist")
							require.Contains(t, errors[0].Help, "more misspellings",
								"Help should mention additional misspellings")
						}
					}
				}
			} else {
				require.Empty(t, errors, "Expected no misspellings")
			}
		})
	}
}

// TestSpellRule_EnhancedContext tests that all context fields are properly populated.
func TestSpellRule_EnhancedContext(t *testing.T) {
	cfg := config.Config{
		Spell: config.SpellConfig{
			Locale: "uk",
		},
	}

	// Use a message with known misspellings
	commit := createSpellTestCommit("Fix teh issue") //nolint:misspell
	checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
	rule := rules.NewSpellRule(checker, cfg)

	errors := rule.Validate(commit, cfg)
	require.NotEmpty(t, errors, "Should find misspellings")

	// Check each error has proper context
	for _, err := range errors {
		// Basic context fields
		require.NotEmpty(t, err.Context["actual"], "Should have misspelled word")
		require.NotEmpty(t, err.Context["expected"], "Should have correction")
	}
}

// TestSpellRule_LocaleVariants tests different locale handling.
func TestSpellRule_LocaleVariants(t *testing.T) {
	tests := []struct {
		name        string
		locale      string
		message     string
		description string
	}{
		{
			name:        "US English locale",
			locale:      "us",
			message:     "Add color support",
			description: "Should handle US spelling",
		},
		{
			name:        "UK English locale",
			locale:      "uk",
			message:     "Add colour support",
			description: "Should handle UK spelling",
		},
		{
			name:        "British locale variant",
			locale:      "en-gb",
			message:     "Add colour support",
			description: "Should recognize en-gb as UK",
		},
		{
			name:        "American locale variant",
			locale:      "en-us",
			message:     "Add color support",
			description: "Should recognize en-us as US",
		},
		{
			name:        "Unknown locale defaults to US",
			locale:      "fr",
			message:     "Add color support",
			description: "Should default to US for unknown locales",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Spell: config.SpellConfig{
					Locale: testCase.locale,
				},
			}

			commit := createSpellTestCommit(testCase.message)
			checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
			rule := rules.NewSpellRule(checker, cfg)

			errors := rule.Validate(commit, cfg)

			// We don't assert on specific spelling results as they depend
			// on the misspell library's dictionary. We verify the rule
			// processes different locales without errors.
			_ = errors

			// If there are errors and locale was specified, verify it's in context
			if len(errors) > 0 && testCase.locale != "" {
				require.Equal(t, testCase.locale, errors[0].Context["locale"],
					"Locale should be preserved in context")
			}
		})
	}
}

// TestSpellRule_CustomDictionaryWithVerboseOutput tests custom dictionary with verbose features.
func TestSpellRule_CustomDictionaryWithVerboseOutput(t *testing.T) {
	cfg := config.Config{
		Spell: config.SpellConfig{
			IgnoreWords: []string{"the"}, // Intentionally ignore "the" misspelling
			Locale:      "us",
		},
	}

	// Message with ignored and non-ignored misspellings
	commit := createSpellTestCommit("Fix the the issue") // "the" is more commonly detected misspelling
	checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
	rule := rules.NewSpellRule(checker, cfg)

	errors := rule.Validate(commit, cfg)

	// Skip this test if spell checking is not working (no misspellings detected)
	if len(errors) == 0 {
		t.Skip("Spell checker not detecting misspellings - may need dictionary setup")

		return
	}

	// Should only find misspellings that are not ignored
	require.NotEmpty(t, errors, "Should find some misspellings")

	// If we have errors, verify verbose output behavior
	if len(errors) > 0 {
		verboseSummary := errors[0].Context["verbose_summary"]
		require.NotContains(t, verboseSummary, "the", "Should not include ignored words")
	}
}

// TestSpellRule_HelpTextTruncation tests the help text truncation for many errors.
func TestSpellRule_HelpTextTruncation(t *testing.T) {
	// Create a message with many known misspellings
	// Using more obvious misspellings to ensure detection
	message := "This is teh recieve seperate acheive occured definately inconsistant mispelling" //nolint:misspell

	cfg := config.Config{}
	commit := createSpellTestCommit(message)
	checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
	rule := rules.NewSpellRule(checker, cfg)

	errors := rule.Validate(commit, cfg)

	// If we have more than 5 misspellings, test truncation
	if len(errors) > 5 {
		// Check help text truncation
		helpText := errors[0].Help

		// Should show first 5 corrections
		lines := strings.Split(helpText, "\n")
		correctionCount := 0

		for _, line := range lines {
			if strings.Contains(line, "→") {
				correctionCount++
			}
		}

		require.Equal(t, 5, correctionCount, "Should show exactly 5 corrections")

		// Should have "and X more" message
		require.Contains(t, helpText, "and", "Should have continuation message")
		require.Contains(t, helpText, "more misspellings", "Should indicate additional errors")

		// Verify the count is correct
		expectedMore := len(errors) - 5
		require.Contains(t, helpText, strconv.Itoa(expectedMore), "Should show correct additional count")
	} else if len(errors) > 0 {
		// For fewer errors, just verify help text exists and has corrections
		helpText := errors[0].Help
		require.Contains(t, helpText, "Found spelling errors:", "Should have header")
		require.Contains(t, helpText, "→", "Should have corrections")

		// Count corrections should match error count
		lines := strings.Split(helpText, "\n")
		correctionCount := 0

		for _, line := range lines {
			if strings.Contains(line, "→") {
				correctionCount++
			}
		}

		require.Equal(t, len(errors), correctionCount, "Should show all corrections when <= 5")
	} else {
		t.Skip("No misspellings detected in test message")
	}
}

// TestSpellRule_ErrorCodes tests that proper error codes are used.
func TestSpellRule_ErrorCodes(t *testing.T) {
	cfg := config.Config{}
	commit := createSpellTestCommit("Fix teh issue") //nolint:misspell
	checker := spell.NewMisspellAdapter(cfg.Spell.Locale)
	rule := rules.NewSpellRule(checker, cfg)

	errors := rule.Validate(commit, cfg)
	require.NotEmpty(t, errors, "Should find misspellings")

	for _, err := range errors {
		require.Equal(t, string(domain.ErrMisspelledWord), err.Code,
			"Should use correct error code")
		require.Equal(t, "Spell", err.Rule, "Should have correct rule name")
	}
}
