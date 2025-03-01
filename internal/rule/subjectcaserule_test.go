// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateSubjectCaseRule(t *testing.T) {
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
			name:            "Valid Uppercase Conventional Commit",
			isConventional:  true,
			subject:         "feat: Some feature description",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "SubjectCaseRule is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid Uppercase Conventional Commit",
			isConventional:  true,
			subject:         "feat: some feature description",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "commit subject case is not upper",
			expectedErrors:  true,
		},
		{
			name:            "Valid Lowercase Conventional Commit",
			isConventional:  true,
			subject:         "feat: some feature description",
			caseChoice:      "lower",
			expectedValid:   true,
			expectedMessage: "subject case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid Lowercase Conventional Commit",
			isConventional:  true,
			subject:         "feat: Some feature description",
			caseChoice:      "lower",
			expectedValid:   false,
			expectedMessage: "commit subject case is not lower",
			expectedErrors:  true,
		},
		{
			name:            "Valid Uppercase Non-Conventional Commit",
			isConventional:  false,
			subject:         "Update something",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "subject case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid Uppercase Non-Conventional Commit",
			isConventional:  false,
			subject:         "update something",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "commit subject case is not upper",
			expectedErrors:  true,
		},
		{
			name:            "Invalid Case Choice",
			isConventional:  false,
			subject:         "Some message",
			caseChoice:      "non-valid choice",
			expectedValid:   false,
			expectedMessage: "commit subject case is not non-valid choice",
			expectedErrors:  true,
		},
		{
			name:            "Empty Message",
			isConventional:  false,
			subject:         "",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "empty message",
			expectedErrors:  true,
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Perform the rule
			rule := rule.ValidateSubjectCaseRule(tabletest.subject, tabletest.caseChoice, tabletest.isConventional)

			// Check errors
			if tabletest.expectedErrors {
				require.NotEmpty(t, rule.Errors(), "Expected errors, but got none")

				// Verify the error message
				require.Equal(t, tabletest.expectedMessage, rule.Result(),
					"Error message does not match expected")
			} else {
				require.Empty(t, rule.Errors(), "Did not expect errors, but got some")
				require.Equal(t, "SubjectCaseRule is valid", rule.Result(),
					"Expected default valid message")
			}

			// Check status
			require.Equal(t, "SubjectCaseRule", rule.Name(), "Status should always be 'SubjectCaseRule'")
		})
	}
}

func TestSubjectCaseRuleMethods(t *testing.T) {
	t.Run("Status Method", func(t *testing.T) {
		rule := &rule.SubjectCaseRule{}
		require.Equal(t, "SubjectCaseRule", rule.Name())
	})

	t.Run("Message Method with No Errors", func(t *testing.T) {
		rule := &rule.SubjectCaseRule{}
		require.Equal(t, "SubjectCaseRule is valid", rule.Result())
	})

	// t.Run("Message Method with Errors", func(t *testing.T) {
	// 	rule := &rule.SubjectCase{}
	// 	require.Equal(t, "test error", rule.Message())
	// })

	t.Run("Errors Method", func(t *testing.T) {
		expectedErrors := []error{
			errors.New("test error"),
			errors.New("second error"),
		}
		check := &rule.SubjectCaseRule{
			RuleErrors: expectedErrors,
		}
		require.Equal(t, expectedErrors, check.Errors())
	})
}
