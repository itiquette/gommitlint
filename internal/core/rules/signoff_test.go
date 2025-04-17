// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignOffRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		requireSignOff bool
		allowMultiple  bool
		customRegex    *regexp.Regexp
		expectedValid  bool
		expectedCode   string
		errorContains  string
	}{
		{
			name: "Valid sign-off",
			message: `Add new feature
Implement automatic logging system.
Signed-off-by: Laval Lion <laval.lion@cavora.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name:           "Valid sign-off with crlf",
			message:        "Update docs\r\n\r\nImprove README\r\n\r\nSigned-off-by: Cragger Crocodile <cragger@svamp.org>",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Valid sign-off with multiple signers (allowed)",
			message: `Fix bug
Update error handling.
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Valid sign-off with multiple signers (not allowed)",
			message: `Fix bug
Update error handling.
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  false,
			expectedValid:  false,
			expectedCode:   "multiple_signoffs",
			errorContains:  "multiple sign-offs",
		},
		{
			name: "Missing sign-off (required)",
			message: `Add feature
Implement new logging system.`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "missing_signoff",
			errorContains:  "Missing Signed-off-by line",
		},
		{
			name: "Missing sign-off (not required)",
			message: `Add feature
Implement new logging system.`,
			requireSignOff: false,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Malformed sign-off - wrong format",
			message: `Add test
Signed by: Laval Lion <laval.lion@cavora.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "invalid_format",
			errorContains:  "Invalid sign-off format",
		},
		{
			name: "Malformed sign-off - invalid email",
			message: `Add test
Signed-off-by: Phoenix Fire <invalid-email>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "invalid_format",
			errorContains:  "Invalid sign-off format",
		},
		{
			name: "Malformed sign-off - missing name",
			message: `Add test
Signed-off-by: <laval@cavora.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "invalid_format",
			errorContains:  "Invalid sign-off format",
		},
		{
			name:           "Empty message",
			message:        "",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "empty_message",
			errorContains:  "commit message body is empty",
		},
		{
			name:           "Whitespace only message",
			message:        "   \n  \t  \n",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "empty_message",
			errorContains:  "commit message body is empty",
		},
		{
			name: "Custom regex - valid",
			message: `Add feature
Author: Test User <test@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			customRegex:    regexp.MustCompile(`^Author: ([^<]+) <([^<>@]+@[^<>]+)>$`),
			expectedValid:  true,
		},
		{
			name: "Custom regex - invalid",
			message: `Add feature
Wrong Format: Test User <test@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			customRegex:    regexp.MustCompile(`^Author: ([^<]+) <([^<>@]+@[^<>]+)>$`),
			expectedValid:  false,
			expectedCode:   "missing_signoff",
			errorContains:  "Missing Signed-off-by line",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Build options based on test case
			var options []rules.SignOffOption

			options = append(options, rules.WithRequireSignOff(testCase.requireSignOff))
			options = append(options, rules.WithAllowMultipleSignOffs(testCase.allowMultiple))

			if testCase.customRegex != nil {
				options = append(options, rules.WithCustomSignOffRegex(testCase.customRegex))
			}

			// Create the rule instance
			rule := rules.NewSignOffRule(options...)

			// Create a commit for testing
			commit := &domain.CommitInfo{
				Body: testCase.message,
			}

			// Execute validation
			errors := rule.Validate(commit)

			// Check for expected validation result
			if testCase.expectedValid {
				assert.Empty(t, errors, "Expected no errors but got: %v", errors)

				// Verify Result() and VerboseResult() methods return expected messages
				assert.Equal(t, "Sign-off exists", rule.Result(), "Expected default valid message")

				if strings.Contains(testCase.message, "Signed-off-by") ||
					(testCase.customRegex != nil && strings.Contains(testCase.message, "Author:")) {
					assert.Contains(t, rule.VerboseResult(), "Valid", "Verbose result should indicate valid signature")
				}

				// Test Help on valid case
				assert.Equal(t, "No errors to fix", rule.Help(), "Help for valid message should indicate nothing to fix")
			} else {
				assert.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" {
					assert.Equal(t, testCase.expectedCode, errors[0].Code,
						"Error code should match expected")
				}

				// Check error message contains expected substring
				if testCase.errorContains != "" {
					found := false

					for _, err := range errors {
						if strings.Contains(err.Error(), testCase.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error containing %q", testCase.errorContains)
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "SignOff", errors[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help() method provides guidance
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				// Test specific help messages based on error code
				if errors[0].Code == "empty_message" {
					assert.Contains(t, helpText, "currently empty", "Help should mention empty message")
				} else if errors[0].Code == "missing_signoff" {
					assert.Contains(t, helpText, "git commit -s", "Help should mention git commit -s")
				} else if errors[0].Code == "invalid_format" {
					assert.Contains(t, helpText, "correctly formatted", "Help should mention correct format")
				} else if errors[0].Code == "multiple_signoffs" {
					assert.Contains(t, helpText, "multiple", "Help should mention multiple sign-offs issue")
				}

				// Verify Result() method returns expected message
				assert.Equal(t, "Missing sign-off", rule.Result(), "Expected error result message")
				assert.NotEqual(t, rule.Result(), rule.VerboseResult(), "Verbose result should be different from regular result")
			}

			// Verify Name() method
			assert.Equal(t, "SignOff", rule.Name(), "Name should be 'SignOff'")
		})
	}

	// Test specific error context information
	t.Run("Error context information", func(t *testing.T) {
		// Test multiple signoffs context
		multiSignoffMessage := `Fix bug
Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev Two <dev2@example.com>`

		rule := rules.NewSignOffRule(
			rules.WithRequireSignOff(true),
			rules.WithAllowMultipleSignOffs(false),
		)

		commit := &domain.CommitInfo{
			Body: multiSignoffMessage,
		}

		errors := rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors for multiple signoffs")
		assert.Equal(t, "multiple_signoffs", errors[0].Code)
		assert.Contains(t, errors[0].Context, "signoff_count")
		assert.Equal(t, "2", errors[0].Context["signoff_count"])

		// Test missing signoff context
		missingSignoffMessage := "Add feature without signoff"

		commit = &domain.CommitInfo{
			Body: missingSignoffMessage,
		}

		errors = rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors for missing signoff")
		assert.Equal(t, "missing_signoff", errors[0].Code)
		assert.Contains(t, errors[0].Context, "message")
	})
}
