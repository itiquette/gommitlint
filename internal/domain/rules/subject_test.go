// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

func TestSubjectRule_Validate(t *testing.T) {
	tests := []struct {
		name           string
		subject        string
		config         config.Config
		wantErrCount   int
		wantErrCodes   []string
		wantErrMessage string
	}{
		{
			name:    "valid subject - all rules pass",
			subject: "Add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:     50,
						Case:          "sentence",
						ForbidEndings: []string{".", ","},
					},
				},
			},
			wantErrCount: 0,
		},
		{
			name:    "too long subject",
			subject: "This is a very long subject line that exceeds the maximum allowed length",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "sentence",
					},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrSubjectTooLong)},
			wantErrMessage: "subject exceeds 50 characters",
		},
		{
			name:    "wrong case",
			subject: "add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "sentence",
					},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrSubjectCase)},
			wantErrMessage: "should be in sentence case",
		},
		{
			name:    "invalid suffix",
			subject: "Add new feature.",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:     50,
						Case:          "sentence",
						ForbidEndings: []string{".", ","},
					},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrSubjectSuffix)},
			wantErrMessage: "ends with invalid character",
		},
		{
			name:    "multiple errors - length and case",
			subject: "this is a very long subject line that exceeds the maximum allowed length and has wrong case",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "sentence",
					},
				},
			},
			wantErrCount: 2,
			wantErrCodes: []string{string(domain.ErrSubjectTooLong), string(domain.ErrSubjectCase)},
		},
		{
			name:    "conventional commit valid",
			subject: "feat: Add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "sentence",
					},
				},
				Rules: config.RulesConfig{
					Enabled: []string{"conventional"},
				},
			},
			wantErrCount: 0,
		},
		{
			name:    "conventional commit wrong case in description",
			subject: "feat: add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "sentence",
					},
				},
				Rules: config.RulesConfig{
					Enabled: []string{"conventional"},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrSubjectCase)},
			wantErrMessage: "should be in sentence case",
		},
		{
			name:    "empty subject",
			subject: "",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "sentence",
					},
				},
			},
			wantErrCount: 2, // Missing subject error from both case and suffix validation
			wantErrCodes: []string{string(domain.ErrMissingSubject)},
		},
		{
			name:    "case ignore - always valid",
			subject: "whatever CaSe ThIs Is",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "ignore",
					},
				},
			},
			wantErrCount: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewSubjectRule(testCase.config)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, nil, &testCase.config)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				// Check that expected error codes are present
				foundCodes := make(map[string]bool)
				for _, err := range errors {
					foundCodes[err.Code] = true

					if testCase.wantErrMessage != "" {
						require.Contains(t, err.Message, testCase.wantErrMessage)
					}
				}

				for _, expectedCode := range testCase.wantErrCodes {
					require.True(t, foundCodes[expectedCode], "expected error code %s not found", expectedCode)
				}
			}
		})
	}
}

func TestSubjectRule_Name(t *testing.T) {
	rule := NewSubjectRule(config.Config{})
	require.Equal(t, "Subject", rule.Name())
}

func TestSubjectRule_DefaultConfig(t *testing.T) {
	// Test default values when config is empty
	rule := NewSubjectRule(config.Config{})

	require.Equal(t, 72, rule.maxLength)
	require.Equal(t, "sentence", rule.caseChoice)
	require.Equal(t, ".,", rule.invalidSuffixes)
	require.True(t, rule.checkCommit) // conventional is enabled by default
	require.False(t, rule.allowNonAlpha)
}

func TestSubjectRule_ConfigOverrides(t *testing.T) {
	cfg := config.Config{
		Message: config.MessageConfig{
			Subject: config.SubjectConfig{
				MaxLength:         100,
				Case:              "lower",
				ForbidEndings:     []string{"!", "?"},
				RequireImperative: true,
			},
		},
		Rules: config.RulesConfig{
			Enabled: []string{"conventional"},
		},
	}

	rule := NewSubjectRule(cfg)

	require.Equal(t, 100, rule.maxLength)
	require.Equal(t, "lower", rule.caseChoice)
	require.Equal(t, "!?", rule.invalidSuffixes)
	require.True(t, rule.checkCommit)
	require.True(t, rule.allowNonAlpha)
}
