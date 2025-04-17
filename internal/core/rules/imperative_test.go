// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImperativeVerbRule(t *testing.T) {
	testCases := []struct {
		name           string
		isConventional bool
		message        string
		expectedValid  bool
		expectedCode   string
	}{
		{
			name:           "Valid imperative verb",
			isConventional: false,
			message:        "Add new feature",
			expectedValid:  true,
		},
		{
			name:           "Valid imperative verb in conventional commit",
			isConventional: true,
			message:        "feat: Add new feature",
			expectedValid:  true,
		},
		{
			name:           "Valid imperative verb in conventional commit with scope",
			isConventional: true,
			message:        "feat(auth): Add login button",
			expectedValid:  true,
		},
		{
			name:           "Valid imperative verb in conventional commit with scope and breaking change",
			isConventional: true,
			message:        "feat(auth)!: Add OAuth support",
			expectedValid:  true,
		},
		{
			name:           "Non-imperative past tense verb",
			isConventional: false,
			message:        "Added new feature",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorPastTense),
		},
		{
			name:           "Non-imperative gerund",
			isConventional: false,
			message:        "Adding new feature",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorGerund),
		},
		{
			name:           "Non-imperative third person",
			isConventional: false,
			message:        "Adds new feature",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorThirdPerson),
		},
		{
			name:           "Empty message",
			isConventional: false,
			message:        "",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorMissingSubject),
		},
		{
			name:           "Invalid conventional commit format",
			isConventional: true,
			message:        "invalid-format",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidFormat),
		},
		{
			name:           "Invalid conventional commit - missing subject",
			isConventional: true,
			message:        "feat: ",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorMissingSubject),
		},
		{
			name:           "Unicode characters",
			isConventional: false,
			message:        "Résolve élément issue",
			expectedValid:  true,
		},
		{
			name:           "Non-imperative verb in conventional commit",
			isConventional: true,
			message:        "fix(core): Added new validation",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorPastTense),
		},
		{
			name:           "Non-verb word",
			isConventional: false,
			message:        "The new feature",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorNonVerb),
		},
		{
			name:           "Leading whitespace",
			isConventional: false,
			message:        "  \t  Fix bug",
			expectedValid:  true,
		},
		{
			name:           "Conventional commit with multi-part scope",
			isConventional: true,
			message:        "feat(user/auth,permissions): Add role-based access",
			expectedValid:  true,
		},
		{
			name:           "Imperative verb ending in 'ed'",
			isConventional: false,
			message:        "Proceed with implementation",
			expectedValid:  true,
		},
		{
			name:           "Imperative verb ending in 's'",
			isConventional: false,
			message:        "Process data correctly",
			expectedValid:  true,
		},
		{
			name:           "Conventional commit - uppercase type",
			isConventional: true,
			message:        "Feat: Add feature",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidFormat),
		},
		{
			name:           "Conventional commit - no space after colon",
			isConventional: true,
			message:        "feat:Add feature",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidFormat),
		},
		{
			name:           "Invalid non-conventional format",
			isConventional: false,
			message:        "!@#$ invalid start",
			expectedValid:  false,
			expectedCode:   string(domain.ValidationErrorInvalidFormat),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a commit info
			commitInfo := &domain.CommitInfo{
				Subject: testCase.message,
			}

			// Create rule
			rule := rules.NewImperativeVerbRule(testCase.isConventional)

			// Validate
			errors := rule.Validate(commitInfo)

			// Check errors
			if testCase.expectedValid {
				require.Empty(t, errors, "Did not expect errors")
				assert.Equal(t, "Commit begins with imperative verb", rule.Result(), "Result should indicate success")

				// Check verbose result for valid case
				if strings.Contains(testCase.message, "Add") {
					assert.Contains(t, rule.VerboseResult(), "Add", "VerboseResult should contain the verb")
				}
			} else {
				require.NotEmpty(t, errors, "Expected errors")

				// Check error code
				if testCase.expectedCode != "" {
					assert.Equal(t, testCase.expectedCode, errors[0].Code, "Error code should match expected")
				}

				// Check result message
				assert.Equal(t, "Non-imperative verb detected", rule.Result(), "Result should indicate error")

				// Verify rule name is set in the error
				assert.Equal(t, "ImperativeVerb", errors[0].Rule, "Rule name should be set in ValidationError")

				// Test VerboseResult for error cases includes helpful info
				assert.NotEmpty(t, rule.VerboseResult(), "VerboseResult should not be empty")
			}

			// Check name method
			assert.Equal(t, "ImperativeVerb", rule.Name(), "Name should be 'ImperativeVerb'")

			// Check help method is not empty
			assert.NotEmpty(t, rule.Help(), "Help text should not be empty")
		})
	}
}

func TestImperativeVerbRuleOptions(t *testing.T) {
	t.Run("Custom non-imperative starters", func(t *testing.T) {
		// Create a custom rule with non-imperative starters
		customWords := map[string]bool{"commit": true}
		rule := rules.NewImperativeVerbRule(false,
			rules.WithCustomNonImperativeStarters(customWords))

		// Test with a message that would normally be valid
		commit := &domain.CommitInfo{Subject: "Commit changes"}
		errors := rule.Validate(commit)

		// Should be marked as non-verb due to our custom setting
		require.NotEmpty(t, errors, "Should reject custom non-verb")
		assert.Equal(t, string(domain.ValidationErrorNonVerb), errors[0].Code, "Error code should be ValidationErrorNonVerb")
	})

	t.Run("Custom base forms with ED", func(t *testing.T) {
		// Add a custom word ending in 'ed' that should be considered valid
		customEdForms := map[string]bool{"crated": true}
		rule := rules.NewImperativeVerbRule(false,
			rules.WithAdditionalBaseFormsEndingWithED(customEdForms))

		// Test with custom word
		commit := &domain.CommitInfo{Subject: "Crated package"}
		errors := rule.Validate(commit)

		// Should accept our custom word
		assert.Empty(t, errors, "Should accept custom ED form")

		// Test default behavior for comparison
		defaultRule := rules.NewImperativeVerbRule(false)
		defaultErrors := defaultRule.Validate(commit)
		require.NotEmpty(t, defaultErrors, "Default should reject 'crated'")
	})

	t.Run("Multiple options combined", func(t *testing.T) {
		// Combine multiple options
		rule := rules.NewImperativeVerbRule(true,
			rules.WithCustomNonImperativeStarters(map[string]bool{"execute": true}),
			rules.WithAdditionalBaseFormsEndingWithS(map[string]bool{"canvas": true}))

		// Test with valid conventional commit using a word that's custom allowed
		commit := &domain.CommitInfo{Subject: "feat(ui): Canvas button"}
		errors := rule.Validate(commit)
		assert.Empty(t, errors, "Should pass with multiple options")

		// Test with a word we've marked as non-verb
		badCommit := &domain.CommitInfo{Subject: "feat(api): Execute function"}
		badErrors := rule.Validate(badCommit)
		require.NotEmpty(t, badErrors, "Should fail with custom non-verb")
		assert.Equal(t, string(domain.ValidationErrorNonVerb), badErrors[0].Code, "Should have correct error code")
	})
}

func TestDifficultVerbCases(t *testing.T) {
	// Test various edge cases for verb detection
	validCases := []string{
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

	for _, message := range validCases {
		t.Run("Valid: "+message, func(t *testing.T) {
			rule := rules.NewImperativeVerbRule(false)
			commit := &domain.CommitInfo{Subject: message}
			errors := rule.Validate(commit)
			assert.Empty(t, errors, "Should accept valid imperative verb: "+message)
		})
	}

	invalidCases := []string{
		"Tests feature",   // 3rd person
		"Testing feature", // Gerund
		"Tested feature",  // Past tense
		"A new feature",   // Article
	}

	for _, message := range invalidCases {
		t.Run("Invalid: "+message, func(t *testing.T) {
			rule := rules.NewImperativeVerbRule(false)
			commit := &domain.CommitInfo{Subject: message}
			errors := rule.Validate(commit)
			assert.NotEmpty(t, errors, "Should reject non-imperative form: "+message)
		})
	}
}
