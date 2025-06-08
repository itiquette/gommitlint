// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"strings"
	"testing"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
	"github.com/stretchr/testify/require"
)

// createSignoffTestCommit creates a test commit with the given message.
func createSignoffTestCommit(message string) domain.Commit {
	commit := domain.Commit{
		Hash:          "abc123def456",
		Subject:       message,
		Message:       message,
		Author:        "Test User",
		AuthorEmail:   "test@example.com",
		CommitDate:    time.Now().Format(time.RFC3339),
		IsMergeCommit: false,
	}

	// Parse subject and body from message
	lines := strings.Split(message, "\n")
	if len(lines) > 0 {
		commit.Subject = lines[0]
	}

	// Parse body - everything after the subject line
	if len(lines) > 1 {
		bodyStart := 1

		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "" {
				bodyStart = i + 1

				break
			}
		}

		if bodyStart < len(lines) {
			commit.Body = strings.Join(lines[bodyStart:], "\n")
		} else if bodyStart == 1 {
			commit.Body = strings.Join(lines[1:], "\n")
		}
	}

	return commit
}

func TestSignOffRule(t *testing.T) {
	// Standard sign-off format for tests
	const validSignOff = "Signed-off-by: Dev Eloper <dev@example.com>"

	tests := []struct {
		name            string
		message         string
		minSignoffCount int
		expectedValid   bool
		errorContains   string
		description     string
	}{
		{
			name: "Valid sign-off",
			message: `Add feature
This is a detailed description of the feature.
` + validSignOff,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with valid sign-off",
		},
		{
			name:            "Valid sign-off with crlf",
			message:         "Add feature\r\n\r\nThis is a description.\r\n\r\n" + validSignOff,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle CRLF line endings correctly",
		},
		{
			name: "Valid sign-off with multiple signers",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with multiple sign-offs",
		},
		{
			name: "Missing sign-off (required)",
			message: `Add feature
This is a detailed description of the feature.`,
			minSignoffCount: 1,
			expectedValid:   false,
			errorContains:   "Missing required sign-off",
			description:     "Should fail when sign-off is required but missing",
		},
		{
			name: "Missing sign-off (not required)",
			message: `Add feature
This is a detailed description of the feature.`,
			minSignoffCount: 0,
			expectedValid:   true,
			description:     "Should pass when sign-off is not required",
		},
		{
			name:            "Empty message",
			message:         "",
			minSignoffCount: 1,
			expectedValid:   false,
			errorContains:   "Missing required sign-off",
			description:     "Should fail with empty message when sign-off required",
		},
		{
			name:            "Whitespace only message",
			message:         "   \t\n  ",
			minSignoffCount: 1,
			expectedValid:   false,
			errorContains:   "Missing required sign-off",
			description:     "Should fail with whitespace-only message when sign-off required",
		},
		// Additional comprehensive test cases
		{
			name: "Multiple sign-offs are allowed by default",
			message: `Fix critical bug
Signed-off-by: Developer One <dev1@example.com>
Signed-off-by: Developer Two <dev2@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Multiple sign-offs are actually allowed by this rule",
		},
		{
			name: "Single sign-off when required",
			message: `Fix bug
Detailed description of the fix.
Signed-off-by: Developer <dev@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with single sign-off when required",
		},
		{
			name: "Sign-off with complex email format",
			message: `Update documentation
Signed-off-by: John Doe Jr. <john.doe+dev@company.co.uk>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle complex email formats in sign-offs",
		},
		{
			name: "Sign-off with special characters in name",
			message: `Fix encoding issue
Signed-off-by: José María González <jose@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle special characters in names",
		},
		{
			name: "Malformed sign-off missing email is rejected",
			message: `Add feature
Signed-off-by: Developer Name`,
			minSignoffCount: 1,
			expectedValid:   false,
			errorContains:   "Missing required sign-off",
			description:     "Strict DCO validation rejects malformed sign-offs without proper email",
		},
		{
			name: "Malformed sign-off missing name is rejected",
			message: `Add feature
Signed-off-by: <dev@example.com>`,
			minSignoffCount: 1,
			expectedValid:   false,
			errorContains:   "Missing required sign-off",
			description:     "Strict DCO validation rejects malformed sign-offs without proper name",
		},
		{
			name: "Sign-off must be at end of message",
			message: `Add feature
Signed-off-by: Developer <dev@example.com>
Additional description after sign-off.`,
			minSignoffCount: 1,
			expectedValid:   false,
			description:     "Sign-off is only detected at the end of message",
		},
		{
			name: "Case sensitive sign-off detection",
			message: `Fix issue
signed-off-by: Developer <dev@example.com>`,
			minSignoffCount: 1,
			expectedValid:   false,
			description:     "Sign-off detection is case sensitive",
		},
		{
			name: "Sign-off with trailing whitespace",
			message: `Update code
Signed-off-by: Developer <dev@example.com>   `,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle sign-offs with trailing whitespace",
		},
		{
			name: "Mixed case sign-offs not detected",
			message: `Collaborative fix
Signed-off-by: Developer One <dev1@example.com>
signed-off-by: Developer Two <dev2@example.com>
Signed-Off-By: Developer Three <dev3@example.com>`,
			minSignoffCount: 1,
			expectedValid:   false,
			description:     "Only exact case sign-offs are detected",
		},
		{
			name: "Sign-off without required field when not required",
			message: `Simple fix
No sign-off here`,
			minSignoffCount: 0,
			expectedValid:   true,
			description:     "Should pass without sign-off when not required",
		},
		{
			name: "Co-authored-by with sign-off",
			message: `Joint development
Co-authored-by: Partner <partner@example.com>
Signed-off-by: Developer <dev@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should work with other commit metadata like Co-authored-by",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := createSignoffTestCommit(testCase.message)

			// Build options based on test case
			// Create config with options
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minSignoffCount,
					},
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
		name            string
		message         string
		minSignoffCount int
		expectedValid   bool
		description     string
	}{
		{
			name: "SignOff required in config",
			message: `Add feature
This is a commit without a sign-off.`,
			minSignoffCount: 1,
			expectedValid:   false,
			description:     "Should fail when sign-off is required by config but missing",
		},
		{
			name: "SignOff not required in config",
			message: `Add feature
This is a detailed description.`,
			minSignoffCount: 0,
			expectedValid:   true,
			description:     "Should pass when sign-off is not required by config",
		},
		{
			name: "Multiple sign-offs accepted",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with multiple sign-offs",
		},
		{
			name: "Multiple sign-offs with minimum count",
			message: `Fix bug
Signed-off-by: Laval Lion <laval.lion@cavora.org>
Signed-off-by: Cragger Crocodile <cragger@svamp.org>`,
			minSignoffCount: 2,
			expectedValid:   true,
			description:     "Should pass when multiple sign-offs meet minimum count",
		},
		// Additional config-based test cases
		{
			name: "Valid single sign-off",
			message: `Implement feature
Detailed implementation notes.
Signed-off-by: Developer <dev@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with single sign-off",
		},
		{
			name: "No sign-off with permissive config",
			message: `Quick fix
Simple one-line change.`,
			minSignoffCount: 0,
			expectedValid:   true,
			description:     "Should pass without sign-off under permissive config",
		},
		{
			name: "Complex commit with multiple contributors",
			message: `Major refactoring
This is a collaborative effort involving multiple developers.

Co-authored-by: Alice <alice@example.com>
Co-authored-by: Bob <bob@example.com>
Signed-off-by: Charlie <charlie@example.com>
Signed-off-by: Diana <diana@example.com>`,
			minSignoffCount: 2,
			expectedValid:   true,
			description:     "Should handle complex multi-contributor commits",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := createSignoffTestCommit(testCase.message)

			// Create rule with config-based options
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minSignoffCount,
					},
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

// TestSignOffRule_MinSignoffCount tests the enhanced minSignoffCount functionality.
func TestSignOffRule_MinSignoffCount(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		minSignoffCount int
		expectedValid   bool
		errorCode       string
		description     string
	}{
		{
			name: "No sign-offs required (0 count)",
			message: `Fix issue
This commit has no sign-offs.`,
			minSignoffCount: 0,
			expectedValid:   true,
			description:     "Should pass when no sign-offs required",
		},
		{
			name: "Single sign-off required and present",
			message: `Fix issue
This commit has one sign-off.
Signed-off-by: Developer <dev@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with single sign-off when 1 required",
		},
		{
			name: "Single sign-off required but missing",
			message: `Fix issue
This commit has no sign-offs.`,
			minSignoffCount: 1,
			expectedValid:   false,
			errorCode:       string(domain.ErrMissingSignoff),
			description:     "Should fail when sign-off required but missing",
		},
		{
			name: "Two sign-offs required and present",
			message: `Fix issue
This commit has two sign-offs.
Signed-off-by: Developer One <dev1@example.com>
Signed-off-by: Developer Two <dev2@example.com>`,
			minSignoffCount: 2,
			expectedValid:   true,
			description:     "Should pass with two sign-offs when 2 required",
		},
		{
			name: "Two sign-offs required but only one present",
			message: `Fix issue
This commit has one sign-off.
Signed-off-by: Developer <dev@example.com>`,
			minSignoffCount: 2,
			expectedValid:   false,
			errorCode:       string(domain.ErrMissingSignoff),
			description:     "Should fail when 2 required but only 1 present",
		},
		{
			name: "Three sign-offs required and present",
			message: `Fix issue
This commit has three sign-offs.
Signed-off-by: Developer One <dev1@example.com>
Signed-off-by: Developer Two <dev2@example.com>
Signed-off-by: Developer Three <dev3@example.com>`,
			minSignoffCount: 3,
			expectedValid:   true,
			description:     "Should pass with three sign-offs when 3 required",
		},
		{
			name: "More sign-offs than required",
			message: `Fix issue
This commit has three sign-offs.
Signed-off-by: Developer One <dev1@example.com>
Signed-off-by: Developer Two <dev2@example.com>
Signed-off-by: Developer Three <dev3@example.com>`,
			minSignoffCount: 2,
			expectedValid:   true,
			description:     "Should pass when more sign-offs than required",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with MinSignoffCount
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minSignoffCount,
					},
				},
			}

			// Create commit using testdata helper
			commit := createSignoffTestCommit(testCase.message)

			// Create rule with enhanced config
			rule := rules.NewSignOffRule(cfg)

			// Validate commit
			failures := rule.Validate(commit, cfg)

			// Check validity
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors for case: %s", testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected errors but found none for case: %s", testCase.description)

				if testCase.errorCode != "" {
					require.Equal(t, testCase.errorCode, failures[0].Code, "Expected specific error code")
				}
			}
		})
	}
}

// TestSignOffRule_UniqueSigners tests validation of unique signers when multiple required.
func TestSignOffRule_UniqueSigners(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		minSignoffCount int
		expectedValid   bool
		errorCode       string
		description     string
	}{
		{
			name: "Two unique signers required and present",
			message: `Fix issue
Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev Two <dev2@example.com>`,
			minSignoffCount: 2,
			expectedValid:   true,
			description:     "Should pass with two different signers",
		},
		{
			name: "Two signers required but one signer signed twice",
			message: `Fix issue
Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev One <dev1@example.com>`,
			minSignoffCount: 2,
			expectedValid:   false,
			errorCode:       string(domain.ErrInsufficientSignoffs),
			description:     "Should fail when same signer signs twice",
		},
		{
			name: "Three unique signers required and present",
			message: `Fix issue
Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev Two <dev2@example.com>
Signed-off-by: Dev Three <dev3@example.com>`,
			minSignoffCount: 3,
			expectedValid:   true,
			description:     "Should pass with three different signers",
		},
		{
			name: "Single sign-off does not trigger unique validation",
			message: `Fix issue
Signed-off-by: Dev One <dev1@example.com>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should pass with single sign-off (uniqueness not checked)",
		},
		{
			name: "Mixed duplicate and unique signers",
			message: `Fix issue
Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev Two <dev2@example.com>
Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev Three <dev3@example.com>`,
			minSignoffCount: 3,
			expectedValid:   false,
			errorCode:       string(domain.ErrInsufficientSignoffs),
			description:     "Should fail when duplicates detected in multi-signer requirement",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with MinSignoffCount
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minSignoffCount,
					},
				},
			}

			// Create commit using testdata helper
			commit := createSignoffTestCommit(testCase.message)

			// Create rule with enhanced config
			rule := rules.NewSignOffRule(cfg)

			// Validate commit
			failures := rule.Validate(commit, cfg)

			// Check validity
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors for case: %s", testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected errors but found none for case: %s", testCase.description)

				if testCase.errorCode != "" {
					require.Equal(t, testCase.errorCode, failures[0].Code, "Expected specific error code")
				}
			}
		})
	}
}

// TestSignOffRule_PlacementValidation tests sign-off placement validation.
func TestSignOffRule_PlacementValidation(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedValid bool
		errorCode     string
		description   string
	}{
		{
			name: "Sign-off at proper end position",
			message: `Fix issue
Detailed description here.

Signed-off-by: Developer <dev@example.com>`,
			expectedValid: true,
			description:   "Should pass with sign-off at end",
		},
		{
			name: "Sign-off with content after",
			message: `Fix issue
Signed-off-by: Developer <dev@example.com>
Additional content after sign-off`,
			expectedValid: false,
			errorCode:     string(domain.ErrMisplacedSignoff),
			description:   "Should fail when content appears after sign-off",
		},
		{
			name: "Multiple sign-offs at end",
			message: `Fix issue
Detailed description.

Signed-off-by: Dev One <dev1@example.com>
Signed-off-by: Dev Two <dev2@example.com>`,
			expectedValid: true,
			description:   "Should pass with multiple sign-offs at end",
		},
		{
			name: "Sign-off with trailing empty lines",
			message: `Fix issue

Signed-off-by: Developer <dev@example.com>


`,
			expectedValid: true,
			description:   "Should pass with trailing empty lines after sign-off",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config requiring sign-offs
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: 1,
					},
				},
			}

			// Create commit using testdata helper
			commit := createSignoffTestCommit(testCase.message)

			// Create rule
			rule := rules.NewSignOffRule(cfg)

			// Validate commit
			failures := rule.Validate(commit, cfg)

			// Check validity
			if testCase.expectedValid {
				require.Empty(t, failures, "Expected no validation errors for case: %s", testCase.description)
			} else {
				require.NotEmpty(t, failures, "Expected errors but found none for case: %s", testCase.description)

				if testCase.errorCode != "" {
					require.Equal(t, testCase.errorCode, failures[0].Code, "Expected specific error code")
				}
			}
		})
	}
}

// Additional edge case testing.
func TestSignOffRuleEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		minSignoffCount int
		expectedValid   bool
		description     string
	}{
		{
			name: "Very long sign-off line",
			message: `Fix issue
Signed-off-by: Very Long Developer Name With Multiple Middle Names And Suffixes Jr. III <very.long.email.address.with.multiple.dots.and.plus.signs+tag@very-long-domain-name.co.uk>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle very long sign-off lines",
		},
		{
			name: "Sign-off with Unicode characters",
			message: `Update internationalization
Signed-off-by: 田中太郎 <tanaka@example.jp>`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle Unicode characters in sign-offs",
		},
		{
			name: "Empty lines around sign-off",
			message: `Fix formatting


Signed-off-by: Developer <dev@example.com>


`,
			minSignoffCount: 1,
			expectedValid:   true,
			description:     "Should handle empty lines around sign-offs",
		},
		{
			name: "Partial sign-off match requires exact format",
			message: `Update code
This line mentions Signed-off-by but is not a real sign-off
Actually signed-off-by: Real Developer <real@example.com>`,
			minSignoffCount: 1,
			expectedValid:   false,
			description:     "Rule requires exact format and position",
		},
		{
			name:            "Subject-only commit without sign-off",
			message:         "Quick fix",
			minSignoffCount: 1,
			expectedValid:   false,
			description:     "Should fail for subject-only commits when sign-off required",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create commit using builder
			commit := createSignoffTestCommit(testCase.message)

			// Create rule with config
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minSignoffCount,
					},
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

// TestSignOffRule_ErrorContextAndHelp tests that errors include proper context and help text.
func TestSignOffRule_ErrorContextAndHelp(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		minCount     int
		wantCode     string
		wantContext  map[string]string
		wantHelpText string
	}{
		{
			name: "Missing sign-off error includes context",
			message: `Fix issue
No sign-offs here`,
			minCount: 1,
			wantCode: string(domain.ErrMissingSignoff),
			wantContext: map[string]string{
				"actual":   "0",
				"expected": "1",
			},
			wantHelpText: "Add DCO sign-off line",
		},
		{
			name: "Insufficient count error includes context",
			message: `Fix issue
Signed-off-by: Developer <dev@example.com>`,
			minCount: 2,
			wantCode: string(domain.ErrMissingSignoff),
			wantContext: map[string]string{
				"actual":   "1",
				"expected": "2",
			},
			wantHelpText: "Add DCO sign-off line",
		},
		{
			name: "Misplaced sign-off error includes context",
			message: `Fix issue
Signed-off-by: Developer <dev@example.com>
Content after sign-off`,
			minCount: 1,
			wantCode: string(domain.ErrMisplacedSignoff),
			wantContext: map[string]string{
				"actual":   "Content found after sign-off",
				"expected": "Sign-offs at end",
			},
			wantHelpText: "Move all sign-off lines to the end",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minCount,
					},
				},
			}

			commit := createSignoffTestCommit(testCase.message)
			rule := rules.NewSignOffRule(cfg)
			errors := rule.Validate(commit, cfg)

			require.NotEmpty(t, errors, "Expected validation error")
			require.Equal(t, testCase.wantCode, errors[0].Code)

			// Check context values
			for key, expectedValue := range testCase.wantContext {
				require.Equal(t, expectedValue, errors[0].Context[key], "Context value mismatch for key: %s", key)
			}

			// Check help text
			require.Contains(t, errors[0].Help, testCase.wantHelpText, "Help text should contain expected guidance")
		})
	}
}

// TestSignOffRule_StrictDCOCompliance tests that the SignOff rule enforces strict DCO compliance.
func TestSignOffRule_StrictDCOCompliance(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantValid   bool
		wantErrCode string
		description string
	}{
		{
			name: "Valid DCO format - full name and email",
			body: `Fix critical bug in authentication

This commit fixes a security vulnerability.

Signed-off-by: John Doe <john.doe@example.com>`,
			wantValid:   true,
			description: "Proper DCO format should pass validation",
		},
		{
			name: "Valid DCO format - name with spaces",
			body: `Add new feature

Signed-off-by: Mary Jane Watson <mary.jane@company.org>`,
			wantValid:   true,
			description: "Names with spaces should be valid",
		},
		{
			name: "Valid DCO format - hyphenated email domain",
			body: `Update documentation

Signed-off-by: Developer Name <dev@sub-domain.example.com>`,
			wantValid:   true,
			description: "Complex email domains should be valid",
		},
		{
			name: "Invalid DCO format - missing email",
			body: `Fix bug

Signed-off-by: John Doe`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMissingSignoff),
			description: "Sign-off without email should be rejected",
		},
		{
			name: "Invalid DCO format - missing name",
			body: `Fix bug

Signed-off-by: <john@example.com>`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMissingSignoff),
			description: "Sign-off without name should be rejected",
		},
		{
			name: "Invalid DCO format - email without angle brackets",
			body: `Fix bug

Signed-off-by: John Doe john@example.com`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMissingSignoff),
			description: "Email without angle brackets should be rejected",
		},
		{
			name: "Invalid DCO format - only angle brackets",
			body: `Fix bug

Signed-off-by: <>`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMissingSignoff),
			description: "Empty angle brackets should be rejected",
		},
		{
			name: "Invalid DCO format - email without @ symbol",
			body: `Fix bug

Signed-off-by: John Doe <johndoeexample.com>`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMissingSignoff),
			description: "Email without @ symbol should be rejected",
		},
		{
			name: "Invalid DCO format - just text without proper format",
			body: `Fix bug

Signed-off-by: some random text here`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMissingSignoff),
			description: "Non-DCO format text should be rejected",
		},
		{
			name: "Valid DCO format - multiple valid sign-offs",
			body: `Fix bug collaboratively

This was a team effort.

Signed-off-by: Alice Developer <alice@example.com>
Signed-off-by: Bob Reviewer <bob@example.com>`,
			wantValid:   true,
			description: "Multiple valid DCO sign-offs should pass",
		},
		{
			name: "Mixed valid and invalid sign-offs",
			body: `Fix bug

Signed-off-by: Alice Developer <alice@example.com>
Signed-off-by: Invalid Line Without Email`,
			wantValid:   false,
			wantErrCode: string(domain.ErrMisplacedSignoff),
			description: "Mixed valid/invalid should fail due to invalid content after valid sign-off",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: 1,
					},
				},
			}

			rule := rules.NewSignOffRule(cfg)
			commit := domain.Commit{
				Subject: "Test commit",
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, cfg)

			if testCase.wantValid {
				require.Empty(t, errors, testCase.description)
			} else {
				require.NotEmpty(t, errors, testCase.description)

				if testCase.wantErrCode != "" {
					require.Equal(t, testCase.wantErrCode, errors[0].Code, testCase.description)
				}
			}
		})
	}
}

// TestSignOffRule_DCOExamples tests real-world DCO examples to ensure compliance.
func TestSignOffRule_DCOExamples(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantValid   bool
		description string
	}{
		{
			name: "Linux kernel style DCO",
			body: `mm: fix memory leak in page allocation

This patch fixes a memory leak that occurs when page allocation
fails under memory pressure.

Signed-off-by: Linus Torvalds <torvalds@linux-foundation.org>`,
			wantValid:   true,
			description: "Linux kernel style should be valid",
		},
		{
			name: "Corporate email DCO",
			body: `security: patch CVE-2024-1234

Applied security patch for vulnerability.

Signed-off-by: Security Team <security@bigcorp.com>`,
			wantValid:   true,
			description: "Corporate email format should be valid",
		},
		{
			name: "International name DCO",
			body: `i18n: add support for Japanese characters

Added Unicode support for Japanese text rendering.

Signed-off-by: Tanaka Hiroshi <tanaka@example.jp>`,
			wantValid:   true,
			description: "International names should be valid",
		},
		{
			name: "Hyphenated name DCO",
			body: `docs: update README

Updated documentation for clarity.

Signed-off-by: Mary-Jane Smith-Wilson <mary-jane@example.com>`,
			wantValid:   true,
			description: "Hyphenated names should be valid",
		},
		{
			name: "Invalid - Git trailer without email",
			body: `fix: resolve authentication bug

Fixed login issue for users.

Signed-off-by: Developer`,
			wantValid:   false,
			description: "Git trailer without email should be invalid",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: 1,
					},
				},
			}

			rule := rules.NewSignOffRule(cfg)
			commit := domain.Commit{
				Subject: "Test commit",
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, cfg)

			if testCase.wantValid {
				require.Empty(t, errors, testCase.description)
			} else {
				require.NotEmpty(t, errors, testCase.description)
			}
		})
	}
}

// TestSignOffRule_StrictValidationBehavior tests the specific behavior of strict validation.
func TestSignOffRule_StrictValidationBehavior(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		minSignoffCount int
		expectFound     int
		expectError     bool
		description     string
	}{
		{
			name: "Two valid DCO sign-offs counted correctly",
			body: `Fix issue together

Signed-off-by: Alice <alice@example.com>
Signed-off-by: Bob <bob@example.com>`,
			minSignoffCount: 2,
			expectFound:     2,
			expectError:     false,
			description:     "Two valid sign-offs should meet requirement",
		},
		{
			name: "One valid, one invalid - only valid counted",
			body: `Fix issue together

Signed-off-by: Alice <alice@example.com>
Signed-off-by: Bob Without Email`,
			minSignoffCount: 2,
			expectFound:     1,
			expectError:     true,
			description:     "Only strictly valid sign-offs should be counted",
		},
		{
			name: "Mixed case - case sensitive validation",
			body: `Fix issue

SIGNED-OFF-BY: Alice <alice@example.com>`,
			minSignoffCount: 1,
			expectFound:     0,
			expectError:     true,
			description:     "Case sensitivity should be enforced",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						MinSignoffCount: testCase.minSignoffCount,
					},
				},
			}

			rule := rules.NewSignOffRule(cfg)
			commit := domain.Commit{
				Subject: "Test commit",
				Body:    testCase.body,
			}

			errors := rule.Validate(commit, cfg)

			if testCase.expectError {
				require.NotEmpty(t, errors, testCase.description)
				// Check that the error context reflects the actual count found
				if len(errors) > 0 && errors[0].Context["found"] != "" {
					require.Contains(t, errors[0].Context["found"], string(rune('0'+testCase.expectFound)),
						"Error should reflect actual valid sign-offs found")
				}
			} else {
				require.Empty(t, errors, testCase.description)
			}
		})
	}
}
