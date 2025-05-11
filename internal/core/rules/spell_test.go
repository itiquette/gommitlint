// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// TestSpellRule tests the configuration aspects of the spell rule without performing actual spell checking.
// This is a simplified version that doesn't require a real spell checker.
func TestSpellRule(t *testing.T) {
	// Create a test commit
	commit := domain.CommitInfo{
		Subject: "Test commit message",
		Message: "Test commit message with some words",
	}

	// Test that options are correctly applied
	t.Run("Rule options", func(t *testing.T) {
		// Create rule with options
		rule := rules.NewSpellRule(
			rules.WithIgnoreCase(true),
			rules.WithIgnoreWords([]string{"test", "words"}),
			rules.WithLocale("en_US"),
		)

		// Create a mock context that will make spell checking inactive
		ctx := rules.WithSpellCheckConfig(context.Background(), rules.SpellCheckConfig{
			Enabled: false,
		})
		errors := rule.Validate(ctx, commit)

		// With spell checking disabled, we should get no errors
		require.Empty(t, errors, "Expected no errors when spell checking is disabled")

		// Check that name is set correctly
		require.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")
	})

	// Test with errors (preconfigured)
	t.Run("With preconfigured errors", func(t *testing.T) {
		// Create a rule with some predefined misspellings using SetErrors
		rule := rules.NewSpellRule()

		// Create some sample validation errors
		sampleErrors := []appErrors.ValidationError{
			appErrors.CreateBasicError(
				"Spell",
				appErrors.ErrMisspelledWord,
				"misspelled word: 'tset'",
			).WithContext("word", "tset").WithContext("suggestions", "test"),
		}

		// Set errors on the rule
		rule = rule.SetErrors(sampleErrors)

		// Check errors
		require.NotEmpty(t, rule.Errors(), "Rule should have errors")
		require.Len(t, rule.Errors(), 1, "Rule should have 1 error")
		require.Equal(t, "misspelled word: 'tset'", rule.Errors()[0].Message, "Error message should match")

		// Test helper methods
		require.True(t, rule.HasErrors(), "HasErrors should return true")
		require.Contains(t, rule.Result(rule.Errors()), "misspelled", "Result should mention misspelling")
		require.Contains(t, rule.VerboseResult(rule.Errors()), "tset", "VerboseResult should contain the misspelled word")
		require.NotEmpty(t, rule.Help(rule.Errors()), "Help should provide guidance")
	})
}

// TestSpellRuleWithConfig tests the integration with configuration.
func TestSpellRuleWithConfig(t *testing.T) {
	// This test checks that configuration parameters are correctly processed,
	// but doesn't perform actual spell checking
	// Create a test configuration context using WithSpellCheckConfig
	ctx := context.Background()
	ctx = rules.WithSpellCheckConfig(ctx, rules.SpellCheckConfig{
		Enabled:          false, // Disable spell checking for the test
		Language:         "en_US",
		CustomDictionary: []string{"customword", "anotherword"},
		IgnoreCase:       true,
	})

	// Create a rule
	rule := rules.NewSpellRule()

	// Create a test commit
	commit := domain.CommitInfo{
		Subject: "Test commit message",
		Message: "Test commit with customword and anotherword",
	}

	// Validate
	errors := rule.Validate(ctx, commit)

	// With spell checking disabled, we should get no errors
	require.Empty(t, errors, "Expected no errors when spell checking is disabled")
}
