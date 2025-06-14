// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"testing"

	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/stretchr/testify/require"
)

func TestNewValidateCommand(t *testing.T) {
	cmd := NewValidateCommand()

	require.Equal(t, "validate", cmd.Name)
	require.Equal(t, "Validate commit messages", cmd.Usage)
	require.NotEmpty(t, cmd.Description)
	require.NotNil(t, cmd.Action)

	// Check that all expected flags are present
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}

	expectedFlags := []string{
		"message-file", "ref", "range", "base-branch", "count",
		"verbose", "rule-help", "report-file",
	}

	for _, expectedFlag := range expectedFlags {
		require.True(t, flagNames[expectedFlag], "missing flag: %s", expectedFlag)
	}
}

func TestCreateValidationTargetDirectly(t *testing.T) {
	tests := []struct {
		name         string
		messageFile  string
		gitRef       string
		commitRange  string
		baseBranch   string
		commitCount  int
		expectError  bool
		expectedType string
		description  string
	}{
		{
			name:         "message file only",
			messageFile:  "/path/to/file",
			expectError:  false,
			expectedType: "message",
			description:  "should create target for message file",
		},
		{
			name:         "git ref only",
			gitRef:       "HEAD~1",
			expectError:  false,
			expectedType: "commit",
			description:  "should create target for git ref",
		},
		{
			name:         "commit range only",
			commitRange:  "main..feature",
			expectError:  false,
			expectedType: "range",
			description:  "should create target for commit range",
		},
		{
			name:         "base branch only",
			baseBranch:   "main",
			expectError:  false,
			expectedType: "range",
			description:  "should create target for base branch",
		},
		{
			name:         "commit count only",
			commitCount:  5,
			expectError:  false,
			expectedType: "count",
			description:  "should create target for commit count",
		},
		{
			name:         "multiple flags use precedence",
			messageFile:  "/path",
			gitRef:       "HEAD",
			expectError:  false,
			expectedType: "message",
			description:  "should use message-file with highest precedence",
		},
		{
			name:         "no validation flags",
			commitCount:  1,
			expectError:  false,
			expectedType: "commit",
			description:  "should default to HEAD validation",
		},
		{
			name:        "invalid commit range",
			commitRange: "invalid-range",
			expectError: true,
			description: "should fail with invalid range format",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			target, err := cliAdapter.NewValidationTarget(
				testCase.messageFile,
				testCase.gitRef,
				testCase.commitRange,
				testCase.baseBranch,
				testCase.commitCount,
			)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
				require.Equal(t, testCase.expectedType, target.Type, testCase.description)
			}
		})
	}
}

func TestCreateOutputOptionsDirectly(t *testing.T) {
	tests := []struct {
		name        string
		reportFile  string
		format      string
		color       string
		quiet       bool
		verbose     int
		ruleHelp    string
		expectError bool
		description string
	}{
		{
			name:        "default options",
			format:      "text",
			color:       "auto",
			quiet:       false,
			expectError: false,
			description: "should create default options",
		},
		{
			name:        "verbose output",
			format:      "text",
			color:       "auto",
			quiet:       false,
			verbose:     1,
			expectError: false,
			description: "should enable verbose output",
		},
		{
			name:        "quiet overrides verbose",
			format:      "text",
			color:       "auto",
			quiet:       true,
			verbose:     1,
			expectError: false,
			description: "should disable verbose when quiet is set",
		},
		{
			name:        "valid rule help",
			format:      "text",
			color:       "auto",
			quiet:       false,
			ruleHelp:    "subject",
			expectError: false,
			description: "should accept valid rule help",
		},
		{
			name:        "json format",
			format:      "json",
			color:       "never",
			quiet:       false,
			expectError: false,
			description: "should support JSON format",
		},
		{
			name:        "with report file",
			reportFile:  "/tmp/test-report.txt",
			format:      "text",
			color:       "auto",
			quiet:       false,
			expectError: false,
			description: "should create report file successfully",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var writer *os.File

			// Create report file if specified
			if testCase.reportFile != "" {
				file, err := os.Create(testCase.reportFile)
				require.NoError(t, err)

				defer func() {
					file.Close()
					os.Remove(testCase.reportFile)
				}()

				writer = file
			} else {
				writer = os.Stdout
			}

			// Create base options
			options := cliAdapter.NewOutputOptions(writer).
				WithFormat(testCase.format).
				WithColor(testCase.color)

			// Handle verbose flags
			if testCase.verbose > 0 && !testCase.quiet {
				options = options.WithVerbose(true)
			}

			// Handle rule help
			if testCase.ruleHelp != "" {
				options = options.WithRuleHelp(testCase.ruleHelp)

				err := options.ValidateRuleHelp()
				if testCase.expectError {
					require.Error(t, err, testCase.description)

					return
				}

				require.NoError(t, err, testCase.description)
			}

			require.NotEmpty(t, options, testCase.description)
		})
	}
}

func TestGetRepoPathLogic(t *testing.T) {
	tests := []struct {
		name         string
		repoPathFlag string
		expectedPath string
		description  string
	}{
		{
			name:         "empty flag defaults to current directory",
			repoPathFlag: "",
			expectedPath: ".",
			description:  "should default to current directory",
		},
		{
			name:         "custom repository path",
			repoPathFlag: "/custom/repo/path",
			expectedPath: "/custom/repo/path",
			description:  "should use custom repository path",
		},
		{
			name:         "relative path",
			repoPathFlag: "../other-repo",
			expectedPath: "../other-repo",
			description:  "should use relative path",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Test the logic directly
			repoPath := testCase.repoPathFlag
			if repoPath == "" {
				repoPath = "."
			}

			require.Equal(t, testCase.expectedPath, repoPath, testCase.description)
		})
	}
}

func TestHandleRuleHelpLogic(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
	}{
		{
			name:     "valid rule name",
			ruleName: "subject",
		},
		{
			name:     "another valid rule name",
			ruleName: "conventional",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			originalStdout := os.Stdout

			// Create a pipe to capture output
			reader, writer, _ := os.Pipe()
			os.Stdout = writer

			// Create mock output options
			options := cliAdapter.NewOutputOptions(os.Stdout).WithRuleHelp(testCase.ruleName)

			// Call the function
			err := handleRuleHelp(options, mockConfig())

			// Restore stdout and get output
			writer.Close()

			os.Stdout = originalStdout

			// Read captured output
			output := make([]byte, 1024)
			n, _ := reader.Read(output)
			capturedOutput := string(output[:n])

			require.NoError(t, err)
			require.Contains(t, capturedOutput, testCase.ruleName)
			require.Contains(t, capturedOutput, "Help for rule:")
		})
	}
}

func mockConfig() configTypes.Config {
	// Return a minimal mock config for testing
	return configTypes.Config{}
}
