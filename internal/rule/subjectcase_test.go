// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateSubjectCase(t *testing.T) {
	testCases := []struct {
		name            string
		isConventional  bool
		subject         string
		caseChoice      string
		expectedValid   bool
		expectedMessage string
		expectedErrors  bool
	}{
		{
			name:            "Valid uppercase conventional commit",
			isConventional:  true,
			subject:         "feat: Some feature description",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Subject case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid uppercase conventional commit",
			isConventional:  true,
			subject:         "feat: some feature description",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "commit subject case is not upper",
			expectedErrors:  true,
		},
		{
			name:            "Valid lowercase conventional commit",
			isConventional:  true,
			subject:         "feat: some feature description",
			caseChoice:      "lower",
			expectedValid:   true,
			expectedMessage: "Subject case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid lowercase conventional commit",
			isConventional:  true,
			subject:         "feat: Some feature description",
			caseChoice:      "lower",
			expectedValid:   false,
			expectedMessage: "commit subject case is not lower",
			expectedErrors:  true,
		},
		{
			name:            "Valid uppercase non-conventional commit",
			isConventional:  false,
			subject:         "Update something",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Subject case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid uppercase non-conventional commit",
			isConventional:  false,
			subject:         "update something",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "commit subject case is not upper",
			expectedErrors:  true,
		},
		{
			name:            "Invalid case choice",
			isConventional:  false,
			subject:         "Some message",
			caseChoice:      "non-valid choice",
			expectedValid:   false,
			expectedMessage: "commit subject case is not non-valid choice",
			expectedErrors:  true,
		},
		{
			name:            "Empty message",
			isConventional:  false,
			subject:         "",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "subject is empty",
			expectedErrors:  true,
		},
		{
			name:            "Invalid conventional commit format",
			isConventional:  true,
			subject:         "this is missing the colon",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "invalid conventional commit format",
			expectedErrors:  true,
		},
		{
			name:            "Missing subject after colon",
			isConventional:  true,
			subject:         "feat: ",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "invalid conventional commit format",
			expectedErrors:  true,
		},
		{
			name:            "With scope",
			isConventional:  true,
			subject:         "feat(api): Add endpoint",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Subject case is valid",
			expectedErrors:  false,
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
			} else {
				require.Empty(t, rule.Errors(), "Did not expect errors, but got some")
				require.Equal(t, tabletest.expectedMessage, rule.Result(),
					"Expected message doesn't match")
			}

			// Check status
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
		require.Equal(t, "Subject case is valid", rule.Result())
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
}
