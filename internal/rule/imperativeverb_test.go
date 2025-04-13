// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strings"
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
			expectedCode:    "invalid_conventional_format",
			expectedMessage: "invalid conventional commit format",
		},
		{
			name:            "Invalid conventional commit - missing subject",
			isConventional:  true,
			message:         "feat: ",
			expectedValid:   false,
			expectedCode:    "missing_conventional_subject",
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
		// Additional test cases
		{
			name:            "Leading whitespace",
			isConventional:  false,
			message:         "  \t  Fix bug",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Conventional commit with multi-part scope",
			isConventional:  true,
			message:         "feat(user/auth,permissions): Add role-based access",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Imperative verb ending in 'ed'",
			isConventional:  false,
			message:         "Proceed with implementation",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Imperative verb ending in 's'",
			isConventional:  false,
			message:         "Process data correctly",
			expectedValid:   true,
			expectedMessage: "Commit begins with imperative verb",
		},
		{
			name:            "Conventional commit - uppercase type",
			isConventional:  true,
			message:         "Feat: Add feature",
			expectedValid:   false,
			expectedCode:    "invalid_conventional_format",
			expectedMessage: "invalid conventional commit format",
		},
		{
			name:            "Conventional commit - no space after colon",
			isConventional:  true,
			message:         "feat:Add feature",
			expectedValid:   false,
			expectedCode:    "invalid_conventional_format",
			expectedMessage: "invalid conventional commit format",
		},
		{
			name:            "Invalid non-conventional format",
			isConventional:  false,
			message:         "!@#$ invalid start",
			expectedValid:   false,
			expectedCode:    "invalid_format",
			expectedMessage: "invalid commit format",
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

				// Check verbose result for valid case
				if tabletest.message == "Add new feature" {
					assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
				} else if tabletest.message == "feat: Add new feature" {
					assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
				} else if tabletest.message == "feat(auth): Add login button" {
					assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
				} else if tabletest.message == "feat(auth)!: Add OAuth support" {
					assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
				}
			} else {
				require.NotEmpty(t, rule.Errors(), "Expected errors")

				// Check error code
				if tabletest.expectedCode != "" {
					assert.Equal(t, tabletest.expectedCode, rule.Errors()[0].Code, "Error code should match expected")
				}

				// Check error message - Implementation now returns a simplified message
				expectedResult := "Non-imperative verb detected"
				require.Equal(t,
					expectedResult,
					rule.Result(),
					"Error message should match expected",
				)

				// Verify rule name is set in the error
				assert.Equal(t, "ImperativeVerb", rule.Errors()[0].Rule, "Rule name should be set in ValidationError")

				// Test VerboseResult for error cases
				switch tabletest.expectedCode {
				case "past_tense":
					if tabletest.message == "Added new feature" {
						assert.Contains(t, rule.VerboseResult(), "Added", "VerboseResult should contain the detected verb")
						assert.Contains(t, rule.VerboseResult(), "past tense", "VerboseResult should explain the issue")
					}
				case "gerund":
					if tabletest.message == "Adding new feature" {
						assert.Contains(t, rule.VerboseResult(), "Adding", "VerboseResult should contain the detected verb")
						assert.Contains(t, rule.VerboseResult(), "gerund", "VerboseResult should explain the issue")
					}
				case "third_person":
					if tabletest.message == "Adds new feature" {
						assert.Contains(t, rule.VerboseResult(), "Adds", "VerboseResult should contain the detected verb")
						assert.Contains(t, rule.VerboseResult(), "third person", "VerboseResult should explain the issue")
					}
				case "non_verb":
					if tabletest.message == "The new feature" {
						assert.Contains(t, rule.VerboseResult(), "The", "VerboseResult should contain the detected word")
						assert.Contains(t, rule.VerboseResult(), "non-verb", "VerboseResult should explain the issue")
					}
				case "invalid_conventional_format":
					assert.Contains(t, rule.VerboseResult(), "Invalid conventional commit format",
						"VerboseResult should explain conventional format issues")
				case "invalid_format":
					assert.Contains(t, rule.VerboseResult(), "Invalid commit format",
						"VerboseResult should explain general format issues")
				case "missing_conventional_subject":
					assert.Contains(t, rule.VerboseResult(), "Missing subject after conventional commit type",
						"VerboseResult should explain missing subject")
				case "empty_message":
					assert.Contains(t, rule.VerboseResult(), "Commit message is empty",
						"VerboseResult should explain empty message")
				}
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
	assert.Contains(t, complexScopeCommit.VerboseResult(), "Add", "VerboseResult should contain the verb 'Add'")

	// Test handling of commits with weird formatting but valid imperative verbs
	weirdFormattingCommit := rule.ValidateImperative("  \t  Fix bug", false)
	require.Empty(t, weirdFormattingCommit.Errors(), "Should handle leading whitespace")
	assert.Contains(t, weirdFormattingCommit.VerboseResult(), "Fix", "VerboseResult should contain the verb 'Fix'")

	// Test handling of base form verbs that end in "ed"
	edBaseFormCommit := rule.ValidateImperative("Proceed with implementation", false)
	require.Empty(t, edBaseFormCommit.Errors(), "Should recognize 'proceed' as base form")
	assert.Contains(t, edBaseFormCommit.VerboseResult(), "Proceed", "VerboseResult should contain the verb 'Proceed'")

	// Test handling of base form verbs that end in "s"
	sBaseFormCommit := rule.ValidateImperative("Process data correctly", false)
	require.Empty(t, sBaseFormCommit.Errors(), "Should recognize 'process' as base form")
	assert.Contains(t, sBaseFormCommit.VerboseResult(), "Process", "VerboseResult should contain the verb 'Process'")

	// Test handling of invalid format for non-conventional commits
	invalidFormatCommit := rule.ValidateImperative("@!# Invalid start", false)
	require.NotEmpty(t, invalidFormatCommit.Errors(), "Should reject invalid format")
	assert.Equal(t, "invalid_format", invalidFormatCommit.Errors()[0].Code, "Should have correct error code")
	assert.Contains(t, invalidFormatCommit.VerboseResult(), "Invalid commit format", "VerboseResult should explain format issue")
}

func TestImperativeVerbMethods(t *testing.T) {
	t.Run("Name method", func(t *testing.T) {
		rule := &rule.ImperativeVerb{}
		require.Equal(t, "ImperativeVerb", rule.Name())
	})

	t.Run("Result method with errors", func(t *testing.T) {
		vrule := &rule.ImperativeVerb{}
		err := model.NewValidationError("ImperativeVerb", "past_tense",
			"first word of commit must be an imperative verb: \"Added\" is invalid")
		vrule.SetErrors([]*model.ValidationError{err})
		require.Equal(t,
			"Non-imperative verb detected",
			vrule.Result(),
			"Implementation now returns a simplified error message",
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

		// Test help text for invalid conventional format
		formatRule := &rule.ImperativeVerb{}
		formatRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "invalid_conventional_format", "invalid conventional commit format"),
		})
		assert.Contains(t, formatRule.Help(), "Format your commit message according to the Conventional Commits")

		// Test help text for invalid general format
		badFormatRule := &rule.ImperativeVerb{}
		badFormatRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "invalid_format", "invalid commit format"),
		})
		assert.Contains(t, badFormatRule.Help(), "Make sure your commit message starts with an imperative verb")

		// Test help text for missing subject
		missingSubjectRule := &rule.ImperativeVerb{}
		missingSubjectRule.SetErrors([]*model.ValidationError{
			model.NewValidationError("ImperativeVerb", "missing_conventional_subject", "missing subject after type"),
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
	require.Equal(t, "Non-imperative verb detected", vrule.Result(), "Error message should match")
	assert.Equal(t, "empty_message", vrule.Errors()[0].Code, "Error code should be set")
	assert.Contains(t, vrule.VerboseResult(), "Commit message is empty", "VerboseResult should explain the issue")

	// Check the context information in errors
	verbRule := rule.ValidateImperative("Added feature", false)
	require.NotEmpty(t, verbRule.Errors(), "Should have errors for past tense")
	assert.Equal(t, "Non-imperative verb detected", verbRule.Result(), "Error message should be generic")
	assert.Equal(t, "past_tense", verbRule.Errors()[0].Code, "Error code should be set")
	assert.Contains(t, verbRule.VerboseResult(), "Added", "VerboseResult should contain the detected verb")
	assert.Contains(t, verbRule.VerboseResult(), "past tense", "VerboseResult should explain the issue")

	// The error context should still contain information about the word
	if value, exists := verbRule.Errors()[0].Context["word"]; exists {
		assert.Equal(t, "Added", value, "Word should be included in context")
	}

	// Test that invalid format for non-conventional commits is handled correctly
	invalidFormatRule := rule.ValidateImperative("#@! Bad start", false)
	require.NotEmpty(t, invalidFormatRule.Errors(), "Should have errors for invalid format")
	assert.Equal(t, "invalid_format", invalidFormatRule.Errors()[0].Code, "Error code should be set")
	assert.Contains(t, invalidFormatRule.VerboseResult(), "Invalid commit format", "VerboseResult should explain format issue")
}

// TestErrorsWithSubjectRegex verifies that the subject regex properly handles different commit formats.
func TestErrorsWithSubjectRegex(t *testing.T) {
	// Test correct conventional format with type and subject
	t.Run("Basic conventional format", func(t *testing.T) {
		rule := rule.ValidateImperative("feat: Add feature", true)
		require.Empty(t, rule.Errors(), "Should accept basic conventional format")
		assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
	})

	// Test with scope
	t.Run("With scope", func(t *testing.T) {
		rule := rule.ValidateImperative("feat(ui): Add button", true)
		require.Empty(t, rule.Errors(), "Should accept conventional format with scope")
		assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
	})

	// Test with breaking change indicator
	t.Run("With breaking change", func(t *testing.T) {
		rule := rule.ValidateImperative("feat(api)!: Change endpoint signatures", true)
		require.Empty(t, rule.Errors(), "Should accept breaking changes")
		assert.Contains(t, rule.VerboseResult(), "Change", "VerboseResult should contain the verb")
	})

	// Test without space after colon (invalid)
	t.Run("Without space after colon", func(t *testing.T) {
		rule := rule.ValidateImperative("feat:Add feature", true)
		require.NotEmpty(t, rule.Errors(), "Should reject without space after colon")
		assert.Equal(t, "invalid_conventional_format", rule.Errors()[0].Code, "Should have correct error code")
		assert.Contains(t, rule.VerboseResult(), "Invalid conventional commit format", "VerboseResult should explain the issue")
	})

	// Test with invalid type format
	t.Run("Invalid type", func(t *testing.T) {
		rule := rule.ValidateImperative("Feat: Add feature", true)
		require.NotEmpty(t, rule.Errors(), "Should reject uppercase type")
		assert.Equal(t, "invalid_conventional_format", rule.Errors()[0].Code, "Should have correct error code")
		assert.Contains(t, rule.VerboseResult(), "Invalid conventional commit format", "VerboseResult should explain the issue")
	})

	// Test with multiple scopes
	t.Run("Multiple scopes", func(t *testing.T) {
		rule := rule.ValidateImperative("feat(ui,api): Add integration", true)
		require.Empty(t, rule.Errors(), "Should accept multiple scopes")
		assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
	})

	// Test with nested scope path
	t.Run("Nested scope path", func(t *testing.T) {
		rule := rule.ValidateImperative("feat(core/validation): Update rules", true)
		require.Empty(t, rule.Errors(), "Should accept nested scope paths")
		assert.Contains(t, rule.VerboseResult(), "Update", "VerboseResult should contain the verb")
	})
}

// TestVerboseResult verifies the behavior of the VerboseResult method.
func TestVerboseResult(t *testing.T) {
	// Test verbose result with no errors
	t.Run("Verbose result with no errors", func(t *testing.T) {
		rule := rule.ValidateImperative("Add feature", false)
		require.Empty(t, rule.Errors(), "Should have no errors")
		assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
	})

	// Test verbose result with past tense error
	t.Run("Verbose result with past tense error", func(t *testing.T) {
		rule := rule.ValidateImperative("Added feature", false)
		require.NotEmpty(t, rule.Errors(), "Should have errors")
		assert.Contains(t, rule.VerboseResult(), "Added", "VerboseResult should contain the detected verb")
		assert.Contains(t, rule.VerboseResult(), "past tense", "VerboseResult should explain the issue")
	})

	// Test verbose result with invalid format error
	t.Run("Verbose result with invalid format", func(t *testing.T) {
		rule := rule.ValidateImperative("!@# Invalid format", false)
		require.NotEmpty(t, rule.Errors(), "Should have errors")
		assert.Equal(t, "invalid_format", rule.Errors()[0].Code, "Should have correct error code")
		assert.Contains(t, rule.VerboseResult(), "Invalid commit format", "VerboseResult should explain format issue")
	})

	// Test verbose result with missing conventional subject
	t.Run("Verbose result with missing subject", func(t *testing.T) {
		rule := rule.ValidateImperative("feat: ", true)
		require.NotEmpty(t, rule.Errors(), "Should have errors")
		assert.Equal(t, "missing_conventional_subject", rule.Errors()[0].Code, "Should have correct error code")
		assert.Contains(t, rule.VerboseResult(), "Missing subject", "VerboseResult should explain missing subject")
	})
}

// TestEdgeCases verifies behavior with edge cases.
func TestEdgeCases(t *testing.T) {
	// Test with non-standard verb forms to verify robustness
	t.Run("Difficult to detect verb forms", func(t *testing.T) {
		validEdgeCases := []string{
			"Refactor code",    // Common imperative
			"Fix bugs",         // Plural noun after verb
			"Update README.md", // File name in capital letters
			"Revert commit",    // Special git term
			"Debug issue",      // Verb that could be a noun
			"Setup framework",  // Could be miscategorized as noun
			"Rebase branch",    // Git-specific verb
			"Optimize query",   // Technical term
			"Test feature",     // Could be noun or verb
		}

		for _, msg := range validEdgeCases {
			rule := rule.ValidateImperative(msg, false)
			assert.Empty(t, rule.Errors(), "Should accept valid imperative verb: "+msg)
			assert.Contains(t, rule.VerboseResult(), strings.Split(msg, " ")[0], "VerboseResult should contain the verb")
		}

		invalidEdgeCases := []string{
			"Tests feature",   // 3rd person
			"Testing feature", // Gerund
			"Tested feature",  // Past tense
			"A new feature",   // Article
		}

		for _, msg := range invalidEdgeCases {
			rule := rule.ValidateImperative(msg, false)
			assert.NotEmpty(t, rule.Errors(), "Should reject non-imperative form: "+msg)

			firstWord := strings.Split(msg, " ")[0]
			assert.Contains(t, rule.VerboseResult(), firstWord, "VerboseResult should contain the detected word")

			switch firstWord {
			case "Tests":
				assert.Contains(t, rule.VerboseResult(), "third person", "VerboseResult should explain the issue")
			case "Testing":
				assert.Contains(t, rule.VerboseResult(), "gerund", "VerboseResult should explain the issue")
			case "Tested":
				assert.Contains(t, rule.VerboseResult(), "past tense", "VerboseResult should explain the issue")
			case "A":
				assert.Contains(t, rule.VerboseResult(), "non-verb", "VerboseResult should explain the issue")
			}
		}
	})

	// Test with special characters and unusual formatting
	t.Run("Special formatting", func(t *testing.T) {
		// Multiple leading spaces and tabs
		vrule := rule.ValidateImperative("\t  \t  Fix bug", false)
		assert.Empty(t, vrule.Errors(), "Should handle excessive whitespace")
		assert.Contains(t, vrule.VerboseResult(), "Fix", "VerboseResult should contain the verb")

		// Unicode characters
		// Unicode characters
		vrule = rule.ValidateImperative("Résolve élément issue", false) // French
		assert.Empty(t, vrule.Errors(), "Should handle non-ASCII characters")

		// Emoji
		vrule = rule.ValidateImperative("✨ Add feature", false)
		// This might fail depending on how the regex is implemented
		if len(vrule.Errors()) > 0 {
			assert.Equal(t, "invalid_format", vrule.Errors()[0].Code,
				"Should handle emoji with appropriate error code if failed")
			assert.Contains(t, vrule.VerboseResult(), "Invalid commit format",
				"VerboseResult should explain the issue with emoji")
		} else {
			assert.Contains(t, vrule.VerboseResult(), "Add", "VerboseResult should contain the verb if passed")
		}
	})
}
