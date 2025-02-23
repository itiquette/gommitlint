// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rule_test

import (
	"errors"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateImperative(t *testing.T) {
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
			// Perform the check
			check := rule.ValidateImperative(tabletest.isConventional, tabletest.message)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, check.Errors(), "Did not expect errors")
				require.Equal(t,
					"Commit begins with imperative verb",
					check.Message(),
					"Message should be valid",
				)
			} else {
				require.NotEmpty(t, check.Errors(), "Expected errors")
				require.Equal(t,
					tabletest.expectedMessage,
					check.Message(),
					"Error message should match expected",
				)
			}

			// Check status method
			require.Equal(t, "Imperative Mood", check.Name(),
				"Status should always be 'Imperative Mood'")
		})
	}
}

func TestImperativeCheckMethods(t *testing.T) {
	t.Run("Status Method", func(t *testing.T) {
		check := &rule.ImperativeVerb{}
		require.Equal(t, "Imperative Mood", check.Name())
	})

	t.Run("Message Method with Errors", func(t *testing.T) {
		check := &rule.ImperativeVerb{}
		check.SetErrors([]error{errors.New("first word of commit must be an imperative verb: \"Added\" is invalid")})
		require.Equal(t,
			"first word of commit must be an imperative verb: \"Added\" is invalid",
			check.Message(),
		)
	})

	t.Run("Message Method without Errors", func(t *testing.T) {
		check := &rule.ImperativeVerb{}
		require.Equal(t, "Commit begins with imperative verb", check.Message())
	})

	t.Run("Errors Method", func(t *testing.T) {
		expectedErrors := []error{
			errors.New("test error"),
		}
		ruleInstance := &rule.ImperativeVerb{}
		ruleInstance.SetErrors(expectedErrors)
		require.Equal(t, expectedErrors, ruleInstance.Errors())
	})
}
