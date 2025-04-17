// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSubjectCaseRule(t *testing.T) {
	testCases := []struct {
		name           string
		isConventional bool
		message        string
		caseChoice     string
		allowNonAlpha  bool
		expectedValid  bool
		expectedCode   string
	}{
		{
			name:           "Valid uppercase conventional commit",
			isConventional: true,
			message:        "feat: Add new feature",
			caseChoice:     "upper",
			expectedValid:  true,
		},
		{
			name:           "Invalid uppercase conventional commit",
			isConventional: true,
			message:        "feat: add new feature",
			caseChoice:     "upper",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidCase),
		},
		{
			name:           "Valid lowercase conventional commit",
			isConventional: true,
			message:        "feat: add new feature",
			caseChoice:     "lower",
			expectedValid:  true,
		},
		{
			name:           "Invalid lowercase conventional commit",
			isConventional: true,
			message:        "feat: Add new feature",
			caseChoice:     "lower",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidCase),
		},
		{
			name:           "Valid uppercase non-conventional commit",
			isConventional: false,
			message:        "Add new feature",
			caseChoice:     "upper",
			expectedValid:  true,
		},
		{
			name:           "Invalid uppercase non-conventional commit",
			isConventional: false,
			message:        "add new feature",
			caseChoice:     "upper",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidCase),
		},
		{
			name:           "Invalid case choice fallbacks to lower",
			isConventional: false,
			message:        "Add new feature",
			caseChoice:     "invalid",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidCase),
		},
		{
			name:           "Empty message",
			isConventional: false,
			message:        "",
			caseChoice:     "lower",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorEmptyDescription),
		},
		{
			name:           "Invalid conventional commit format",
			isConventional: true,
			message:        "invalid format",
			caseChoice:     "lower",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidFormat),
		},
		{
			name:           "With scope",
			isConventional: true,
			message:        "feat(auth): add login button",
			caseChoice:     "lower",
			expectedValid:  true,
		},
		{
			name:           "Allow non-alpha characters with option",
			isConventional: false,
			message:        "123 numbers first",
			caseChoice:     "lower",
			allowNonAlpha:  true,
			expectedValid:  true,
		},
		{
			name:           "Ignore case option",
			isConventional: false,
			message:        "Any Case Is OK",
			caseChoice:     "ignore",
			expectedValid:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create options
			var options []rules.SubjectCaseOption

			// Add case choice
			if testCase.caseChoice != "" {
				options = append(options, rules.WithCaseChoice(testCase.caseChoice))
			}

			// Configure conventional if needed
			if testCase.isConventional {
				options = append(options, rules.WithConventionalCommitCase())
			}

			// Configure allow non-alpha if needed
			if testCase.allowNonAlpha {
				options = append(options, rules.WithAllowNonAlphaCase())
			}

			// Create the rule
			rule := rules.NewSubjectCaseRule(options...)

			// Create a commit for validation
			commit := &domain.CommitInfo{
				Subject: testCase.message,
			}

			// Validate
			result := rule.Validate(commit)

			// Check validity
			if testCase.expectedValid {
				assert.Empty(t, result, "Expected no validation errors")
				assert.Equal(t, "Valid subject case", rule.Result(), "Result should indicate valid case")
			} else {
				assert.NotEmpty(t, result, "Expected validation errors")
				assert.Equal(t, "Invalid subject case", rule.Result(), "Result should indicate invalid case")

				// Verify error code if expected
				if testCase.expectedCode != "" {
					assert.Equal(t, testCase.expectedCode, result[0].Code, "Error code should match expected")
				}

				// Check rule name is set
				assert.Equal(t, "SubjectCase", result[0].Rule, "Rule name should be set in ValidationError")

				// Check verbose result for expected content
				verboseResult := rule.VerboseResult()

				//nolint:exhaustive
				switch domain.ValidationErrorCode(result[0].Code) {
				case domain.ValidationErrorEmptyDescription:
					assert.Contains(t, verboseResult, "empty", "VerboseResult should explain empty subject")
				case domain.ValidationErrorInvalidFormat:
					assert.Contains(t, verboseResult, "Invalid conventional commit format",
						"VerboseResult should explain format issue")
				case domain.ValidationErrorInvalidCase:
					// Different messages based on case choice
					isLowerCaseTest := testCase.name == "Invalid lowercase conventional commit" ||
						testCase.name == "Invalid case choice fallbacks to lower"

					if isLowerCaseTest {
						assert.Contains(t, verboseResult, "lowercase",
							"VerboseResult should explain lowercase requirement")
					} else {
						assert.Contains(t, verboseResult, "uppercase",
							"VerboseResult should explain uppercase requirement")
					}
				}

				// Check help text
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				// Check context for case validation errors
				if testCase.expectedCode == string(domain.ValidationErrorInvalidCase) {
					assert.Contains(t, result[0].Context, "expected_case", "Context should include expected case")
					assert.Contains(t, result[0].Context, "actual_case", "Context should include actual case")
				}
			}

			// Name should always return SubjectCase
			assert.Equal(t, "SubjectCase", rule.Name(), "Name should be 'SubjectCase'")
		})
	}
}

func TestSubjectCaseHelpMessages(t *testing.T) {
	testCases := []struct {
		name            string
		isConventional  bool
		message         string
		caseChoice      string
		expectedCode    string
		expectedContent string
	}{
		{
			name:            "Help for empty subject",
			isConventional:  false,
			message:         "",
			caseChoice:      "lower",
			expectedCode:    string(domain.ValidationErrorEmptyDescription),
			expectedContent: "non-empty commit message",
		},
		{
			name:            "Help for invalid conventional format",
			isConventional:  true,
			message:         "bad-format",
			caseChoice:      "lower",
			expectedCode:    string(domain.ValidationErrorInvalidFormat),
			expectedContent: "Conventional Commits specification",
		},
		{
			name:            "Help for wrong case - upper",
			isConventional:  false,
			message:         "lowercase when uppercase required",
			caseChoice:      "upper",
			expectedCode:    string(domain.ValidationErrorInvalidCase),
			expectedContent: "Capitalize the first letter",
		},
		{
			name:            "Help for wrong case - lower",
			isConventional:  false,
			message:         "Uppercase when lowercase required",
			caseChoice:      "lower",
			expectedCode:    string(domain.ValidationErrorInvalidCase),
			expectedContent: "Use lowercase for the first letter",
		},
		{
			name:            "No errors to fix",
			isConventional:  false,
			message:         "valid message",
			caseChoice:      "lower",
			expectedContent: "No errors to fix",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create options
			var options []rules.SubjectCaseOption

			// Configure based on test case
			options = append(options, rules.WithCaseChoice(testCase.caseChoice))
			if testCase.isConventional {
				options = append(options, rules.WithConventionalCommitCase())
			}

			// Create rule
			rule := rules.NewSubjectCaseRule(options...)

			// Create commit and validate
			commit := &domain.CommitInfo{
				Subject: testCase.message,
			}
			_ = rule.Validate(commit)

			// Check help message
			helpText := rule.Help()
			assert.Contains(t, helpText, testCase.expectedContent, "Help should contain expected content")
		})
	}
}

func TestSubjectCaseErrors(t *testing.T) {
	// Test that Errors() returns the errors slice
	rule := rules.NewSubjectCaseRule()
	assert.Empty(t, rule.Errors(), "New rule should have no errors")

	// Validate with invalid subject to generate errors
	_ = rule.Validate(&domain.CommitInfo{Subject: ""})
	assert.NotEmpty(t, rule.Errors(), "Errors() should return validation errors after validation")

	// Errors should be the same as those returned by Validate
	errors := rule.Validate(&domain.CommitInfo{Subject: ""})
	assert.Equal(t, errors, rule.Errors(), "Errors() should match Validate's return value")
}
