// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSubjectCase(t *testing.T) {
	testCases := []struct {
		name            string
		isConventional  bool
		subject         string
		caseChoice      string
		expectedValid   bool
		errorCode       string
		expectedMessage string
		expectedErrors  bool
	}{
		{
			name:            "Valid uppercase conventional commit",
			isConventional:  true,
			subject:         "feat: Some feature description",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Valid subject case",
			expectedErrors:  false,
		},
		{
			name:            "Invalid uppercase conventional commit",
			isConventional:  true,
			subject:         "feat: some feature description",
			caseChoice:      "upper",
			errorCode:       "wrong_case_upper",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "Valid lowercase conventional commit",
			isConventional:  true,
			subject:         "feat: some feature description",
			caseChoice:      "lower",
			expectedValid:   true,
			expectedMessage: "Valid subject case",
			expectedErrors:  false,
		},
		{
			name:            "Invalid lowercase conventional commit",
			isConventional:  true,
			subject:         "feat: Some feature description",
			caseChoice:      "lower",
			errorCode:       "wrong_case_lower",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "Valid uppercase non-conventional commit",
			isConventional:  false,
			subject:         "Update something",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Valid subject case",
			expectedErrors:  false,
		},
		{
			name:            "Invalid uppercase non-conventional commit",
			isConventional:  false,
			subject:         "update something",
			caseChoice:      "upper",
			errorCode:       "wrong_case_upper",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "Invalid case choice",
			isConventional:  false,
			subject:         "Some message",
			caseChoice:      "non-valid choice",
			errorCode:       "wrong_case_lower", // Defaults to lower
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "Empty message",
			isConventional:  false,
			subject:         "",
			caseChoice:      "upper",
			errorCode:       "empty_subject",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "Invalid conventional commit format",
			isConventional:  true,
			subject:         "this is missing the colon",
			caseChoice:      "upper",
			errorCode:       "invalid_conventional_format",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "Missing subject after colon",
			isConventional:  true,
			subject:         "feat: ",
			caseChoice:      "upper",
			errorCode:       "invalid_conventional_format",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
		{
			name:            "With scope",
			isConventional:  true,
			subject:         "feat(api): Add endpoint",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Valid subject case",
			expectedErrors:  false,
		},
		{
			name:            "Invalid format non-conventional commit",
			isConventional:  false,
			subject:         "!@# Invalid start",
			caseChoice:      "upper",
			errorCode:       "invalid_format",
			expectedValid:   false,
			expectedMessage: "Invalid subject case",
			expectedErrors:  true,
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Perform the rule
			rule := rule.ValidateSubjectCase(tabletest.subject, tabletest.caseChoice, tabletest.isConventional)

			// Check errors
			if tabletest.expectedErrors {
				require.NotEmpty(t, rule.Errors(), "Expected errors, but got none")

				// Verify the error message
				require.Equal(t, tabletest.expectedMessage, rule.Result(),
					"Error message does not match expected")

				// Verify error code if specified
				if tabletest.errorCode != "" {
					assert.Equal(t, tabletest.errorCode, rule.Errors()[0].Code,
						"Error code should match expected")
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "SubjectCase", rule.Errors()[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify context information is present
				if tabletest.errorCode == "wrong_case_upper" || tabletest.errorCode == "wrong_case_lower" {
					assert.Contains(t, rule.Errors()[0].Context, "expected_case",
						"Context should contain expected case")
					assert.Contains(t, rule.Errors()[0].Context, "actual_case",
						"Context should contain actual case")
				}

				// Check VerboseResult for error cases
				switch tabletest.errorCode {
				case "empty_subject":
					assert.Contains(t, rule.VerboseResult(), "empty", "VerboseResult should explain empty subject")
				case "invalid_conventional_format":
					assert.Contains(t, rule.VerboseResult(), "Invalid conventional commit format",
						"VerboseResult should explain conventional format issue")
				case "invalid_format":
					assert.Contains(t, rule.VerboseResult(), "Invalid commit format",
						"VerboseResult should explain format issue")
				case "missing_conventional_subject":
					assert.Contains(t, rule.VerboseResult(), "Missing subject after conventional commit",
						"VerboseResult should explain missing subject")
				case "wrong_case_upper":
					assert.Contains(t, rule.VerboseResult(), "uppercase",
						"VerboseResult should explain uppercase requirement")
				case "wrong_case_lower":
					assert.Contains(t, rule.VerboseResult(), "lowercase",
						"VerboseResult should explain lowercase requirement")
				}
			} else {
				require.Empty(t, rule.Errors(), "Did not expect errors, but got some")
				require.Equal(t, tabletest.expectedMessage, rule.Result(),
					"Expected message doesn't match")

				// Check VerboseResult for successful cases
				assert.Contains(t, rule.VerboseResult(), "correct", "VerboseResult should mention correct case")

				firstWord := ""

				if tabletest.isConventional {
					if tabletest.subject == "feat: Some feature description" {
						firstWord = "Some"
					} else if tabletest.subject == "feat: some feature description" {
						firstWord = "some"
					} else if tabletest.subject == "feat(api): Add endpoint" {
						firstWord = "Add"
					}
				} else {
					if tabletest.subject == "Update something" {
						firstWord = "Update"
					}
				}

				if firstWord != "" {
					assert.Contains(t, rule.VerboseResult(), firstWord, "VerboseResult should contain the first word")
				}
			}

			// Check name
			require.Equal(t, "SubjectCase", rule.Name(), "Name should always be 'SubjectCase'")

			// Check Help method returns meaningful text
			helpText := rule.Help()
			require.NotEmpty(t, helpText, "Help text should not be empty")
		})
	}
}

func TestSubjectCaseRuleMethods(t *testing.T) {
	const noErrMsg = "No errors to fix"

	t.Run("Name method", func(t *testing.T) {
		rule := &rule.SubjectCase{}
		require.Equal(t, "SubjectCase", rule.Name())
	})

	t.Run("Result method with no errors", func(t *testing.T) {
		rule := &rule.SubjectCase{}
		require.Equal(t, "Valid subject case", rule.Result())
	})

	t.Run("Help method with no errors", func(t *testing.T) {
		rule := &rule.SubjectCase{}
		require.Equal(t, noErrMsg, rule.Help())
	})

	t.Run("Help method with case error", func(t *testing.T) {
		rule := rule.ValidateSubjectCase("lowercase start", "upper", false)
		helpText := rule.Help()
		require.Contains(t, helpText, "Capitalize the first letter", "Help should provide guidance on capitalizing")
	})

	t.Run("Test specific help messages by error code", func(t *testing.T) {
		// Empty subject
		emptyRule := rule.ValidateSubjectCase("", "upper", false)
		assert.Contains(t, emptyRule.Help(), "Provide a non-empty commit message",
			"Help for empty subject should suggest providing a subject")

		// Invalid conventional format
		invalidRule := rule.ValidateSubjectCase("not conventional", "upper", true)
		assert.Contains(t, invalidRule.Help(), "Format your commit message according to the Conventional Commits",
			"Help for invalid format should explain conventional format")

		// Invalid format (non-conventional)
		badFormatRule := rule.ValidateSubjectCase("!@# Invalid", "upper", false)
		assert.Contains(t, badFormatRule.Help(), "Ensure your commit message starts with a valid word",
			"Help for invalid format should explain format requirements")

		// Wrong case - upper
		upperRule := rule.ValidateSubjectCase("lowercase", "upper", false)
		assert.Contains(t, upperRule.Help(), "Capitalize the first letter",
			"Help for uppercase violation should mention capitalization")

		// Wrong case - lower
		lowerRule := rule.ValidateSubjectCase("Uppercase", "lower", false)
		assert.Contains(t, lowerRule.Help(), "Use lowercase",
			"Help for lowercase violation should mention using lowercase")
	})
}
