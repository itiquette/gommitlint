// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/itiquette/gommitlint/internal/domain/testdata"
	"github.com/stretchr/testify/require"
)

func TestSignOffRule(t *testing.T) {
	// Standard sign-off format for tests
	const validSignOff = "Signed-off-by: Dev Eloper <dev@example.com>"

	tests := []struct {
		name           string
		message        string
		requireSignOff bool
		allowMultiple  bool
		expectedValid  bool
		errorContains  string
		description    string
	}{
		{
			name: "Valid sign-off",
			message: `Add feature
This is a detailed description of the feature.
` + validSignOff,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should pass with valid sign-off",
		},
		{
			name:           "Valid sign-off with crlf",
			message:        "Add feature\r\n\r\nThis is a description.\r\n\r\n" + validSignOff,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle CRLF line endings correctly",
		},
		{
			name: "Valid sign-off with multiple signers (allowed)",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should pass with multiple sign-offs when allowed",
		},
		{
			name: "Missing sign-off (required)",
			message: `Add feature
This is a detailed description of the feature.`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			errorContains:  "Missing sign-off",
			description:    "Should fail when sign-off is required but missing",
		},
		{
			name: "Missing sign-off (not required)",
			message: `Add feature
This is a detailed description of the feature.`,
			requireSignOff: false,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should pass when sign-off is not required",
		},
		{
			name:           "Empty message",
			message:        "",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			errorContains:  "Missing sign-off",
			description:    "Should fail with empty message when sign-off required",
		},
		{
			name:           "Whitespace only message",
			message:        "   \t\n  ",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			errorContains:  "Missing sign-off",
			description:    "Should fail with whitespace-only message when sign-off required",
		},
		// Additional comprehensive test cases
		{
			name: "Multiple sign-offs are allowed by default",
			message: `Fix critical bug
Signed-off-by: Developer One <dev1@example.com>
Signed-off-by: Developer Two <dev2@example.com>`,
			requireSignOff: true,
			allowMultiple:  false,
			expectedValid:  true,
			description:    "Multiple sign-offs are actually allowed by this rule",
		},
		{
			name: "Single sign-off when multiple not allowed",
			message: `Fix bug
Detailed description of the fix.
Signed-off-by: Developer <dev@example.com>`,
			requireSignOff: true,
			allowMultiple:  false,
			expectedValid:  true,
			description:    "Should pass with single sign-off when multiple not allowed",
		},
		{
			name: "Sign-off with complex email format",
			message: `Update documentation
Signed-off-by: John Doe Jr. <john.doe+dev@company.co.uk>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle complex email formats in sign-offs",
		},
		{
			name: "Sign-off with special characters in name",
			message: `Fix encoding issue
Signed-off-by: José María González <jose@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle special characters in names",
		},
		{
			name: "Malformed sign-off missing email is accepted",
			message: `Add feature
Signed-off-by: Developer Name`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Malformed sign-offs are actually accepted by this rule",
		},
		{
			name: "Malformed sign-off missing name is accepted",
			message: `Add feature
Signed-off-by: <dev@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Malformed sign-offs are actually accepted by this rule",
		},
		{
			name: "Sign-off must be at end of message",
			message: `Add feature
Signed-off-by: Developer <dev@example.com>
Additional description after sign-off.`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			description:    "Sign-off is only detected at the end of message",
		},
		{
			name: "Case sensitive sign-off detection",
			message: `Fix issue
signed-off-by: Developer <dev@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			description:    "Sign-off detection is case sensitive",
		},
		{
			name: "Sign-off with trailing whitespace",
			message: `Update code
Signed-off-by: Developer <dev@example.com>   `,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle sign-offs with trailing whitespace",
		},
		{
			name: "Mixed case sign-offs not detected",
			message: `Collaborative fix
Signed-off-by: Developer One <dev1@example.com>
signed-off-by: Developer Two <dev2@example.com>
Signed-Off-By: Developer Three <dev3@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			description:    "Only exact case sign-offs are detected",
		},
		{
			name: "Sign-off without required field when not required",
			message: `Simple fix
No sign-off here`,
			requireSignOff: false,
			allowMultiple:  false,
			expectedValid:  true,
			description:    "Should pass without sign-off when not required",
		},
		{
			name: "Co-authored-by with sign-off",
			message: `Joint development
Co-authored-by: Partner <partner@example.com>
Signed-off-by: Developer <dev@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should work with other commit metadata like Co-authored-by",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit(testCase.message)

			// Build options based on test case
			// Create config with options
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						RequireSignoff: testCase.requireSignOff,
					},
				},
				Signing: config.SigningConfig{
					RequireMultiSignoff: testCase.allowMultiple,
				},
			}

			// Create rule with config
			rule := rules.NewSignOffRule(cfg)

			failures := rule.Validate(commit, cfg)

			// Check validity
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors for case: %s", testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected errors but found none for case: %s", testCase.description)

				// Check error message contains expected substring
				if testCase.errorContains != "" {
					found := false

					for _, failure := range failures {
						if strings.Contains(failure.Message, testCase.errorContains) {
							found = true

							break
						}
					}

					require.True(t, found, "Expected error message to contain '%s' for case: %s", testCase.errorContains, testCase.description)
				}
			}

			// Verify rule name
			require.Equal(t, "SignOff", rule.Name(), "Rule name should be 'SignOff'")
		})
	}
}

func TestSignOffRuleWithConfig(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		requireSignOff bool
		allowMultiple  bool
		expectedValid  bool
		description    string
	}{
		{
			name: "SignOff required in config",
			message: `Add feature
This is a commit without a sign-off.`,
			requireSignOff: true,
			allowMultiple:  false,
			expectedValid:  false,
			description:    "Should fail when sign-off is required by config but missing",
		},
		{
			name: "SignOff not required in config",
			message: `Add feature
This is a detailed description.`,
			requireSignOff: false,
			allowMultiple:  false,
			expectedValid:  true,
			description:    "Should pass when sign-off is not required by config",
		},
		{
			name: "Multiple sign-offs allowed in config",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should pass with multiple sign-offs when allowed by config",
		},
		{
			name: "Multiple sign-offs allowed regardless of config",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			requireSignOff: true,
			allowMultiple:  false,
			expectedValid:  true,
			description:    "Multiple sign-offs are actually allowed by this rule implementation",
		},
		// Additional config-based test cases
		{
			name: "Valid single sign-off with strict config",
			message: `Implement feature
Detailed implementation notes.
Signed-off-by: Developer <dev@example.com>`,
			requireSignOff: true,
			allowMultiple:  false,
			expectedValid:  true,
			description:    "Should pass with single sign-off under strict config",
		},
		{
			name: "No sign-off with permissive config",
			message: `Quick fix
Simple one-line change.`,
			requireSignOff: false,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should pass without sign-off under permissive config",
		},
		{
			name: "Complex commit with multiple contributors allowed",
			message: `Major refactoring
This is a collaborative effort involving multiple developers.

Co-authored-by: Alice <alice@example.com>
Co-authored-by: Bob <bob@example.com>
Signed-off-by: Charlie <charlie@example.com>
Signed-off-by: Diana <diana@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle complex multi-contributor commits",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit(testCase.message)

			// Create rule with config-based options
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						RequireSignoff: testCase.requireSignOff,
					},
				},
				Signing: config.SigningConfig{
					RequireMultiSignoff: testCase.allowMultiple,
				},
			}

			rule := rules.NewSignOffRule(cfg)

			failures := rule.Validate(commit, cfg)

			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors for case: %s", testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected validation errors but got none for case: %s", testCase.description)
			}
		})
	}
}

// Additional edge case testing.
func TestSignOffRuleEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		requireSignOff bool
		allowMultiple  bool
		expectedValid  bool
		description    string
	}{
		{
			name: "Very long sign-off line",
			message: `Fix issue
Signed-off-by: Very Long Developer Name With Multiple Middle Names And Suffixes Jr. III <very.long.email.address.with.multiple.dots.and.plus.signs+tag@very-long-domain-name.co.uk>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle very long sign-off lines",
		},
		{
			name: "Sign-off with Unicode characters",
			message: `Update internationalization
Signed-off-by: 田中太郎 <tanaka@example.jp>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle Unicode characters in sign-offs",
		},
		{
			name: "Empty lines around sign-off",
			message: `Fix formatting


Signed-off-by: Developer <dev@example.com>


`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  true,
			description:    "Should handle empty lines around sign-offs",
		},
		{
			name: "Partial sign-off match requires exact format",
			message: `Update code
This line mentions Signed-off-by but is not a real sign-off
Actually signed-off-by: Real Developer <real@example.com>`,
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			description:    "Rule requires exact format and position",
		},
		{
			name:           "Subject-only commit without sign-off",
			message:        "Quick fix",
			requireSignOff: true,
			allowMultiple:  true,
			expectedValid:  false,
			description:    "Should fail for subject-only commits when sign-off required",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := testdata.Commit(testCase.message)

			// Create rule with config
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						RequireSignoff: testCase.requireSignOff,
					},
				},
				Signing: config.SigningConfig{
					RequireMultiSignoff: testCase.allowMultiple,
				},
			}

			rule := rules.NewSignOffRule(cfg)

			failures := rule.Validate(commit, cfg)

			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors for case: %s", testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected validation errors but got none for case: %s", testCase.description)
			}
		})
	}
}
