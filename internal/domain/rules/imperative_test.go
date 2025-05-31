// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

func TestImperativeVerbRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
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
			expectedCode:  string(domain.ErrPastTense),
		},
		{
			name:          "Gerund form",
			message:       "Adding feature",
			expectedValid: false,
			expectedCode:  string(domain.ErrGerund),
		},
		{
			name:          "Third person",
			message:       "Adds feature",
			expectedValid: false,
			expectedCode:  string(domain.ErrThirdPerson),
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
			expectedCode:   string(domain.ErrPastTense),
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
			expectedCode:   string(domain.ErrPastTense),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commit.Message = testCase.message
			commit.Subject = strings.Split(testCase.message, "\n")[0]

			// Create rule with imperative required and the conventional setting
			cfg := config.Config{}
			if testCase.isConventional {
				cfg.Rules.Enabled = []string{"conventional"}
			}

			rule := rules.NewImperativeVerbRule(cfg)

			ctx := context.Background()
			errs := rule.Validate(ctx, commit)

			// Check errors
			if testCase.expectedValid {
				require.Empty(t, errs, "Did not expect errors")
			} else {
				require.NotEmpty(t, errs, "Expected errors")

				// Check error code
				if testCase.expectedCode != "" {
					testdata.AssertValidationError(t, errs[0], testCase.expectedCode, "ImperativeVerb")
				}
			}

			// Check name method
			require.Equal(t, "ImperativeVerb", rule.Name(), "Name should be 'ImperativeVerb'")
		})
	}
}

func TestImperativeVerbRuleOptions(t *testing.T) {
	// Test words ending in 'ed' detection
	t.Run("Words ending in ED detected as past tense", func(t *testing.T) {
		// Create a commit with a past tense word from the default list
		commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		commit.Message = "Added new feature"
		commit.Subject = "Added new feature"

		// Create a rule
		cfg := config.Config{}
		rule := rules.NewImperativeVerbRule(cfg)

		ctx := context.Background()
		errs := rule.Validate(ctx, commit)

		// Should fail validation because 'added' is in the past tense list
		require.NotEmpty(t, errs, "Should detect 'added' as past tense")
		testdata.AssertValidationError(t, errs[0], string(domain.ErrPastTense), "ImperativeVerb")
	})

	t.Run("Base forms ending with ED", func(t *testing.T) {
		// Create a commit with a valid imperative verb
		// The current implementation only checks specific words,
		// not all words ending in 'ed'
		commit := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		commit.Message = "Embed feature"
		commit.Subject = "Embed feature"

		cfg := config.Config{}
		rule := rules.NewImperativeVerbRule(cfg)

		ctx := context.Background()
		errs := rule.Validate(ctx, commit)

		// Should pass validation because 'embed' is not in the predefined lists
		// The rule only checks specific known non-imperative words
		require.Empty(t, errs, "Should not detect 'embed' as past tense")
	})
}
