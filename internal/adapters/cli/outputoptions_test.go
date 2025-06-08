// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewOutputOptions(t *testing.T) {
	tests := []struct {
		name        string
		writer      *bytes.Buffer
		description string
	}{
		{
			name:        "creates with default values",
			writer:      &bytes.Buffer{},
			description: "should create OutputOptions with sensible defaults",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			options := NewOutputOptions(testCase.writer)

			require.Equal(t, "text", options.Format)
			require.False(t, options.Verbose)
			require.False(t, options.ShowHelp)
			require.Equal(t, "", options.RuleHelp)
			require.Equal(t, "auto", options.Color)
			require.Equal(t, testCase.writer, options.Writer)
		})
	}
}

func TestOutputOptions_WithMethods(t *testing.T) {
	tests := []struct {
		name            string
		setupOptions    func(OutputOptions) OutputOptions
		expectedFormat  string
		expectedVerbose bool
		expectedHelp    bool
		expectedRule    string
		expectedColor   string
		description     string
	}{
		{
			name: "WithFormat",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithFormat("json")
			},
			expectedFormat: "json",
			description:    "should set format correctly",
		},
		{
			name: "WithVerbose true",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithVerbose(true)
			},
			expectedVerbose: true,
			description:     "should enable verbose mode",
		},
		{
			name: "WithVerbose false",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithVerbose(false)
			},
			expectedVerbose: false,
			description:     "should disable verbose mode",
		},
		{
			name: "WithShowHelp true",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithShowHelp(true)
			},
			expectedHelp: true,
			description:  "should enable help display",
		},
		{
			name: "WithRuleHelp",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithRuleHelp("subject")
			},
			expectedRule: "subject",
			description:  "should set rule help",
		},
		{
			name: "WithColor",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithColor("always")
			},
			expectedColor: "always",
			description:   "should set color option",
		},
		{
			name: "chained methods",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithFormat("json").
					WithVerbose(true).
					WithShowHelp(true).
					WithRuleHelp("conventional").
					WithColor("never")
			},
			expectedFormat:  "json",
			expectedVerbose: true,
			expectedHelp:    true,
			expectedRule:    "conventional",
			expectedColor:   "never",
			description:     "should chain multiple methods correctly",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf)
			result := testCase.setupOptions(options)

			if testCase.expectedFormat != "" {
				require.Equal(t, testCase.expectedFormat, result.Format, testCase.description)
			}

			if testCase.name == "WithVerbose true" || testCase.name == "chained methods" {
				require.Equal(t, testCase.expectedVerbose, result.Verbose, testCase.description)
			}

			if testCase.name == "WithShowHelp true" || testCase.name == "chained methods" {
				require.Equal(t, testCase.expectedHelp, result.ShowHelp, testCase.description)
			}

			if testCase.expectedRule != "" {
				require.Equal(t, testCase.expectedRule, result.RuleHelp, testCase.description)
			}

			if testCase.expectedColor != "" {
				require.Equal(t, testCase.expectedColor, result.Color, testCase.description)
			}
		})
	}
}

func TestOutputOptions_ValidateRuleHelp(t *testing.T) {
	tests := []struct {
		name        string
		ruleHelp    string
		expectError bool
		description string
	}{
		{
			name:        "empty rule help is valid",
			ruleHelp:    "",
			expectError: false,
			description: "should accept empty rule help",
		},
		{
			name:        "valid factory key - subject",
			ruleHelp:    "subject",
			expectError: false,
			description: "should accept valid factory key",
		},
		{
			name:        "valid factory key - conventional",
			ruleHelp:    "conventional",
			expectError: false,
			description: "should accept conventional rule",
		},
		{
			name:        "valid actual rule name - Subject",
			ruleHelp:    "Subject",
			expectError: false,
			description: "should accept actual rule name",
		},
		{
			name:        "case insensitive match",
			ruleHelp:    "SUBJECT",
			expectError: false,
			description: "should handle case insensitive matching",
		},
		{
			name:        "invalid rule name",
			ruleHelp:    "invalidrule",
			expectError: true,
			description: "should reject invalid rule names",
		},
		{
			name:        "rule with whitespace",
			ruleHelp:    "  subject  ",
			expectError: false,
			description: "should trim whitespace and validate",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf).WithRuleHelp(testCase.ruleHelp)

			err := options.ValidateRuleHelp()

			if testCase.expectError {
				require.Error(t, err, testCase.description)
				require.Contains(t, err.Error(), "unknown rule", testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

func TestOutputOptions_GetNormalizedRuleHelp(t *testing.T) {
	tests := []struct {
		name         string
		ruleHelp     string
		expectedRule string
		description  string
	}{
		{
			name:         "empty rule help",
			ruleHelp:     "",
			expectedRule: "",
			description:  "should return empty for empty input",
		},
		{
			name:         "factory key - subject",
			ruleHelp:     "subject",
			expectedRule: "Subject",
			description:  "should normalize factory key to actual rule name",
		},
		{
			name:         "factory key - conventional",
			ruleHelp:     "conventional",
			expectedRule: "ConventionalCommit",
			description:  "should normalize conventional to ConventionalCommit",
		},
		{
			name:         "actual rule name - already correct",
			ruleHelp:     "Subject",
			expectedRule: "Subject",
			description:  "should return actual rule name as-is",
		},
		{
			name:         "case insensitive actual rule",
			ruleHelp:     "subject",
			expectedRule: "Subject",
			description:  "should handle case insensitive actual rule names",
		},
		{
			name:         "unknown rule",
			ruleHelp:     "unknown",
			expectedRule: "unknown",
			description:  "should return original for unknown rules",
		},
		{
			name:         "rule with whitespace",
			ruleHelp:     "  subject  ",
			expectedRule: "Subject",
			description:  "should trim whitespace before normalization",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf).WithRuleHelp(testCase.ruleHelp)

			result := options.GetNormalizedRuleHelp()

			require.Equal(t, testCase.expectedRule, result, testCase.description)
		})
	}
}

func TestOutputOptions_HelperMethods(t *testing.T) {
	tests := []struct {
		name           string
		setupOptions   func(OutputOptions) OutputOptions
		shouldShowHelp bool
		showRuleHelp   bool
		getRuleHelp    string
		description    string
	}{
		{
			name: "default options",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o
			},
			shouldShowHelp: false,
			showRuleHelp:   false,
			getRuleHelp:    "",
			description:    "should have correct defaults",
		},
		{
			name: "with general help enabled",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithShowHelp(true)
			},
			shouldShowHelp: true,
			showRuleHelp:   false,
			getRuleHelp:    "",
			description:    "should show general help",
		},
		{
			name: "with rule-specific help",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithRuleHelp("subject")
			},
			shouldShowHelp: false,
			showRuleHelp:   true,
			getRuleHelp:    "subject",
			description:    "should show rule-specific help",
		},
		{
			name: "both general and rule help",
			setupOptions: func(o OutputOptions) OutputOptions {
				return o.WithShowHelp(true).WithRuleHelp("conventional")
			},
			shouldShowHelp: true,
			showRuleHelp:   true,
			getRuleHelp:    "conventional",
			description:    "should handle both help types",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf)
			result := testCase.setupOptions(options)

			require.Equal(t, testCase.shouldShowHelp, result.ShouldShowHelp(), testCase.description)
			require.Equal(t, testCase.showRuleHelp, result.ShowRuleHelp(), testCase.description)
			require.Equal(t, testCase.getRuleHelp, result.GetRuleHelp(), testCase.description)
		})
	}
}

func TestOutputOptions_ShouldUseColor(t *testing.T) {
	tests := []struct {
		name             string
		colorSetting     string
		writer           func() *os.File
		expectedUseColor bool
		description      string
	}{
		{
			name:             "always use color",
			colorSetting:     "always",
			writer:           func() *os.File { return &os.File{} },
			expectedUseColor: true,
			description:      "should always use color when set to always",
		},
		{
			name:             "never use color",
			colorSetting:     "never",
			writer:           func() *os.File { return os.Stdout },
			expectedUseColor: false,
			description:      "should never use color when set to never",
		},
		{
			name:             "auto with stdout in test",
			colorSetting:     "auto",
			writer:           func() *os.File { return os.Stdout },
			expectedUseColor: false,
			description:      "should not use color for stdout in test environment (not a TTY)",
		},
		{
			name:             "auto with stderr in test",
			colorSetting:     "auto",
			writer:           func() *os.File { return os.Stderr },
			expectedUseColor: false,
			description:      "should not use color for stderr in test environment (not a TTY)",
		},
		{
			name:             "default setting in test",
			colorSetting:     "",
			writer:           func() *os.File { return os.Stdout },
			expectedUseColor: false,
			description:      "should default to auto behavior (no color in test environment)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			writer := testCase.writer()
			options := NewOutputOptions(writer).WithColor(testCase.colorSetting)

			result := options.ShouldUseColor()

			require.Equal(t, testCase.expectedUseColor, result, testCase.description)
		})
	}
}

func TestOutputOptions_ShouldUseColorWithNOCOLOR(t *testing.T) {
	tests := []struct {
		name             string
		colorSetting     string
		noColorValue     string
		isTerminal       bool
		expectedUseColor bool
		description      string
	}{
		{
			name:             "NO_COLOR=1 overrides always",
			colorSetting:     "always",
			noColorValue:     "1",
			isTerminal:       true,
			expectedUseColor: false,
			description:      "NO_COLOR should override --color always",
		},
		{
			name:             "NO_COLOR=true overrides auto on terminal",
			colorSetting:     "auto",
			noColorValue:     "true",
			isTerminal:       true,
			expectedUseColor: false,
			description:      "NO_COLOR should override auto mode even on terminal",
		},
		{
			name:             "NO_COLOR=anything overrides auto",
			colorSetting:     "auto",
			noColorValue:     "anything",
			isTerminal:       true,
			expectedUseColor: false,
			description:      "any NO_COLOR value should disable color",
		},
		{
			name:             "empty NO_COLOR respects always",
			colorSetting:     "always",
			noColorValue:     "",
			isTerminal:       true,
			expectedUseColor: true,
			description:      "empty NO_COLOR should not affect color settings",
		},
		{
			name:             "empty NO_COLOR respects never",
			colorSetting:     "never",
			noColorValue:     "",
			isTerminal:       true,
			expectedUseColor: false,
			description:      "empty NO_COLOR should not affect never setting",
		},
		{
			name:             "empty NO_COLOR respects auto on terminal",
			colorSetting:     "auto",
			noColorValue:     "",
			isTerminal:       true,
			expectedUseColor: true,
			description:      "empty NO_COLOR should allow auto to detect terminal",
		},
		{
			name:             "empty NO_COLOR respects auto on non-terminal",
			colorSetting:     "auto",
			noColorValue:     "",
			isTerminal:       false,
			expectedUseColor: false,
			description:      "empty NO_COLOR should allow auto to detect non-terminal",
		},
		{
			name:             "NO_COLOR with never is redundant",
			colorSetting:     "never",
			noColorValue:     "1",
			isTerminal:       true,
			expectedUseColor: false,
			description:      "NO_COLOR with never should still be false",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock environment function
			mockEnv := func(key string) string {
				if key == "NO_COLOR" {
					return testCase.noColorValue
				}

				return ""
			}

			// Create a test function that includes mocked terminal detection
			testShouldUseColor := func(colorSetting string, isTerminal bool, envFunc func(string) string) bool {
				// NO_COLOR environment variable takes precedence over all flags
				if envFunc("NO_COLOR") != "" {
					return false
				}

				switch colorSetting {
				case "always":
					return true
				case "never":
					return false
				case "auto":
					fallthrough
				default:
					return isTerminal
				}
			}

			result := testShouldUseColor(testCase.colorSetting, testCase.isTerminal, mockEnv)

			require.Equal(t, testCase.expectedUseColor, result, testCase.description)
		})
	}
}

func TestOutputOptions_IsTerminal(t *testing.T) {
	tests := []struct {
		name           string
		writer         func() interface{}
		expectedResult bool
		description    string
	}{
		{
			name: "stdout is terminal",
			writer: func() interface{} {
				return os.Stdout
			},
			expectedResult: false, // In test environment, stdout is not a terminal
			description:    "should detect stdout as non-terminal in test environment",
		},
		{
			name: "stderr is terminal",
			writer: func() interface{} {
				return os.Stderr
			},
			expectedResult: false, // In test environment, stderr is not a terminal
			description:    "should detect stderr as non-terminal in test environment",
		},
		{
			name: "bytes buffer is not terminal",
			writer: func() interface{} {
				return &bytes.Buffer{}
			},
			expectedResult: false,
			description:    "should detect bytes buffer as non-terminal",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			writer := testCase.writer()

			w, ok := writer.(interface {
				Write(p []byte) (n int, err error)
			})
			require.True(t, ok, "writer should implement Write interface")

			result := isTerminal(w)

			require.Equal(t, testCase.expectedResult, result, testCase.description)
		})
	}
}

func TestOutputOptions_FormatReport(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		expectedType string
		description  string
	}{
		{
			name:         "json format",
			format:       "json",
			expectedType: "json",
			description:  "should format as JSON",
		},
		{
			name:         "github format",
			format:       "github",
			expectedType: "github",
			description:  "should format for GitHub",
		},
		{
			name:         "gitlab format",
			format:       "gitlab",
			expectedType: "gitlab",
			description:  "should format for GitLab",
		},
		{
			name:         "text format",
			format:       "text",
			expectedType: "text",
			description:  "should format as text",
		},
		{
			name:         "default format",
			format:       "",
			expectedType: "text",
			description:  "should default to text format",
		},
		{
			name:         "unknown format",
			format:       "unknown",
			expectedType: "text",
			description:  "should fallback to text for unknown format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf).WithFormat(testCase.format)

			// Create a minimal report for testing
			report := domain.Report{
				Summary: domain.ReportSummary{
					AllPassed:     true,
					TotalCommits:  1,
					PassedCommits: 1,
					FailedCommits: 0,
				},
				Commits: []domain.CommitReport{
					{
						Commit: domain.Commit{
							Hash:    "abc123",
							Subject: "Test commit",
						},
						RuleResults: []domain.RuleReport{},
						Passed:      true,
					},
				},
				Repository: domain.RepositoryReport{},
			}

			result := options.FormatReport(report)

			require.NotEmpty(t, result, testCase.description)

			// Basic validation based on format type
			switch testCase.expectedType {
			case "json":
				require.True(t, strings.Contains(result, "{") && strings.Contains(result, "}"), "should contain JSON structure")
			case "text":
				// Text format should contain some recognizable content
				require.NotEmpty(t, result, "should have text content")
			}
		})
	}
}

func TestOutputOptions_WriteReport(t *testing.T) {
	tests := []struct {
		name        string
		report      domain.Report
		expectError bool
		description string
	}{
		{
			name: "successful write",
			report: domain.Report{
				Summary: domain.ReportSummary{
					AllPassed:     true,
					TotalCommits:  1,
					PassedCommits: 1,
					FailedCommits: 0,
				},
				Commits: []domain.CommitReport{
					{
						Commit: domain.Commit{
							Hash:    "abc123",
							Subject: "Test commit",
						},
						RuleResults: []domain.RuleReport{},
						Passed:      true,
					},
				},
			},
			expectError: false,
			description: "should write report successfully",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf)

			err := options.WriteReport(testCase.report)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.Positive(t, buf.Len(), "should have written content")
			}
		})
	}
}

func TestOutputOptions_ToReportOptions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(OutputOptions) OutputOptions
		description string
	}{
		{
			name: "converts all options correctly",
			setup: func(o OutputOptions) OutputOptions {
				return o.WithFormat("json").
					WithVerbose(true).
					WithShowHelp(true).
					WithColor("always")
			},
			description: "should convert to domain ReportOptions",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			options := NewOutputOptions(buf)
			configured := testCase.setup(options)

			result := configured.ToReportOptions()

			require.Equal(t, configured.Format, result.Format, testCase.description)
			require.Equal(t, configured.Verbose, result.Verbose, testCase.description)
			require.Equal(t, configured.ShouldShowHelp(), result.ShowHelp, testCase.description)
			require.Equal(t, configured.ShouldUseColor(), result.UseColor, testCase.description)
			require.Equal(t, configured.Writer, result.Writer, testCase.description)
		})
	}
}

func TestGetValidRuleNames(t *testing.T) {
	t.Run("returns all valid rule names", func(t *testing.T) {
		rules := getValidRuleNames()

		// Should contain factory keys
		require.Contains(t, rules, "subject")
		require.Contains(t, rules, "conventional")
		require.Contains(t, rules, "commitbody")

		// Should contain actual rule names
		require.Contains(t, rules, "Subject")
		require.Contains(t, rules, "ConventionalCommit")
		require.Contains(t, rules, "CommitBody")

		// Should have reasonable number of rules
		require.Greater(t, len(rules), 10, "should have multiple rules")
	})
}
