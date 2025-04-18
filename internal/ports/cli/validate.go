// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package cli provides command line interfaces for the application.
package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/application/validate"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/infrastructure/output"
)

// OutputFormat defines how validation results should be formatted.
type OutputFormat string

const (
	// OutputFormatText outputs results in human-readable text.
	OutputFormatText OutputFormat = "text"

	// OutputFormatJSON outputs results in JSON format.
	OutputFormatJSON OutputFormat = "json"

	// OutputFormatGitHubActions outputs results in GitHub Actions format.
	OutputFormatGitHubActions OutputFormat = "github"
)

// ValidateCommand handles the validate command.
type ValidateCommand struct {
	// Command line flags
	CommitRef        string
	CommitCount      int
	FromRef          string
	ToRef            string
	MessageFile      string
	OutputFormat     OutputFormat
	Verbose          bool
	SkipMergeCommits bool
	RepoPath         string
}

// NewValidateCommand creates a new ValidateCommand.
func NewValidateCommand() *ValidateCommand {
	return &ValidateCommand{
		OutputFormat:     OutputFormatText,
		SkipMergeCommits: true,
	}
}

// Execute runs the validation command.
func (c *ValidateCommand) Execute() error {
	// Create context
	ctx := context.Background()

	// Create validation service
	service, err := validate.CreateDefaultValidationService(c.RepoPath)
	if err != nil {
		return fmt.Errorf("failed to create validation service: %w", err)
	}

	// Convert command line flags to validation options
	opts := validate.ValidationOptions{
		CommitHash:       c.CommitRef,
		CommitCount:      c.CommitCount,
		FromHash:         c.FromRef,
		ToHash:           c.ToRef,
		MessageFile:      c.MessageFile,
		SkipMergeCommits: c.SkipMergeCommits,
	}

	// Validate according to options
	results, err := service.ValidateWithOptions(ctx, opts)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Output results
	if err := c.outputResults(results); err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	// Return exit code based on validation results
	if !results.AllPassed() {
		return fmt.Errorf("validation failed: %d/%d commits failed", results.TotalCommits-results.PassedCommits, results.TotalCommits)
	}

	return nil
}

// outputResults outputs validation results in the specified format.
func (c *ValidateCommand) outputResults(results *domain.ValidationResults) error {
	// Choose output format
	switch c.OutputFormat {
	case OutputFormatText:
		return c.outputTextResults(results)
	case OutputFormatJSON:
		return c.outputJSONResults(results)
	case OutputFormatGitHubActions:
		return c.outputGitHubResults(results)
	default:
		return fmt.Errorf("unsupported output format: %s", c.OutputFormat)
	}
}

// outputTextResults outputs results in human-readable text format.
func (c *ValidateCommand) outputTextResults(results *domain.ValidationResults) error {
	// Create a text formatter from the infrastructure layer
	formatter := output.NewTextFormatter(c.Verbose, true, false)

	// Format the results to text
	textOutput := formatter.Format(results)

	// Print the text output
	fmt.Print(textOutput)

	return nil
}

// outputJSONResults outputs results in JSON format.
func (c *ValidateCommand) outputJSONResults(results *domain.ValidationResults) error {
	// Create a JSON formatter from the infrastructure layer
	formatter := output.NewJSONFormatter()

	// Format the results to JSON
	jsonOutput := formatter.Format(results)

	// Print the JSON output
	fmt.Println(jsonOutput)

	return nil
}

// outputGitHubResults outputs results in GitHub Actions format.
func (c *ValidateCommand) outputGitHubResults(results *domain.ValidationResults) error {
	// Print summary
	fmt.Printf("::group::Summary\n")
	fmt.Printf("Validated %d commits\n", results.TotalCommits)
	fmt.Printf("Passed: %d, Failed: %d\n", results.PassedCommits, results.TotalCommits-results.PassedCommits)
	fmt.Printf("::endgroup::\n")

	// Print details for each commit
	for i, commitResult := range results.CommitResults {
		fmt.Printf("::group::Commit #%d: %s\n", i+1, commitResult.CommitInfo.Hash)
		fmt.Printf("Subject: %s\n", commitResult.CommitInfo.Subject)

		// Print rule results
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.Status == domain.StatusFailed {
				for _, err := range ruleResult.Errors {
					// Print error in GitHub format
					fmt.Printf("::error file=%s,line=1,title=%s::%s\n",
						commitResult.CommitInfo.Hash,
						ruleResult.RuleName,
						err.Error())
				}
			}

			// Print rule status
			status := "✅"
			if ruleResult.Status == domain.StatusFailed {
				status = "❌"
			}

			message := ruleResult.Message
			if c.Verbose {
				message = ruleResult.VerboseMessage
			}

			fmt.Printf("%s %s: %s\n", status, ruleResult.RuleName, message)
		}

		fmt.Printf("::endgroup::\n")
	}

	// Set output for GitHub Actions
	if !results.AllPassed() {
		fmt.Printf("::set-output name=passed::false\n")
	} else {
		fmt.Printf("::set-output name=passed::true\n")
	}

	return nil
}

// Run executes the validate command from command line arguments.
func Run(args []string) error {
	cmd := NewValidateCommand()

	// Parse command line arguments
	for argIndex := 0; argIndex < len(args); argIndex++ {
		switch args[argIndex] {
		case "-c", "--commit":
			if argIndex+1 < len(args) {
				cmd.CommitRef = args[argIndex+1]
				argIndex++
			}
		case "-n", "--count":
			if argIndex+1 < len(args) {
				count, err := strconv.Atoi(args[argIndex+1])
				if err != nil {
					return fmt.Errorf("invalid commit count: %s", args[argIndex+1])
				}

				cmd.CommitCount = count
				argIndex++
			}
		case "-f", "--from":
			if argIndex+1 < len(args) {
				cmd.FromRef = args[argIndex+1]
				argIndex++
			}
		case "-t", "--to":
			if argIndex+1 < len(args) {
				cmd.ToRef = args[argIndex+1]
				argIndex++
			}
		case "-m", "--message-file":
			if argIndex+1 < len(args) {
				cmd.MessageFile = args[argIndex+1]
				argIndex++
			}
		case "-o", "--output":
			if argIndex+1 < len(args) {
				format := strings.ToLower(args[argIndex+1])
				switch format {
				case "text", "json", "github":
					cmd.OutputFormat = OutputFormat(format)
				default:
					return fmt.Errorf("invalid output format: %s", format)
				}

				argIndex++
			}
		case "-v", "--verbose":
			cmd.Verbose = true
		case "--no-skip-merge":
			cmd.SkipMergeCommits = false
		case "-p", "--path":
			if argIndex+1 < len(args) {
				cmd.RepoPath = args[argIndex+1]
				argIndex++
			}
		case "-h", "--help":
			printHelp()

			return nil
		}
	}

	// Execute the command
	return cmd.Execute()
}

// printHelp prints help information for the validate command.
func printHelp() {
	fmt.Fprintln(os.Stderr, "Usage: gommitlint validate [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -c, --commit <ref>       Validate a specific commit")
	fmt.Fprintln(os.Stderr, "  -n, --count <n>          Validate the last n commits")
	fmt.Fprintln(os.Stderr, "  -f, --from <ref>         Start commit for range validation")
	fmt.Fprintln(os.Stderr, "  -t, --to <ref>           End commit for range validation")
	fmt.Fprintln(os.Stderr, "  -m, --message-file <file> Validate a commit message from a file")
	fmt.Fprintln(os.Stderr, "  -o, --output <format>    Output format (text, json, github)")
	fmt.Fprintln(os.Stderr, "  -v, --verbose            Verbose output")
	fmt.Fprintln(os.Stderr, "      --no-skip-merge      Include merge commits")
	fmt.Fprintln(os.Stderr, "  -p, --path <path>        Path to Git repository")
	fmt.Fprintln(os.Stderr, "  -h, --help               Show help")
}
