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

			errors := rule.Validate(commit, testCase.config)

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
	require.False(t, rule.requireImperative)
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
	require.False(t, rule.allowNonAlpha) // Fixed: this should not be tied to imperative setting
	require.True(t, rule.requireImperative)
}

func TestSubjectRule_ImperativeValidation(t *testing.T) {
	tests := []struct {
		name                string
		subject             string
		requireImperative   bool
		conventionalEnabled bool
		wantErrCount        int
		wantErrCode         string
		wantErrMessage      string
	}{
		{
			name:              "imperative disabled - no validation",
			subject:           "Added new feature",
			requireImperative: false,
			wantErrCount:      0,
		},
		{
			name:              "valid imperative",
			subject:           "Add new feature",
			requireImperative: true,
			wantErrCount:      0,
		},
		{
			name:              "past tense violation",
			subject:           "Added new feature",
			requireImperative: true,
			wantErrCount:      1,
			wantErrCode:       string(domain.ErrNonImperative),
			wantErrMessage:    "not in imperative mood",
		},
		{
			name:              "gerund violation",
			subject:           "Adding new feature",
			requireImperative: true,
			wantErrCount:      1,
			wantErrCode:       string(domain.ErrNonImperative),
			wantErrMessage:    "not in imperative mood",
		},
		{
			name:              "third person violation",
			subject:           "Adds new feature",
			requireImperative: true,
			wantErrCount:      1,
			wantErrCode:       string(domain.ErrNonImperative),
			wantErrMessage:    "not in imperative mood",
		},
		{
			name:                "conventional commit - valid imperative",
			subject:             "feat: Add new feature",
			requireImperative:   true,
			conventionalEnabled: true,
			wantErrCount:        0,
		},
		{
			name:                "conventional commit - past tense violation",
			subject:             "feat: Added new feature",
			requireImperative:   true,
			conventionalEnabled: true,
			wantErrCount:        1,
			wantErrCode:         string(domain.ErrNonImperative),
			wantErrMessage:      "not in imperative mood",
		},
		{
			name:              "empty subject - no imperative error",
			subject:           "",
			requireImperative: true,
			wantErrCount:      0, // Empty subject handled by other validation
		},
		{
			name:              "single word imperative",
			subject:           "Fix",
			requireImperative: true,
			wantErrCount:      0,
		},
		{
			name:              "single word past tense",
			subject:           "Fixed",
			requireImperative: true,
			wantErrCount:      1,
			wantErrCode:       string(domain.ErrNonImperative),
			wantErrMessage:    "not in imperative mood",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:         72,
						Case:              "ignore", // Focus on imperative testing
						RequireImperative: testCase.requireImperative,
					},
				},
			}

			if testCase.conventionalEnabled {
				cfg.Rules.Enabled = []string{"ConventionalCommit"}
			}

			rule := NewSubjectRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			if testCase.wantErrCount == 0 {
				// Filter out non-imperative errors to focus on imperative validation
				imperativeErrors := make([]domain.ValidationError, 0)

				for _, err := range errors {
					if err.Code == string(domain.ErrNonImperative) || err.Code == string(domain.ErrNoFirstWord) {
						imperativeErrors = append(imperativeErrors, err)
					}
				}

				require.Empty(t, imperativeErrors, "unexpected imperative validation errors")
			} else {
				// Check that we have the expected imperative error
				found := false

				for _, err := range errors {
					if err.Code == testCase.wantErrCode {
						found = true

						if testCase.wantErrMessage != "" {
							require.Contains(t, err.Message, testCase.wantErrMessage)
						}

						break
					}
				}

				require.True(t, found, "expected error code %s not found", testCase.wantErrCode)
			}
		})
	}
}

func TestSubjectRule_ExtractFirstWord(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected string
	}{
		{
			name:     "simple word",
			subject:  "Add feature",
			expected: "Add",
		},
		{
			name:     "single word",
			subject:  "Fix",
			expected: "Fix",
		},
		{
			name:     "empty string",
			subject:  "",
			expected: "",
		},
		{
			name:     "whitespace only",
			subject:  "   ",
			expected: "",
		},
		{
			name:     "leading whitespace",
			subject:  "  Update docs",
			expected: "Update",
		},
	}

	rule := NewSubjectRule(config.Config{})

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := rule.extractFirstWord(testCase.subject)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestSubjectRule_CategorizeVerb(t *testing.T) {
	tests := []struct {
		name          string
		word          string
		wantCategory  string
		wantViolation bool
	}{
		{
			name:          "imperative verb",
			word:          "add",
			wantCategory:  "",
			wantViolation: false,
		},
		{
			name:          "past tense verb",
			word:          "added",
			wantCategory:  "past_tense",
			wantViolation: true,
		},
		{
			name:          "gerund verb",
			word:          "adding",
			wantCategory:  "gerund",
			wantViolation: true,
		},
		{
			name:          "third person verb",
			word:          "adds",
			wantCategory:  "third_person",
			wantViolation: true,
		},
		{
			name:          "unknown verb",
			word:          "unknown",
			wantCategory:  "",
			wantViolation: false,
		},
		{
			name:          "case insensitive",
			word:          "ADDED",
			wantCategory:  "past_tense",
			wantViolation: true,
		},
		{
			name:          "base form ending with ed",
			word:          "need",
			wantCategory:  "",
			wantViolation: false,
		},
	}

	rule := NewSubjectRule(config.Config{})
	// Add some common base forms ending with "ed" for testing
	rule.baseFormsEndingWithED["need"] = true
	rule.baseFormsEndingWithED["seed"] = true
	rule.baseFormsEndingWithED["feed"] = true

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			category, violation := rule.categorizeVerb(strings.ToLower(testCase.word))
			require.Equal(t, testCase.wantCategory, category)
			require.Equal(t, testCase.wantViolation, violation)
		})
	}
}
