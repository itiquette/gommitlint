// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateImperative(t *testing.T) {
	testCases := []struct {
		name            string
		isConventional  bool
		message         string
		expectedValid   bool
		expectedCode    string
		expectedMessage string
	}{
		{
			name:            "Valid imperative verb",
			isConventional:  false,
			message:         "Add new feature",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Valid imperative verb in conventional commit",
			isConventional:  true,
			message:         "feat: Add new feature",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Valid imperative verb in conventional commit with scope",
			isConventional:  true,
			message:         "feat(auth): Add login button",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Valid imperative verb in conventional commit with scope and breaking change",
			isConventional:  true,
			message:         "feat(auth)!: Add OAuth support",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Non-imperative past tense verb",
			isConventional:  false,
			message:         "Added new feature",
			expectedValid:   false,
			expectedCode:    "past_tense",
			expectedMessage: "first word of commit must be an imperative verb: \"Added\" appears to be past tense",
		},
		{
			name:            "Non-imperative gerund",
			isConventional:  false,
			message:         "Adding new feature",
			expectedValid:   false,
			expectedCode:    "gerund",
			expectedMessage: "first word of commit must be an imperative verb: \"Adding\" appears to be a gerund",
		},
		{
			name:            "Non-imperative third person",
			isConventional:  false,
			message:         "Adds new feature",
			expectedValid:   false,
			expectedCode:    "third_person",
			expectedMessage: "first word of commit must be an imperative verb: \"Adds\" appears to be 3rd person present",
		},
		{
			name:            "Empty message",
			isConventional:  false,
			message:         "",
			expectedValid:   false,
			expectedCode:    "empty_message",
			expectedMessage: "empty message",
		},
		{
			name:            "Invalid conventional commit format",
			isConventional:  true,
			message:         "invalid-format",
			expectedValid:   false,
			expectedCode:    "invalid_format",
			expectedMessage: "invalid conventional commit format",
		},
		{
			name:            "Invalid conventional commit - missing subject",
			isConventional:  true,
			message:         "feat: ",
			expectedValid:   false,
			expectedCode:    "missing_subject",
			expectedMessage: "missing subject after type",
		},
		{
			name:            "Unicode characters",
			isConventional:  false,
			message:         "Résolve élément issue",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Non-imperative verb in conventional commit",
			isConventional:  true,
			message:         "fix(core): Added new validation",
			expectedValid:   false,
			expectedCode:    "past_tense",
			expectedMessage: "first word of commit must be an imperative verb: \"Added\" appears to be past tense",
		},
		{
			name:            "Non-verb word",
			isConventional:  false,
			message:         "The new feature",
			expectedValid:   false,
			expectedCode:    "non_verb",
			expectedMessage: "first word of commit must be an imperative verb: \"The\" is not a verb",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			rule := rule.ValidateImperative(tabletest.message, tabletest.isConventional)

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

				// Check error code
				if tabletest.expectedCode != "" {
					assert.Equal(t, tabletest.expectedCode, rule.Errors()[0].Code, "Error code should match expected")
				}

				// Check error message
				require.Equal(t,
					tabletest.expectedMessage,
					rule.Result(),
					"Error message should match expected",
				)

				// Verify rule name is set in the error
				assert.Equal(t, "ImperativeVerb", rule.Errors()[0].Rule, "Rule name should be set in ValidationError")
			}

			// Check name method
			require.Equal(t, "ImperativeVerb", rule.Name(),
				"Name should be 'ImperativeVerb'")

			// Check help method is not empty
			assert.NotEmpty(t, rule.Help(), "Help text should not be empty")
		})
	}
}

func TestMixedConventionalAndNonConventional(t *testing.T) {
	// Test handling of conventional commits with complex scopes
	complexScopeCommit := rule.ValidateImperative("feat(user/auth,permissions): Add role-based access", true)
	require.Empty(t, complexScopeCommit.Errors(), "Should allow complex scopes in conventional commits")

	// Test handling of commits with weird formatting but valid imperative verbs
	weirdFormattingCommit := rule.ValidateImperative("  \t  Fix bug", false)
	require.Empty(t, weirdFormattingCommit.Errors(), "Should handle leading whitespace")

	// Test handling of base form verbs that end in "ed"
	edBaseFormCommit := rule.ValidateImperative("Proceed with implementation", false)
	require.Empty(t, edBaseFormCommit.Errors(), "Should recognize 'proceed' as base form")

	// Test handling of base form verbs that end in "s"
	sBaseFormCommit := rule.ValidateImperative("Process data correctly", false)
	require.Empty(t, sBaseFormCommit.Errors(), "Should recognize 'process' as base form")
}

func TestImperativeVerbMethods(t *testing.T) {
	t.Run("Name method", func(t *testing.T) {
		rule := &rule.ImperativeVerb{}
		require.Equal(t, "ImperativeVerb", rule.Name())
	})

	t.Run("Result method with errors", func(t *testing.T) {
		rule := &rule.ImperativeVerb{}
		err := model.NewValidationError("ImperativeVerb", "past_tense",
			"first word of commit must be an imperative verb: \"Added\" is invalid")
		rule.SetErrors([]*model.ValidationError{err})
		require.Equal(t,
			"first word of commit must be an imperative verb: \"Added\" is invalid",
			rule.Result(),
		)
	})

	t.Run("Result method without errors", func(t *testing.T) {
		rule := &rule.ImperativeVerb{}
		require.Equal(t, "Commit begins with imperative verb", rule.Result())
	})

	t.Run("Errors method", func(t *testing.T) {
		expectedError := model.NewValidationError("ImperativeVerb", "test_error", "test error")
		expectedErrors := []*model.ValidationError{expectedError}

		ruleInstance := &rule.ImperativeVerb{}
		ruleInstance.SetErrors(expectedErrors)
		require.Equal(t, expectedErrors, ruleInstance.Errors())
	})

	t.Run("Help method with different errors", func(t *testing.T) {
		// Test help text for empty message
		emptyRule := &rule.ImperativeVerb{}
		emptyRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "empty_message", "empty message"),
		})
		assert.Contains(t, emptyRule.Help(), "Provide a non-empty commit message")

		// Test help text for invalid format
		formatRule := &rule.ImperativeVerb{}
		formatRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "invalid_format", "invalid conventional commit format"),
		})
		assert.Contains(t, formatRule.Help(), "Format your commit message according to the Conventional Commits")

		// Test help text for missing subject
		missingSubjectRule := &rule.ImperativeVerb{}
		missingSubjectRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "missing_subject", "missing subject after type"),
		})
		assert.Contains(t, missingSubjectRule.Help(), "Add a description after the type and colon")

		// Test help text for non-imperative verb
		verbRule := &rule.ImperativeVerb{}
		verbRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "non_verb", "first word of commit must be an imperative verb"),
		})
		assert.Contains(t, verbRule.Help(), "Use the imperative mood")

		// Test help text for past tense
		pastTenseRule := &rule.ImperativeVerb{}
		pastTenseRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "past_tense", "word appears to be past tense"),
		})
		assert.Contains(t, pastTenseRule.Help(), "Avoid using past tense verbs")

		// Test help text for no errors
		validRule := &rule.ImperativeVerb{}
		assert.Equal(t, "No errors to fix", validRule.Help())
	})
}

// TestAddError verifies the behavior of the ValidationError creation.
func TestAddError(t *testing.T) {
	vrule := rule.ValidateImperative("", false)
	require.NotEmpty(t, vrule.Errors(), "Should have errors for empty message")
	require.Equal(t, "empty message", vrule.Result(), "Error message should match")
	assert.Equal(t, "empty_message", vrule.Errors()[0].Code, "Error code should be set")

	// Check the context information in errors
	verbRule := rule.ValidateImperative("Added feature", false)
	require.NotEmpty(t, verbRule.Errors(), "Should have errors for past tense")
	assert.Contains(t, verbRule.Result(), "\"Added\"", "Error should contain the word in quotes")
	assert.Equal(t, "past_tense", verbRule.Errors()[0].Code, "Error code should be set")
	assert.Equal(t, "Added", verbRule.Errors()[0].Context["word"], "Word should be included in context")
}

// TestErrorsWithSubjectRegex verifies that the

// TestErrorsWithSubjectRegex verifies that the subject regex properly handles different commit formats.
func TestErrorsWithSubjectRegex(t *testing.T) {
	// Test correct conventional format with type and subject
	t.Run("Basic conventional format", func(t *testing.T) {
		rule := rule.ValidateImperative("feat: Add feature", true)
		require.Empty(t, rule.Errors(), "Should accept basic conventional format")
	})

	// Test with scope
	t.Run("With scope", func(t *testing.T) {
		rule := rule.ValidateImperative("feat(ui): Add button", true)
		require.Empty(t, rule.Errors(), "Should accept conventional format with scope")
	})

	// Test with breaking change indicator
	t.Run("With breaking change", func(t *testing.T) {
		rule := rule.ValidateImperative("feat(api)!: Change endpoint signatures", true)
		require.Empty(t, rule.Errors(), "Should accept breaking changes")
	})

	// Test without space after colon (invalid)
	t.Run("Without space after colon", func(t *testing.T) {
		rule := rule.ValidateImperative("feat:Add feature", true)
		require.NotEmpty(t, rule.Errors(), "Should reject without space after colon")
	})

	// Test with invalid type format
	t.Run("Invalid type", func(t *testing.T) {
		rule := rule.ValidateImperative("Feat: Add feature", true)
		require.NotEmpty(t, rule.Errors(), "Should reject uppercase type")
	})
}
