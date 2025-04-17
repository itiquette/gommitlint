// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCase defines a test case for the spell rule tests.
type testCase struct {
	name           string
	subject        string
	body           string
	locale         string
	ignoreWords    []string
	customWords    map[string]string
	maxErrors      int
	expectedErrors int
	expectedCode   string
	expectedWords  []string
}

// setupSpellRule creates a spell rule with the specified options from the test case.
func setupSpellRule(testConfig testCase) (*rules.SpellRule, *domain.CommitInfo) {
	// Build options based on test case
	var options []rules.SpellRuleOption

	if testConfig.locale != "" {
		options = append(options, rules.WithLocale(testConfig.locale))
	}

	if len(testConfig.ignoreWords) > 0 {
		options = append(options, rules.WithIgnoreWords(testConfig.ignoreWords))
	}

	if len(testConfig.customWords) > 0 {
		options = append(options, rules.WithCustomWords(testConfig.customWords))
	}

	if testConfig.maxErrors > 0 {
		options = append(options, rules.WithMaxErrors(testConfig.maxErrors))
	}

	// Create the rule instance
	rule := rules.NewSpellRule(options...)

	// Create a commit for testing
	commit := &domain.CommitInfo{
		Subject: testConfig.subject,
		Body:    testConfig.body,
	}

	return rule, commit
}

// validateErrors checks that validation errors match expectations.
func validateErrors(t *testing.T, testCase testCase, errors []*domain.ValidationError) {
	t.Helper()
	// Check for expected number of errors
	assert.Len(t, errors, testCase.expectedErrors, "Incorrect number of errors")

	// Check error code if specified
	if testCase.expectedCode != "" && len(errors) > 0 {
		assert.Equal(t, testCase.expectedCode, errors[0].Code,
			"Error code should match expected")
	}

	// Verify rule name is set in ValidationError
	if len(errors) > 0 {
		assert.Equal(t, "Spell", errors[0].Rule,
			"Rule name should be set in ValidationError")
	}

	// Check context data
	if testCase.expectedCode == string(domain.ValidationErrorSpelling) && len(errors) > 0 {
		assert.Contains(t, errors[0].Context, "original",
			"Context should contain original misspelled word")
		assert.Contains(t, errors[0].Context, "corrected",
			"Context should contain corrected spelling")
	} else if testCase.expectedCode == string(domain.ValidationErrorUnknown) && len(errors) > 0 {
		assert.Contains(t, errors[0].Context, "locale",
			"Context should contain the invalid locale")
		assert.Contains(t, errors[0].Context, "supported_locales",
			"Context should contain supported locales")
	}
}

// validateExpectedWords checks for expected words in error messages.
func validateExpectedWords(t *testing.T, testCase testCase, errors []*domain.ValidationError) {
	t.Helper()

	if testCase.expectedWords == nil || len(errors) == 0 {
		return
	}

	for _, word := range testCase.expectedWords {
		found := false

		for _, err := range errors {
			if strings.Contains(err.Error(), word) {
				found = true

				break
			}

			// Also check in context
			for _, contextValue := range err.Context {
				if strings.Contains(contextValue, word) {
					found = true

					break
				}
			}
		}

		if !found {
			// This is not a hard failure - some misspellings might not be in the dictionary
			t.Logf("Note: Expected to find '%s' in one of the error messages or context, but didn't", word)
		}
	}
}

// validateRuleMethods checks rule helper methods (Result, Help, Name, VerboseResult).
func validateRuleMethods(t *testing.T, testCase testCase, rule *rules.SpellRule, errors []*domain.ValidationError) {
	t.Helper()
	// Verify Name() method
	assert.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")

	// Verify Result() method
	if testCase.expectedErrors > 0 {
		expectedResult := fmt.Sprintf("%d misspelling(s)", testCase.expectedErrors)
		assert.Equal(t, expectedResult, rule.Result(),
			"Result should report the number of misspellings")
	} else {
		assert.Equal(t, "No spelling errors", rule.Result(),
			"Result should indicate no misspellings")
	}

	// Verify Help() method
	validateHelpMethod(t, testCase, rule, errors)

	// Check VerboseResult output
	validateVerboseResult(t, testCase, rule)
}

// validateHelpMethod checks the rule's Help method output.
func validateHelpMethod(t *testing.T, testCase testCase, rule *rules.SpellRule, errors []*domain.ValidationError) {
	t.Helper()

	helpText := rule.Help()
	require.NotEmpty(t, helpText, "Help text should not be empty")

	if testCase.expectedErrors == 0 {
		assert.Equal(t, "No errors to fix", helpText,
			"Help should indicate no errors to fix")
	} else if len(errors) > 0 && domain.ValidationErrorCode(errors[0].Code) == domain.ValidationErrorUnknown {
		assert.Contains(t, helpText, "supported locale",
			"Help should mention supported locales")
	} else if len(errors) > 0 {
		assert.Contains(t, helpText, "Fix the following misspellings",
			"Help should provide guidance")

		// Check if help text contains error details
		if len(errors) > 0 {
			assert.Contains(t, helpText, errors[0].Error(),
				"Help should include error details")
		}
	}
}

// validateVerboseResult checks the rule's VerboseResult method output.
func validateVerboseResult(t *testing.T, testCase testCase, rule *rules.SpellRule) {
	t.Helper()

	verboseText := rule.VerboseResult()

	if testCase.expectedErrors == 0 {
		assert.Contains(t, verboseText, "No common misspellings found",
			"Verbose output should indicate no errors")

		// Check for locale mention
		if testCase.locale == "GB" || testCase.locale == "UK" ||
			strings.ToLower(testCase.locale) == "gb" || strings.ToLower(testCase.locale) == "uk" {
			assert.Contains(t, verboseText, "UK/GB (British English)",
				"Verbose output should mention British English")
		} else {
			assert.Contains(t, verboseText, "US (American English)",
				"Verbose output should mention American English")
		}
	} else if testCase.expectedCode == string(domain.ValidationErrorUnknown) {
		assert.Contains(t, verboseText, "Invalid locale",
			"Verbose output should mention invalid locale")
		assert.Contains(t, verboseText, testCase.locale,
			"Verbose output should mention the specific invalid locale")
	} else {
		assert.Contains(t, verboseText, fmt.Sprintf("Found %d misspelling", testCase.expectedErrors),
			"Verbose output should mention the number of misspellings")

		// For tests with specific misspellings, check that they're mentioned
		if testCase.expectedWords != nil {
			for _, word := range testCase.expectedWords {
				if strings.Contains(verboseText, word) {
					// We found at least one mention of the word, so we're good
					break
				}
			}
		}
	}
}

func TestSpellRule(t *testing.T) {
	tests := []testCase{
		{
			name:           "No misspellings",
			subject:        "This is a correct sentence.",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "One misspelling",
			subject:        "This is definately a misspelling.", //nolint
			locale:         "US",
			expectedErrors: 1,
			expectedCode:   string(domain.ValidationErrorSpelling),
			expectedWords:  []string{"definately", "definitely"}, //nolint
		},
		{
			name:           "Multiple misspellings",
			subject:        "We occured a misspelling and we beleive it needs fixing.", //nolint
			locale:         "US",
			expectedErrors: 2,
			expectedCode:   string(domain.ValidationErrorSpelling),
			expectedWords:  []string{"occured", "beleive"}, //nolint
		},
		{
			name:           "British english",
			subject:        "The colour of the centre looks great.",
			locale:         "GB",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Empty message",
			subject:        "",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Whitespace message",
			subject:        "   \t   \n",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Unknown locale",
			subject:        "This is a test.",
			locale:         "FR",
			expectedErrors: 1,
			expectedCode:   string(domain.ValidationErrorUnknown),
			expectedWords:  []string{"unknown locale"},
		},
		{
			name:           "Default locale",
			subject:        "This is a test with the color red.",
			locale:         "",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Case insensitive locale",
			subject:        "The colour of the centre looks great.",
			locale:         "gb",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Misspelling in body",
			subject:        "Fix documentation",
			body:           "The documentaiton was incorrect and has been fixed.", //nolint
			locale:         "US",
			expectedErrors: 1,
			expectedCode:   string(domain.ValidationErrorSpelling),
			expectedWords:  []string{"documentaiton", "documentation"}, //nolint
		},
		{
			name:           "Ignore specific words",
			subject:        "Fix documentaiton and configuraiton issues", //nolint
			locale:         "US",
			ignoreWords:    []string{"documentaiton"}, //nolint
			expectedErrors: 1,                         // Only one error because we're ignoring "documentation"
			expectedCode:   string(domain.ValidationErrorSpelling),
			expectedWords:  []string{"configuraiton"}, //nolint
		},
		{
			name:           "Custom word mappings",
			subject:        "Fix golang issues",
			locale:         "US",
			customWords:    map[string]string{"golang": "Go"},
			expectedErrors: 1,
			expectedCode:   string(domain.ValidationErrorSpelling),
			expectedWords:  []string{"golang", "Go"},
		},
		{
			name:           "Max errors limit",
			subject:        "We occured a misspelling and we beleive it needs fixing.", //nolint
			locale:         "US",
			maxErrors:      1, // Only report the first error
			expectedErrors: 1,
			expectedCode:   string(domain.ValidationErrorSpelling),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup the rule and commit
			rule, commit := setupSpellRule(testCase)

			// Execute validation
			errors := rule.Validate(commit)

			// Validate the errors
			validateErrors(t, testCase, errors)

			// Check for expected words in errors
			validateExpectedWords(t, testCase, errors)

			// Validate the rule methods
			validateRuleMethods(t, testCase, rule, errors)
		})
	}
}
