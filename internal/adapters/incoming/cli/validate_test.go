// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/application/report"
	"github.com/itiquette/gommitlint/internal/domain"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// TestValidationParameters tests the functional parameters type.
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

	// Set specific flag values for testing
	err := cmd.Flags().Set("git-reference", "HEAD")
	require.NoError(t, err, "Failed to set git-reference flag")
	err = cmd.Flags().Set("verbose", "true")
	require.NoError(t, err, "Failed to set verbose flag")
	err = cmd.Flags().Set("light-mode", "true")
	require.NoError(t, err, "Failed to set light-mode flag")
	err = cmd.Flags().Set("format", "json")
	require.NoError(t, err, "Failed to set format flag")
	err = cmd.Flags().Set("rulehelp", "test-rule")
	require.NoError(t, err, "Failed to set rulehelp flag")

	// Create the parameters object
	params := NewValidationParameters(cmd)

	// Since we no longer store context in ValidationParameters, we don't need to test context operations

	// Test flag extraction
	require.Equal(t, "HEAD", params.GitReference)
	require.True(t, params.Verbose)
	require.False(t, params.ExtraVerbose)
	require.True(t, params.LightMode)
	require.Equal(t, "json", params.Format)
	require.Equal(t, "test-rule", params.RuleHelp)
	require.True(t, params.SkipMergeCommits)

	// Test conversion to validation options
	opts, err := params.ToValidationOptions()
	require.NoError(t, err)
	require.Equal(t, "HEAD", opts.CommitHash)
	require.True(t, opts.SkipMergeCommits)

	// Test conversion to report options
	reportOpts := params.ToReportOptions()
	require.Equal(t, report.FormatJSON, reportOpts.Format)
	require.True(t, reportOpts.Verbose)
	require.True(t, reportOpts.ShowHelp)
	require.Equal(t, "test-rule", reportOpts.RuleToShowHelp)
	require.True(t, reportOpts.LightMode)

	// Test formatter creation
	formatter := params.CreateFormatter()
	require.NotNil(t, formatter)

	// Test with various validation options scenarios
	t.Run("revision range conversion", func(t *testing.T) {
		cmdRange := &cobra.Command{}
		cmdRange.Flags().String("revision-range", "", "")
		err := cmdRange.Flags().Set("revision-range", "main..feature")
		require.NoError(t, err, "Failed to set revision-range flag")

		rangeParams := NewValidationParameters(cmdRange)
		opts, err := rangeParams.ToValidationOptions()
		require.NoError(t, err)
		require.Equal(t, "main", opts.FromHash)
		require.Equal(t, "feature", opts.ToHash)
	})

	t.Run("invalid revision range", func(t *testing.T) {
		cmdRange := &cobra.Command{}
		cmdRange.Flags().String("revision-range", "", "")
		err := cmdRange.Flags().Set("revision-range", "invalid-format")
		require.NoError(t, err, "Failed to set revision-range flag")

		rangeParams := NewValidationParameters(cmdRange)

		var valErr error
		_, valErr = rangeParams.ToValidationOptions()
		require.Error(t, valErr)
		require.Contains(t, valErr.Error(), "invalid revision range format")
	})

	t.Run("base branch conversion", func(t *testing.T) {
		cmdBase := &cobra.Command{}
		cmdBase.Flags().String("base-branch", "", "")
		err := cmdBase.Flags().Set("base-branch", "develop")
		require.NoError(t, err, "Failed to set base-branch flag")

		baseParams := NewValidationParameters(cmdBase)
		opts, err := baseParams.ToValidationOptions()
		require.NoError(t, err)
		require.Equal(t, "develop", opts.FromHash)
		require.Equal(t, "HEAD", opts.ToHash)
	})

	t.Run("message file has priority", func(t *testing.T) {
		cmdMulti := &cobra.Command{}
		cmdMulti.Flags().String("message-file", "", "")
		cmdMulti.Flags().String("git-reference", "", "")
		cmdMulti.Flags().String("base-branch", "", "")

		err := cmdMulti.Flags().Set("message-file", "commit.txt")
		require.NoError(t, err, "Failed to set message-file flag")
		err = cmdMulti.Flags().Set("git-reference", "abc123")
		require.NoError(t, err, "Failed to set git-reference flag")
		err = cmdMulti.Flags().Set("base-branch", "main")
		require.NoError(t, err, "Failed to set base-branch flag")

		multiParams := NewValidationParameters(cmdMulti)
		opts, err := multiParams.ToValidationOptions()
		require.NoError(t, err)
		require.Equal(t, "commit.txt", opts.MessageFile)
		require.Equal(t, "", opts.CommitHash) // Should be empty because MessageFile takes precedence
		require.Equal(t, "", opts.FromHash)   // Should be empty because MessageFile takes precedence
	})
}

// TestValidationParametersFormatters tests the different formatters.
func TestValidationParametersFormatters(t *testing.T) {
	formats := []struct {
		name       string
		formatFlag string
	}{
		{
			name:       "text formatter",
			formatFlag: "text",
		},
		{
			name:       "json formatter",
			formatFlag: "json",
		},
		{
			name:       "github formatter",
			formatFlag: "github",
		},
		{
			name:       "gitlab formatter",
			formatFlag: "gitlab",
		},
		{
			name:       "default formatter for unknown format",
			formatFlag: "unknown",
		},
	}

	for _, testCase := range formats {
		t.Run(testCase.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("format", "", "")
			cmd.Flags().Bool("verbose", false, "")
			cmd.Flags().Bool("extra-verbose", false, "")
			cmd.Flags().String("rulehelp", "", "")
			cmd.Flags().Bool("light-mode", false, "")

			err := cmd.Flags().Set("format", testCase.formatFlag)
			require.NoError(t, err, "Failed to set format flag")

			params := NewValidationParameters(cmd)
			formatter := params.CreateFormatter()
			require.NotNil(t, formatter)

			// Test that it implements the domain.ResultFormatter interface
			ctx := testcontext.CreateTestContext()
			formatted := formatter.Format(ctx, domain.ValidationResults{})
			require.NotNil(t, formatted)
		})
	}
}
