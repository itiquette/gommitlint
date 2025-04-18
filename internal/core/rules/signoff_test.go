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
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignOffRule tests the sign-off validation logic.
func TestSignOffRule(t *testing.T) {
	// Standard sign-off format for tests
	const validSignOff = "Signed-off-by: Dev Eloper <dev@example.com>"

	tests := []struct {
		name           string
		message        string
		requireSignOff bool
		allowMultiple  bool
		expectedValid  bool
		expectedCode   string
		errorContains  string
	}{
		{
			name: "Valid sign-off",
			message: `Add feature

This is a detailed description of the feature.

` + validSignOff,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name:           "Valid sign-off with crlf",
			message:        "Add feature\r\n\r\nThis is a description.\r\n\r\n" + validSignOff,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Valid sign-off with multiple signers (allowed)",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Valid sign-off with multiple signers (not allowed)",
			message: `Fix bug
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

This is a detailed description of the feature.`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "missing_signoff",
			errorContains:  "Missing Signed-off-by line",
		},
		{
			name: "Missing sign-off (not required)",
			message: `Add feature

This is a detailed description of the feature.`,
			requireSignOff: false,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Malformed sign-off - wrong format",
			message: `Add feature

This is a detailed description of the feature.

Signed by: Dev Eloper <dev@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "invalid_format",
			errorContains:  "Invalid sign-off format",
		},
		{
			name: "Malformed sign-off - invalid email",
			message: `Add feature

This is a detailed description of the feature.

Signed-off-by: Dev Eloper <dev@example>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "invalid_format",
			errorContains:  "Invalid sign-off format",
		},
		{
			name: "Malformed sign-off - missing name",
			message: `Add feature

This is a detailed description of the feature.

Signed-off-by: <dev@example.com>`,
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
			message:        "   \t\n  ",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "empty_message",
			errorContains:  "commit message body is empty",
		},
		{
			name: "Custom regex - valid",
			message: `Add feature

This is a detailed description.

Developer Certificate: John Doe <john@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
		},
		{
			name: "Custom regex - invalid",
			message: `Add feature

This is a detailed description.`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			expectedCode:   "missing_signoff",
			errorContains:  "Missing Signed-off-by line",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Build options
			var options []rules.SignOffOption

			// Configure based on test case
			if !testCase.requireSignOff {
				options = append(options, rules.WithRequireSignOff(false))
			}

			if testCase.allowMultiple {
				options = append(options, rules.WithAllowMultipleSignOffs(true))
			} else {
				options = append(options, rules.WithAllowMultipleSignOffs(false))
			}

			// Use a custom regex for one test to ensure that works too
			if testCase.name == "Custom regex - valid" || testCase.name == "Custom regex - invalid" {
				options = append(options, rules.WithCustomSignOffRegex(regexp.MustCompile(`Developer Certificate: .+ <.+@.+\..+>`)))
			}

			// Create rule with options
			rule := rules.NewSignOffRule(options...)

			// Create commit for testing
			commit := &domain.CommitInfo{
				Body: testCase.message,
			}

			// Execute validation
			errors := rule.Validate(commit)

			// Check validity
			if testCase.expectedValid {
				assert.Empty(t, errors, "Expected no validation errors")
				assert.Equal(t, "Sign-off is present", rule.Result(), "Expected success message")

				// Different checks for sign-off not required vs found sign-off
				if !testCase.requireSignOff {
					assert.Contains(t, rule.VerboseResult(), "not required by configuration", "Verbose result should indicate sign-off not required")
				} else {
					assert.Contains(t, rule.VerboseResult(), "Valid Developer Certificate of Origin sign-off found", "Verbose result should indicate valid sign-off")
				}

				assert.Contains(t, rule.Help(), "No errors to fix", "Help for valid message should indicate nothing to fix")
			} else {
				assert.NotEmpty(t, errors, "Expected errors but found none")

				// Check error code if specified
				if testCase.expectedCode != "" {
					// Check if the original_code in context matches the expected code
					assert.Equal(t, testCase.expectedCode, errors[0].Context["original_code"],
						"Original error code in context should match expected")
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

					if !found && testCase.expectedCode == "missing_signoff" {
						// Special case for missing_signoff since we've changed the error message
						assert.Contains(t, errors[0].Error(), "Missing Signed-off-by line",
							"Error should mention missing sign-off")
					} else {
						require.True(t, found, "Expected error containing %q", testCase.errorContains)
					}
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "SignOff", errors[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify Help() method provides guidance
				helpText := rule.Help()
				assert.NotEmpty(t, helpText, "Help text should not be empty")

				// Test specific help messages based on error code
				if errors[0].Context["original_code"] == "empty_message" {
					assert.Contains(t, helpText, "currently empty", "Help should mention empty message")
				} else if errors[0].Context["original_code"] == "missing_signoff" {
					assert.Contains(t, helpText, "git commit -s", "Help should mention git commit -s")
				} else if errors[0].Context["original_code"] == "invalid_format" {
					assert.Contains(t, helpText, "correctly formatted", "Help should mention correct format")
				} else if errors[0].Context["original_code"] == "multiple_signoffs" {
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
			rules.WithAllowMultipleSignOffs(false),
		)

		commit := &domain.CommitInfo{
			Body: multiSignoffMessage,
		}

		errors := rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors for multiple signoffs")
		assert.Equal(t, string(appErrors.ErrMissingSignoff), errors[0].Code)
		assert.Equal(t, "multiple_signoffs", errors[0].Context["original_code"])
		assert.Contains(t, errors[0].Context, "signoff_count")
		assert.Equal(t, "2", errors[0].Context["signoff_count"])

		// Test missing signoff context
		missingSignoffMessage := "Add feature without signoff"

		commit = &domain.CommitInfo{
			Body: missingSignoffMessage,
		}

		errors = rule.Validate(commit)
		require.NotEmpty(t, errors, "Should have errors for missing signoff")
		assert.Equal(t, string(appErrors.ErrMissingSignoff), errors[0].Code)
		assert.Equal(t, "missing_signoff", errors[0].Context["original_code"])
		assert.Contains(t, errors[0].Context, "message")
	})
}
