// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

func TestCommitBodyRule_BlankLineValidation(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectError     bool
		expectedErrCode string
	}{
		{
			name:        "Valid structure with blank line",
			message:     "feat: add feature\n\nThis is the body explaining the feature",
			expectError: false,
		},
		{
			name:            "Missing blank line",
			message:         "feat: add feature\nThis is the body without blank line",
			expectError:     true,
			expectedErrCode: "missing_blank_line",
		},
		{
			name:        "Subject only - no error",
			message:     "feat: add feature",
			expectError: false,
		},
		{
			name:        "Multiple blank lines - still valid",
			message:     "feat: add feature\n\n\nThis is the body with multiple blank lines",
			expectError: false,
		},
		{
			name:        "Merge commit - skip validation",
			message:     "Merge branch 'feature'\nSome content",
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewCommitBodyRule(config.Config{})
			commit := domain.ParseCommitMessage(testCase.message)

			// Mark as merge commit for the merge test case
			if strings.Contains(testCase.name, "Merge") {
				commit.IsMergeCommit = true
			}

			errors := rule.Validate(commit, config.Config{})

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected validation error")

				found := false

				for _, err := range errors {
					if err.Code == testCase.expectedErrCode {
						found = true

						break
					}
				}

				require.True(t, found, "expected error code %s not found", testCase.expectedErrCode)
			} else {
				require.Empty(t, errors, "expected no validation errors")
			}
		})
	}
}

func TestCommitBodyRule_StructureValidation(t *testing.T) {
	tests := []struct {
		name            string
		config          config.Config
		message         string
		expectError     bool
		expectedErrCode string
	}{
		{
			name: "Valid three-part structure when required",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:  true,
						MinLength: 10,
					},
				},
			},
			message:     "feat: add feature\n\nDetailed explanation of the feature",
			expectError: false,
		},
		{
			name: "Invalid structure when body required",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:  true,
						MinLength: 10,
					},
				},
			},
			message:         "feat: add feature",
			expectError:     true,
			expectedErrCode: "invalid_structure",
		},
		{
			name: "No structure requirement when body not required",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required: false,
					},
				},
			},
			message:     "feat: add feature",
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewCommitBodyRule(testCase.config)
			commit := domain.ParseCommitMessage(testCase.message)

			errors := rule.Validate(commit, config.Config{})

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected validation error")

				found := false

				for _, err := range errors {
					if err.Code == testCase.expectedErrCode {
						found = true

						break
					}
				}

				require.True(t, found, "expected error code %s not found", testCase.expectedErrCode)
			} else {
				require.Empty(t, errors, "expected no validation errors")
			}
		})
	}
}

func TestCommitBodyRule_SignOffPositioning(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectError     bool
		expectedErrCode string
	}{
		{
			name: "Valid sign-off at end",
			message: `feat: add feature

This is a detailed explanation of the feature.

Signed-off-by: John Doe <john@example.com>`,
			expectError: false,
		},
		{
			name: "Multiple sign-offs at end",
			message: `feat: add feature

This is a detailed explanation.

Signed-off-by: John Doe <john@example.com>
Co-authored-by: Jane Smith <jane@example.com>`,
			expectError: false,
		},
		{
			name: "Co-authored-by after sign-off (valid trailer)",
			message: `feat: add feature

This is a detailed explanation.

Signed-off-by: John Doe <john@example.com>
Co-authored-by: Jane Smith <jane@example.com>
Reviewed-by: Bob Wilson <bob@example.com>`,
			expectError: false,
		},
		{
			name: "All supported trailers after sign-off",
			message: `feat: add feature

This fixes a critical bug in the authentication system.

Signed-off-by: John Doe <john@example.com>
Co-authored-by: Jane Smith <jane@example.com>
Reviewed-by: Bob Wilson <bob@example.com>
Tested-by: Alice Cooper <alice@example.com>
Reported-by: Charlie Brown <charlie@example.com>
Fixes: #123
Closes: https://github.com/org/repo/issues/456`,
			expectError: false,
		},
		{
			name: "Content after sign-off",
			message: `feat: add feature

This is a detailed explanation.

Signed-off-by: John Doe <john@example.com>

Additional content after sign-off`,
			expectError:     true,
			expectedErrCode: "misplaced_signoff",
		},
		{
			name: "Sign-off in middle",
			message: `feat: add feature

This is a detailed explanation.

Signed-off-by: John Doe <john@example.com>

More explanation here`,
			expectError:     true,
			expectedErrCode: "misplaced_signoff",
		},
		{
			name: "No sign-offs - no error",
			message: `feat: add feature

This is a detailed explanation without any sign-offs.`,
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewCommitBodyRule(config.Config{})
			commit := domain.ParseCommitMessage(testCase.message)

			errors := rule.Validate(commit, config.Config{})

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected validation error")

				found := false

				for _, err := range errors {
					if err.Code == testCase.expectedErrCode {
						found = true

						break
					}
				}

				require.True(t, found, "expected error code %s not found", testCase.expectedErrCode)
			} else {
				require.Empty(t, errors, "expected no validation errors")
			}
		})
	}
}

func TestCommitBodyRule_EnhancedBodyValidation(t *testing.T) {
	tests := []struct {
		name            string
		config          config.Config
		message         string
		expectError     bool
		expectedErrCode string
	}{
		{
			name: "Minimum length validation",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:  true,
						MinLength: 20,
					},
				},
			},
			message:         "feat: add feature\n\nShort body",
			expectError:     true,
			expectedErrCode: "body_too_short",
		},
		{
			name: "Sign-off only when not allowed",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:         true,
						AllowSignoffOnly: false,
					},
				},
			},
			message:         "feat: add feature\n\nSigned-off-by: John Doe <john@example.com>",
			expectError:     true,
			expectedErrCode: "invalid_body",
		},
		{
			name: "Sign-off only when allowed",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:         true,
						AllowSignoffOnly: true,
					},
				},
			},
			message:     "feat: add feature\n\nSigned-off-by: John Doe <john@example.com>",
			expectError: false,
		},
		{
			name: "Missing body when required",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:  true,
						MinLength: 10,
					},
				},
			},
			message:         "feat: add feature",
			expectError:     true,
			expectedErrCode: "invalid_structure",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewCommitBodyRule(testCase.config)
			commit := domain.ParseCommitMessage(testCase.message)

			errors := rule.Validate(commit, config.Config{})

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected validation error")

				found := false

				for _, err := range errors {
					if err.Code == testCase.expectedErrCode {
						found = true

						break
					}
				}

				require.True(t, found, "expected error code %s not found", testCase.expectedErrCode)
			} else {
				require.Empty(t, errors, "expected no validation errors")
			}
		})
	}
}

func TestCommitBodyRule_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		config          config.Config
		message         string
		expectError     bool
		expectedErrCode string
	}{
		{
			name: "Empty body when required",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{Required: true},
				},
			},
			message:         "feat: add feature",
			expectError:     true,
			expectedErrCode: "invalid_structure",
		},
		{
			name: "RequireSignoff true but no signoff",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:        true,
						MinSignoffCount: 1,
					},
				},
			},
			message:     "feat: add feature\n\nThis is a detailed body without signoff",
			expectError: false, // SignOff rule handles this, not CommitBody
		},
		{
			name:   "Mixed trailers after signoff",
			config: config.Config{},
			message: `feat: add feature

This is the body content.

Signed-off-by: John Doe <john@example.com>
Co-authored-by: Jane Smith <jane@example.com>
Fixes: #123`,
			expectError: false,
		},
		{
			name: "Zero minLength when required",
			config: config.Config{
				Message: config.MessageConfig{
					Body: config.BodyConfig{
						Required:  true,
						MinLength: 0,
					},
				},
			},
			message:     "feat: add feature\n\nX",
			expectError: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewCommitBodyRule(testCase.config)
			commit := domain.ParseCommitMessage(testCase.message)

			errors := rule.Validate(commit, config.Config{})

			if testCase.expectError {
				require.NotEmpty(t, errors, "expected validation error")

				found := false

				for _, err := range errors {
					if err.Code == testCase.expectedErrCode {
						found = true

						break
					}
				}

				require.True(t, found, "expected error code %s not found", testCase.expectedErrCode)
			} else {
				require.Empty(t, errors, "expected no validation errors")
			}
		})
	}
}

func TestCommitBodyRule_IntegrationTest(t *testing.T) {
	// Test complete functionality integration
	t.Run("Complex commit with all validations", func(t *testing.T) {
		config := config.Config{
			Message: config.MessageConfig{
				Body: config.BodyConfig{
					Required:         true,
					MinLength:        20,
					AllowSignoffOnly: false,
				},
			},
		}

		rule := NewCommitBodyRule(config)

		// Valid complex commit
		validMessage := `feat: implement user authentication

This commit implements a comprehensive user authentication system
with support for JWT tokens and refresh token rotation.

The implementation includes proper error handling and security
measures to prevent common attack vectors.

Signed-off-by: John Doe <john@example.com>`

		commit := domain.ParseCommitMessage(validMessage)
		errors := rule.Validate(commit, config)
		require.Empty(t, errors, "valid complex commit should pass all validations")

		// Invalid complex commit - missing body when required (no blank line after subject)
		invalidMessage := `feat: implement user authentication
This commit implements authentication without blank line.

Signed-off-by: John Doe <john@example.com>

Additional content after sign-off`

		commit = domain.ParseCommitMessage(invalidMessage)
		errors = rule.Validate(commit, config)
		require.NotEmpty(t, errors, "invalid commit should have validation errors")

		// Should have invalid structure error (no proper subject + blank line + body)
		foundStructureError := false

		for _, err := range errors {
			if err.Code == "invalid_structure" {
				foundStructureError = true

				break
			}
		}

		require.True(t, foundStructureError, "should detect invalid structure (missing blank line)")
	})
}
