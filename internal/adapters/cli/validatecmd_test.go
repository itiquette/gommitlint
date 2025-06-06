// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
)

// TestValidationParameters tests the ValidationParameters type.
func TestValidationParameters(t *testing.T) {
	// Create a command for testing
	cmd := &cobra.Command{}

	// Add flags to the command
	cmd.Flags().String("message-file", "", "commit message file path to validate")
	cmd.Flags().String("git-reference", "", "git reference to validate (defaults to HEAD)")
	cmd.Flags().Int("commit-count", 1, "number of commits from HEAD to validate")
	cmd.Flags().String("revision-range", "", "range of commits to validate (<commit1>..<commit2>)")
	cmd.Flags().String("base-branch", "", "base branch to compare with (sets revision-range to <base-branch>..HEAD)")
	cmd.Flags().BoolP("verbose", "v", false, "show detailed validation results")
	cmd.Flags().Bool("extra-verbose", false, "show extra detailed validation results")
	cmd.Flags().Bool("light-mode", false, "use light background color scheme")
	cmd.Flags().String("rulehelp", "", "show detailed help for a specific rule")
	cmd.Flags().String("format", "text", "output format: text, json, github, or gitlab")
	cmd.Flags().Bool("skip-merge-commits", true, "skip merge commits in validation")
	cmd.Flags().String("repo-path", "", "path to the repository to validate")

	// Set some flags for testing
	require.NoError(t, cmd.Flags().Set("verbose", "true"))
	require.NoError(t, cmd.Flags().Set("light-mode", "true"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.Flags().Set("rulehelp", "test-rule"))

	// Create parameters
	params, err := NewValidateParams(cmd)
	require.NoError(t, err)

	// Verify that parameters are correctly set
	require.Equal(t, "", params.RepoPath)
	require.True(t, params.SkipMergeCommits)

	// Verify output options
	require.True(t, params.Output.Verbose)
	require.False(t, params.Output.ExtraVerbose)
	require.True(t, params.Output.LightMode)
	require.Equal(t, "json", params.Output.Format)
	require.Equal(t, "test-rule", params.Output.RuleHelp)

	// Verify target defaults to HEAD commit
	require.Equal(t, "commit", params.Target.Type)
	require.Equal(t, "HEAD", params.Target.Source)

	// Test conversion to report options
	reportOpts := params.ToReportOptions()
	require.Equal(t, "json", reportOpts.Format)
	require.True(t, reportOpts.Verbose)
	require.True(t, reportOpts.ShowHelp)
	require.True(t, reportOpts.LightMode)
}

// TestValidationParametersScenarios tests various validation scenarios.
func TestValidationParametersScenarios(t *testing.T) {
	// Test with various validation options scenarios
	t.Run("revision range conversion", func(t *testing.T) {
		cmdRange := &cobra.Command{}
		cmdRange.Flags().String("revision-range", "", "")
		err := cmdRange.Flags().Set("revision-range", "main..feature")
		require.NoError(t, err, "Failed to set revision-range flag")

		rangeParams, err := NewValidateParams(cmdRange)
		require.NoError(t, err)
		require.Equal(t, "range", rangeParams.Target.Type)
		require.Equal(t, "main", rangeParams.Target.Source)
		require.Equal(t, "feature", rangeParams.Target.Target)
	})

	t.Run("invalid revision range", func(t *testing.T) {
		cmdRange := &cobra.Command{}
		cmdRange.Flags().String("revision-range", "", "")
		err := cmdRange.Flags().Set("revision-range", "invalid-format")
		require.NoError(t, err, "Failed to set revision-range flag")

		_, err = NewValidateParams(cmdRange)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid revision range format")
	})

	t.Run("base branch conversion", func(t *testing.T) {
		cmdBase := &cobra.Command{}
		cmdBase.Flags().String("base-branch", "", "")
		err := cmdBase.Flags().Set("base-branch", "develop")
		require.NoError(t, err, "Failed to set base-branch flag")

		baseParams, err := NewValidateParams(cmdBase)
		require.NoError(t, err)
		require.Equal(t, "range", baseParams.Target.Type)
		require.Equal(t, "develop", baseParams.Target.Source)
		require.Equal(t, "HEAD", baseParams.Target.Target)
	})

	t.Run("message file takes precedence", func(t *testing.T) {
		cmdMulti := &cobra.Command{}
		cmdMulti.Flags().String("message-file", "", "")
		cmdMulti.Flags().String("git-reference", "", "")
		cmdMulti.Flags().String("base-branch", "", "")

		require.NoError(t, cmdMulti.Flags().Set("message-file", "msg.txt"))
		require.NoError(t, cmdMulti.Flags().Set("git-reference", "abc123"))
		require.NoError(t, cmdMulti.Flags().Set("base-branch", "main"))

		multiParams, err := NewValidateParams(cmdMulti)
		require.NoError(t, err)

		// Test that message-file takes precedence (highest priority)
		require.Equal(t, "message", multiParams.Target.Type)
		require.Equal(t, "msg.txt", multiParams.Target.Source)
	})
}

// TestFormatConversion tests conversion between formatter types.
func TestFormatConversion(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		expectedType   string
		expectVerbose  bool
		expectShowHelp bool
	}{
		{
			name:           "JSON formatter",
			format:         "json",
			expectedType:   "*output.JSONFormatter",
			expectVerbose:  false,
			expectShowHelp: false,
		},
		{
			name:           "GitHub formatter",
			format:         "github",
			expectedType:   "*output.GitHubFormatter",
			expectVerbose:  true,
			expectShowHelp: true,
		},
		{
			name:           "GitLab formatter",
			format:         "gitlab",
			expectedType:   "*output.GitLabFormatter",
			expectVerbose:  true,
			expectShowHelp: true,
		},
		{
			name:           "Text formatter",
			format:         "text",
			expectedType:   "*output.TextFormatter",
			expectVerbose:  true,
			expectShowHelp: true,
		},
		{
			name:           "Default formatter",
			format:         "unknown",
			expectedType:   "*output.TextFormatter",
			expectVerbose:  true,
			expectShowHelp: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("format", "text", "")
			cmd.Flags().Bool("verbose", false, "")
			cmd.Flags().Bool("extra-verbose", false, "")
			cmd.Flags().String("rulehelp", "", "")

			if testCase.expectVerbose {
				require.NoError(t, cmd.Flags().Set("verbose", "true"))
			}

			if testCase.expectShowHelp {
				require.NoError(t, cmd.Flags().Set("extra-verbose", "true"))
			}

			require.NoError(t, cmd.Flags().Set("format", testCase.format))

			params, err := NewValidateParams(cmd)
			require.NoError(t, err)

			// Test that the format functionality works
			mockReport := domain.Report{
				Metadata: domain.ReportMetadata{
					Timestamp: time.Now(),
				},
				Summary: domain.ReportSummary{
					TotalCommits:  1,
					PassedCommits: 1,
					AllPassed:     true,
				},
				Commits: []domain.CommitReport{
					{
						Commit: domain.Commit{
							Hash:    "abc123",
							Subject: "Test commit",
						},
						Passed:      true,
						RuleResults: []domain.RuleReport{},
					},
				},
			}

			output := params.Output.FormatReport(mockReport)
			require.NotEmpty(t, output)
		})
	}
}
