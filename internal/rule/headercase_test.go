// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateHeaderCase(t *testing.T) {
	testCases := []struct {
		name            string
		isConventional  bool
		message         string
		caseChoice      string
		expectedValid   bool
		expectedMessage string
		expectedErrors  bool
	}{
		{
			name:            "Valid Uppercase Conventional Commit",
			isConventional:  true,
			message:         "feat: Some feature description",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Header case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid Uppercase Conventional Commit",
			isConventional:  true,
			message:         "feat: some feature description",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "commit header case is not upper",
			expectedErrors:  true,
		},
		{
			name:            "Valid Lowercase Conventional Commit",
			isConventional:  true,
			message:         "feat: some feature description",
			caseChoice:      "lower",
			expectedValid:   true,
			expectedMessage: "Header case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid Lowercase Conventional Commit",
			isConventional:  true,
			message:         "feat: Some feature description",
			caseChoice:      "lower",
			expectedValid:   false,
			expectedMessage: "commit header case is not lower",
			expectedErrors:  true,
		},
		{
			name:            "Valid Uppercase Non-Conventional Commit",
			isConventional:  false,
			message:         "Update something",
			caseChoice:      "upper",
			expectedValid:   true,
			expectedMessage: "Header case is valid",
			expectedErrors:  false,
		},
		{
			name:            "Invalid Uppercase Non-Conventional Commit",
			isConventional:  false,
			message:         "update something",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "commit header case is not upper",
			expectedErrors:  true,
		},
		{
			name:            "Invalid Case Choice",
			isConventional:  false,
			message:         "Some message",
			caseChoice:      "mixed",
			expectedValid:   false,
			expectedMessage: "invalid configured case: mixed",
			expectedErrors:  true,
		},
		{
			name:            "Empty Message",
			isConventional:  false,
			message:         "",
			caseChoice:      "upper",
			expectedValid:   false,
			expectedMessage: "empty message",
			expectedErrors:  true,
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Perform the check
			check := rule.ValidateHeaderCase(tabletest.isConventional, tabletest.message, tabletest.caseChoice)

			// Check errors
			if tabletest.expectedErrors {
				require.NotEmpty(t, check.Errors(), "Expected errors, but got none")

				// Verify the error message
				require.Equal(t, tabletest.expectedMessage, check.Message(),
					"Error message does not match expected")
			} else {
				require.Empty(t, check.Errors(), "Did not expect errors, but got some")
				require.Equal(t, "Header case is valid", check.Message(),
					"Expected default valid message")
			}

			// Check status
			require.Equal(t, "Header Case", check.Name(), "Status should always be 'Header Case'")
		})
	}
}

func TestHeaderCaseCheckMethods(t *testing.T) {
	t.Run("Status Method", func(t *testing.T) {
		check := &rule.HeaderCase{}
		require.Equal(t, "Header Case", check.Name())
	})

	t.Run("Message Method with No Errors", func(t *testing.T) {
		check := &rule.HeaderCase{}
		require.Equal(t, "Header case is valid", check.Message())
	})

	t.Run("Message Method with Errors", func(t *testing.T) {
		check := &rule.HeaderCase{
			RuleErrors: []error{
				errors.New("test error"),
				errors.New("second error"),
			},
		}
		require.Equal(t, "test error", check.Message())
	})

	t.Run("Errors Method", func(t *testing.T) {
		expectedErrors := []error{
			errors.New("test error"),
			errors.New("second error"),
		}
		check := &rule.HeaderCase{
			RuleErrors: expectedErrors,
		}
		require.Equal(t, expectedErrors, check.Errors())
	})
}
