// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
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
		},
		{
			name:          "Gerund form",
			message:       "Adding feature",
			expectedValid: false,
		},
		{
			name:          "Third person",
			message:       "Adds feature",
			expectedValid: false,
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
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
			commitInfo.Message = testCase.message
			commitInfo.Subject = strings.Split(testCase.message, "\n")[0]
			commit := commitInfo

			// Create rule with imperative required and the conventional setting
			cfg := config.Config{}
			if testCase.isConventional {
				cfg.Rules.Enabled = []string{"conventional"}
			}

			rule := rules.NewImperativeVerbRule(cfg)

			ctx := domain.ValidationContext{
				Commit:     commit,
				Repository: nil,
				Config:     &cfg,
			}
			failures := rule.Validate(ctx)

			// Check errors
			if testCase.expectedValid {
				require.Empty(t, failures, "Did not expect failures")
			} else {
				require.NotEmpty(t, failures, "Expected failures")
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
		commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		commitInfo.Message = "Added new feature"
		commitInfo.Subject = "Added new feature"
		commit := commitInfo

		// Create a rule
		cfg := config.Config{}
		rule := rules.NewImperativeVerbRule(cfg)

		ctx := domain.ValidationContext{
			Commit:     commit,
			Repository: nil,
			Config:     &cfg,
		}
		failures := rule.Validate(ctx)

		// Should fail validation because 'added' is in the past tense list
		require.NotEmpty(t, failures, "Should detect 'added' as past tense")
		testdata.AssertRuleFailure(t, failures[0], "ImperativeVerb")
	})

	t.Run("Base forms ending with ED", func(t *testing.T) {
		// Create a commit with a valid imperative verb
		// The current implementation only checks specific words,
		// not all words ending in 'ed'
		commitInfo := testdata.Commit("feat: add new feature\n\nThis commit adds a new feature that enhances the user experience.")
		commitInfo.Message = "Embed feature"
		commitInfo.Subject = "Embed feature"
		commit := commitInfo

		cfg := config.Config{}
		rule := rules.NewImperativeVerbRule(cfg)

		ctx := domain.ValidationContext{
			Commit:     commit,
			Repository: nil,
			Config:     &cfg,
		}
		failures := rule.Validate(ctx)

		// Should pass validation because 'embed' is not in the predefined lists
		// The rule only checks specific known non-imperative words
		require.Empty(t, failures, "Should not detect 'embed' as past tense")
	})
}
