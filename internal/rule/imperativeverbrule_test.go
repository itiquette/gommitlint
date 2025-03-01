// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"errors"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateImperativeRule(t *testing.T) {
	testCases := []struct {
		name            string
		isConventional  bool
		message         string
		expectedValid   bool
		expectedMessage string
	}{
		{
			name:            "Valid Imperative Verb",
			isConventional:  false,
			message:         "Add new feature",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Valid Imperative Verb in Conventional Commit",
			isConventional:  true,
			message:         "feat: Add new feature",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Non-Imperative Past Tense Verb",
			isConventional:  false,
			message:         "Added new feature",
			expectedValid:   false,
			expectedMessage: "first word of commit must be an imperative verb: \"Added\" is invalid",
		},
		{
			name:            "Non-Imperative Gerund",
			isConventional:  false,
			message:         "Adding new feature",
			expectedValid:   false,
			expectedMessage: "first word of commit must be an imperative verb: \"Adding\" is invalid",
		},
		{
			name:            "Non-Imperative Third Person",
			isConventional:  false,
			message:         "Adds new feature",
			expectedValid:   false,
			expectedMessage: "first word of commit must be an imperative verb: \"Adds\" is invalid",
		},
		{
			name:            "Empty Message",
			isConventional:  false,
			message:         "",
			expectedValid:   false,
			expectedMessage: "empty message",
		},
		{
			name:            "Invalid Conventional Commit Format",
			isConventional:  true,
			message:         "invalid-format",
			expectedValid:   false,
			expectedMessage: "invalid conventional commit format",
		},
		{
			name:            "Unicode Characters",
			isConventional:  false,
			message:         "Résolve élément issue",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			rule := rule.ValidateImperativeRule(tabletest.message, tabletest.isConventional)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, rule.Errors(), "Did not expect errors")
				require.Equal(t,
					"Commit begins with imperative verb",
					rule.Result(),
					"Message should be valid",
				)
			} else {
				require.NotEmpty(t, rule.Errors(), "Expected errors")
				require.Equal(t,
					tabletest.expectedMessage,
					rule.Result(),
					"Error message should match expected",
				)
			}

			// Check status method
			require.Equal(t, "ImperativeVerbRule", rule.Name(),
				"Status should always be 'ImperativeVerbRule'")
		})
	}
}

func TestImperativeRuleMethods(t *testing.T) {
	t.Run("Status Method", func(t *testing.T) {
		rule := &rule.ImperativeVerbRule{}
		require.Equal(t, "ImperativeVerbRule", rule.Name())
	})

	t.Run("Message Method with Errors", func(t *testing.T) {
		rule := &rule.ImperativeVerbRule{}
		rule.SetErrors([]error{errors.New("first word of commit must be an imperative verb: \"Added\" is invalid")})
		require.Equal(t,
			"first word of commit must be an imperative verb: \"Added\" is invalid",
			rule.Result(),
		)
	})

	t.Run("Message Method without Errors", func(t *testing.T) {
		rule := &rule.ImperativeVerbRule{}
		require.Equal(t, "Commit begins with imperative verb", rule.Result())
	})

	t.Run("Errors Method", func(t *testing.T) {
		expectedErrors := []error{
			errors.New("test error"),
		}
		ruleInstance := &rule.ImperativeVerbRule{}
		ruleInstance.SetErrors(expectedErrors)
		require.Equal(t, expectedErrors, ruleInstance.Errors())
	})
}
