// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

	infraConfig "github.com/itiquette/gommitlint/internal/adapters/outgoing/config"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// createImperativeBaseTestContext creates a new context for testing.
// This is the only place in this test file where context.Background() should be called.
func createImperativeBaseTestContext() context.Context {
	return context.Background()
}

// createImperativeTestContext creates a context with test configuration.
func createImperativeTestContext() context.Context {
	// Create a base config to adapt
	cfg := config.NewDefaultConfig()
	cfg.Subject.RequireImperative = true
	cfg.Subject.MaxLength = 72
	cfg.Subject.Case = "sentence"
	cfg.Body.Required = false
	cfg.Conventional.Required = false
	cfg.Conventional.Types = []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}

	// Add the test config to the context using direct adapter pattern
	ctx := createImperativeBaseTestContext()
	adapter := infraConfig.NewAdapter(cfg)

	return contextx.WithConfig(ctx, adapter)
}

func TestImperativeVerbRule(t *testing.T) {
	// Modified tests to use specific test words
	tests := []struct {
		name           string
		message        string
		testWord       string // For explicit checking of specific test words
		isConventional bool
		expectedValid  bool
		expectedCode   string
	}{
		{
			name:          "Valid imperative",
			message:       "Add feature",
			expectedValid: true,
		},
		{
			name:          "Another valid imperative",
			message:       "Fix bug in code",
			expectedValid: true,
		},
		{
			name:          "Past tense form",
			message:       "Added feature",
			expectedValid: false,
			expectedCode:  string(appErrors.ErrPastTense),
		},
		{
			name:          "Gerund form",
			message:       "Adding feature",
			expectedValid: false,
			expectedCode:  string(appErrors.ErrGerund),
		},
		{
			name:          "Third person",
			message:       "Adds feature",
			expectedValid: false,
			expectedCode:  string(appErrors.ErrThirdPerson),
		},
		{
			name:           "Conventional commit valid",
			message:        "feat: add feature",
			isConventional: true,
			expectedValid:  true,
		},
		{
			name:           "Conventional commit invalid",
			message:        "feat: added feature",
			isConventional: true,
			expectedValid:  false,
			expectedCode:   string(appErrors.ErrPastTense),
		},
		{
			name:           "Conventional with scope valid",
			message:        "feat(api): add endpoint",
			isConventional: true,
			expectedValid:  true,
		},
		{
			name:           "Conventional with scope invalid",
			message:        "feat(api): added endpoint",
			isConventional: true,
			expectedValid:  false,
			expectedCode:   string(appErrors.ErrPastTense),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit info
			commitInfo := domain.CommitInfo{
				Subject: testCase.message,
				Message: testCase.message,
			}

			// Create rule with imperative required and the conventional setting
			var options []rules.ImperativeVerbOption
			if testCase.isConventional {
				options = append(options, rules.WithImperativeConventionalCommit(true))
			}

			rule := rules.NewImperativeVerbRule(options...)

			// Create context with test configuration
			ctx := createImperativeTestContext()

			// Validate and get errors
			errors := rule.Validate(ctx, commitInfo)

			// Check errors
			if testCase.expectedValid {
				require.Empty(t, errors, "Did not expect errors")
			} else {
				require.NotEmpty(t, errors, "Expected errors")

				// Check error code
				if testCase.expectedCode != "" {
					require.Equal(t, testCase.expectedCode, errors[0].Code, "Error code should match expected")
				}

				// Verify rule name is set in the error
				require.Equal(t, "ImperativeVerb", errors[0].Rule, "Rule name should be set in ValidationError")
			}

			// Check name method
			require.Equal(t, "ImperativeVerb", rule.Name(), "Name should be 'ImperativeVerb'")
		})
	}
}

func TestImperativeVerbRuleOptions(t *testing.T) {
	// Test custom non-imperative starters
	t.Run("Custom non-imperative starters", func(t *testing.T) {
		// Create a commit with a past tense word that's not in the default list
		commitInfo := domain.CommitInfo{
			Subject: "Accomplished task",
			Message: "Accomplished task",
		}

		// Create a rule with custom non-imperative starters
		rule := rules.NewImperativeVerbRule(
			rules.WithCustomNonImperativeStarters(map[string][]string{
				"past_tense": {"accomplished"},
			}),
		)

		// Create context with test configuration
		ctx := createImperativeTestContext()

		errors := rule.Validate(ctx, commitInfo)

		// Should fail validation
		require.NotEmpty(t, errors, "Should detect 'accomplished' as past tense")
		require.Equal(t, string(appErrors.ErrPastTense), errors[0].Code, "Error code should be past_tense")
	})

	t.Run("Custom base forms with ED", func(t *testing.T) {
		// Create a commit with a word ending in 'ed' that should be treated as base form
		commitInfo := domain.CommitInfo{
			Subject: "Embed feature",
			Message: "Embed feature",
		}

		// First validate without custom base forms - should falsely detect as past tense
		// because it ends with 'ed'
		rule := rules.NewImperativeVerbRule(
			rules.WithCustomNonImperativeStarters(map[string][]string{
				"past_tense": {"embed"},
			}),
		)

		// Create context with test configuration
		ctx := createImperativeTestContext()

		defaultErrors := rule.Validate(ctx, commitInfo)

		// Should detect as past tense without the custom setting
		require.NotEmpty(t, defaultErrors, "Should initially detect 'embed' as past tense")

		// Now create a rule that allows 'embed' as a base form
		fixedRule := rules.NewImperativeVerbRule(
			rules.WithCustomNonImperativeStarters(map[string][]string{
				"past_tense": {"embed"},
			}),
			rules.WithAdditionalBaseFormsEndingWithED([]string{"embed"}),
		)

		fixedErrors := fixedRule.Validate(ctx, commitInfo)

		// Should now pass validation
		require.Empty(t, fixedErrors, "Should not detect 'embed' as past tense after allowing it as base form")
	})
}
