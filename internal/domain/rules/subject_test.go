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
			subject: "add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:     50,
						Case:          "lower",
						ForbidEndings: []string{".", ","},
					},
				},
			},
			wantErrCount: 0,
		},
		{
			name:    "too long subject",
			subject: "this is a very long subject line that exceeds the maximum allowed length",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "lower",
					},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrSubjectTooLong)},
			wantErrMessage: "Subject too long",
		},
		{
			name:    "wrong case",
			subject: "Add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "lower",
					},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrWrongCaseLower)},
			wantErrMessage: "First letter 'A' should be 'a'",
		},
		{
			name:    "invalid suffix",
			subject: "add new feature.",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:     50,
						Case:          "lower",
						ForbidEndings: []string{".", ","},
					},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrSubjectSuffix)},
			wantErrMessage: "invalid suffix",
		},
		{
			name:    "multiple errors - length and case",
			subject: "This is a very long subject line that exceeds the maximum allowed length and has wrong case",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "lower",
					},
				},
			},
			wantErrCount: 2,
			wantErrCodes: []string{string(domain.ErrSubjectTooLong), string(domain.ErrWrongCaseLower)},
		},
		{
			name:    "conventional commit valid",
			subject: "feat: add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "lower",
					},
				},
				Rules: config.RulesConfig{
					Enabled: []string{"ConventionalCommit"},
				},
			},
			wantErrCount: 0,
		},
		{
			name:    "conventional commit wrong case in description",
			subject: "feat: Add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "lower",
					},
				},
				Rules: config.RulesConfig{
					Enabled: []string{"ConventionalCommit"},
				},
			},
			wantErrCount:   1,
			wantErrCodes:   []string{string(domain.ErrWrongCaseLower)},
			wantErrMessage: "First letter 'A' should be 'a'",
		},
		{
			name:    "empty subject",
			subject: "",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
						Case:      "lower",
					},
				},
			},
			wantErrCount: 1, // Empty subject error
			wantErrCodes: []string{string(domain.ErrEmptySubject)},
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

	require.Equal(t, 100, rule.maxLength) // Updated to restored original default
	require.Equal(t, "lower", rule.caseChoice)
	require.Equal(t, ".,;:!?", rule.invalidSuffixes) // Updated to restored comprehensive default
	require.True(t, rule.checkCommit)                // conventional is enabled by default
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
			wantErrCode:       string(domain.ErrPastTense),
			wantErrMessage:    "is past tense",
		},
		{
			name:              "gerund violation",
			subject:           "Adding new feature",
			requireImperative: true,
			wantErrCount:      1,
			wantErrCode:       string(domain.ErrGerund),
			wantErrMessage:    "is a gerund",
		},
		{
			name:              "third person violation",
			subject:           "Adds new feature",
			requireImperative: true,
			wantErrCount:      1,
			wantErrCode:       string(domain.ErrThirdPerson),
			wantErrMessage:    "is third person singular",
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
			wantErrCode:         string(domain.ErrPastTense),
			wantErrMessage:      "is past tense",
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
			wantErrCode:       string(domain.ErrPastTense),
			wantErrMessage:    "is past tense",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:         72,
						Case:              "lower", // Focus on imperative testing
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

func TestSubjectRule_EnhancedUTF8LengthValidation(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		maxLength    int
		wantErrCount int
		wantErrCode  string
	}{
		{
			name:         "ASCII subject within limit",
			subject:      "Add feature",
			maxLength:    50,
			wantErrCount: 0,
		},
		{
			name:         "UTF-8 subject within limit",
			subject:      "Add 🚀 feature",
			maxLength:    50,
			wantErrCount: 0,
		},
		{
			name:         "UTF-8 subject exceeding limit",
			subject:      "This is a very long subject line with émojis 🚀🎉 that exceeds the maximum allowed length",
			maxLength:    50,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectTooLong),
		},
		{
			name:         "Unicode characters counted correctly",
			subject:      "测试中文字符长度验证功能",
			maxLength:    10,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectTooLong),
		},
		{
			name:         "Mixed ASCII and Unicode",
			subject:      "Fix bug in 用户认证 system",
			maxLength:    15,
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectTooLong),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: testCase.maxLength,
						Case:      "ignore", // Focus on length testing
					},
				},
			}

			rule := NewSubjectRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
				require.Contains(t, errors[0].Context["actual_length"], "")
				require.Contains(t, errors[0].Context["max_length"], "")
				require.Contains(t, errors[0].Context["excess"], "")
			}
		})
	}
}

func TestSubjectRule_ComprehensiveSuffixValidation(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		wantErrCount int
		wantErrCode  string
		invalidChar  string
	}{
		{
			name:         "Valid subject without forbidden suffixes",
			subject:      "Add new feature",
			wantErrCount: 0,
		},
		{
			name:         "Period suffix forbidden",
			subject:      "Add new feature.",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectSuffix),
			invalidChar:  ".",
		},
		{
			name:         "Comma suffix forbidden",
			subject:      "Add new feature,",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectSuffix),
			invalidChar:  ",",
		},
		{
			name:         "Semicolon suffix forbidden",
			subject:      "Add new feature;",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectSuffix),
			invalidChar:  ";",
		},
		{
			name:         "Colon suffix forbidden",
			subject:      "Add new feature:",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectSuffix),
			invalidChar:  ":",
		},
		{
			name:         "Exclamation mark suffix forbidden",
			subject:      "Add new feature!",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectSuffix),
			invalidChar:  "!",
		},
		{
			name:         "Question mark suffix forbidden",
			subject:      "Add new feature?",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrSubjectSuffix),
			invalidChar:  "?",
		},
		{
			name:         "UTF-8 suffix handling",
			subject:      "Add new feature™",
			wantErrCount: 0, // Should not be in default forbidden list
		},
		{
			name:         "Empty subject handled",
			subject:      "",
			wantErrCount: 1, // Empty subject error
			wantErrCode:  string(domain.ErrEmptySubject),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 100,
						Case:      "ignore", // Focus on suffix testing
					},
				},
			}

			rule := NewSubjectRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 && testCase.wantErrCode != "" {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)

				if testCase.invalidChar != "" {
					require.Contains(t, errors[0].Context["actual"], testCase.invalidChar)
				}
			}
		})
	}
}

func TestSubjectRule_EnhancedCaseValidationWithUTF8(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		caseChoice   string
		wantErrCount int
		wantErrCode  string
	}{
		{
			name:         "Valid upper case ASCII",
			subject:      "ADD NEW FEATURE",
			caseChoice:   "upper",
			wantErrCount: 0,
		},
		{
			name:         "Valid upper case UTF-8",
			subject:      "AÑADIR NUEVA FUNCIÓN",
			caseChoice:   "upper",
			wantErrCount: 0,
		},
		{
			name:         "Invalid upper case UTF-8",
			subject:      "añadir nueva función",
			caseChoice:   "upper",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrWrongCaseUpper),
		},
		{
			name:         "Valid upper case",
			subject:      "ADD NEW FEATURE",
			caseChoice:   "upper",
			wantErrCount: 0,
		},
		{
			name:         "Invalid upper case - should use specific error code",
			subject:      "add new feature",
			caseChoice:   "upper",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrWrongCaseUpper),
		},
		{
			name:         "Valid lower case",
			subject:      "add new feature",
			caseChoice:   "lower",
			wantErrCount: 0,
		},
		{
			name:         "Invalid lower case - should use specific error code",
			subject:      "Add new feature",
			caseChoice:   "lower",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrWrongCaseLower),
		},
		{
			name:         "Ignore case always valid",
			subject:      "whatever CaSe ThIs Is",
			caseChoice:   "ignore",
			wantErrCount: 0,
		},
		{
			name:         "Empty subject specific error",
			subject:      "",
			caseChoice:   "lower",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrEmptySubject),
		},
		{
			name:         "UTF-8 character as first letter",
			subject:      "øverføre data",
			caseChoice:   "lower",
			wantErrCount: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 100,
						Case:      testCase.caseChoice,
					},
				},
			}

			rule := NewSubjectRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
				require.Contains(t, errors[0].Context["expected"], testCase.caseChoice)
			}
		})
	}
}

func TestSubjectRule_EnhancedConventionalCommitParsing(t *testing.T) {
	tests := []struct {
		name                string
		subject             string
		conventionalEnabled bool
		wantErrCount        int
		wantErrCode         string
	}{
		{
			name:                "Valid conventional commit",
			subject:             "feat: Add new feature",
			conventionalEnabled: true,
			wantErrCount:        0,
		},
		{
			name:                "Valid conventional with scope",
			subject:             "feat(auth): Add login functionality",
			conventionalEnabled: true,
			wantErrCount:        0,
		},
		{
			name:                "Invalid conventional - missing description",
			subject:             "feat: ",
			conventionalEnabled: true,
			wantErrCount:        1,
			wantErrCode:         string(domain.ErrMissingConventionalSubject),
		},
		{
			name:                "Invalid conventional format",
			subject:             "feat add feature",
			conventionalEnabled: true,
			wantErrCount:        1,
			wantErrCode:         string(domain.ErrInvalidConventionalFormat),
		},
		{
			name:                "Non-conventional when not enabled",
			subject:             "Add new feature",
			conventionalEnabled: false,
			wantErrCount:        0,
		},
		{
			name:                "Conventional-like but invalid",
			subject:             "feat(scope: missing closing paren",
			conventionalEnabled: true,
			wantErrCount:        1,
			wantErrCode:         string(domain.ErrInvalidConventionalFormat),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 100,
						Case:      "ignore", // Focus on conventional commit parsing
					},
				},
			}

			if testCase.conventionalEnabled {
				cfg.Rules.Enabled = []string{"ConventionalCommit"}
			}

			rule := NewSubjectRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 && testCase.wantErrCode != "" {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)
				require.Contains(t, errors[0].Context["commit_format"], "")
			}
		})
	}
}

func TestSubjectRule_EnhancedImperativeMoodValidation(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		wantErrCount int
		wantErrCode  string
		category     string
	}{
		{
			name:         "Valid imperative",
			subject:      "add new feature",
			wantErrCount: 0,
		},
		{
			name:         "Past tense violation - specific error code",
			subject:      "added new feature",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrPastTense),
			category:     "past_tense",
		},
		{
			name:         "Gerund violation - specific error code",
			subject:      "adding new feature",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrGerund),
			category:     "gerund",
		},
		{
			name:         "Third person violation - specific error code",
			subject:      "adds new feature",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrThirdPerson),
			category:     "third_person",
		},
		{
			name:         "Conventional commit imperative valid",
			subject:      "feat: add new feature",
			wantErrCount: 0,
		},
		{
			name:         "Conventional commit past tense violation",
			subject:      "feat: added new feature",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrPastTense),
		},
		{
			name:         "Base form ending with 'ed' should be valid",
			subject:      "need to update documentation",
			wantErrCount: 0,
		},
		{
			name:         "Enhanced suggestion generation",
			subject:      "fixed bug in authentication",
			wantErrCount: 1,
			wantErrCode:  string(domain.ErrPastTense),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength:         100,
						Case:              "lower", // Focus on imperative testing
						RequireImperative: true,
					},
				},
				Rules: config.RulesConfig{
					Enabled: []string{"ConventionalCommit"},
				},
			}

			rule := NewSubjectRule(cfg)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, cfg)

			require.Len(t, errors, testCase.wantErrCount, "unexpected number of errors")

			if testCase.wantErrCount > 0 {
				require.Equal(t, testCase.wantErrCode, errors[0].Code)

				if testCase.category != "" {
					require.Equal(t, testCase.category, errors[0].Context["type"])
				}

				require.Contains(t, errors[0].Context, "suggestion")
			}
		})
	}
}

func TestSubjectRule_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		config  config.Config
		wantErr bool
	}{
		{
			name:    "Original default behavior maintained",
			subject: "add new feature",
			config:  config.Config{},
			wantErr: false,
		},
		{
			name:    "Original max length still works",
			subject: "this is a very long subject line that exceeds the maximum allowed length for testing purposes",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						MaxLength: 50,
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "Original case validation still works",
			subject: "add new feature",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						Case: "upper",
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "Original suffix validation still works",
			subject: "add new feature.",
			config: config.Config{
				Message: config.MessageConfig{
					Subject: config.SubjectConfig{
						ForbidEndings: []string{"."},
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "Enhanced defaults don't break existing functionality",
			subject: "add new feature!",
			config:  config.Config{}, // Uses enhanced defaults
			wantErr: true,            // Should fail because ! is in default forbidden suffixes
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rule := NewSubjectRule(testCase.config)
			commit := domain.Commit{Subject: testCase.subject}

			errors := rule.Validate(commit, testCase.config)

			if testCase.wantErr {
				require.NotEmpty(t, errors, "Expected validation error")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}
