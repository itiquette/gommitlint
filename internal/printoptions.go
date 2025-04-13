// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package internal

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/model"
)

// PrintOptions configures the PrintReport function's behavior.
type PrintOptions struct {
	Verbose        bool
	ShowHelp       bool
	RuleToShowHelp string
	LightMode      bool // Whether to use light mode colors
}

// ColorScheme defines colors for different UI elements.
type ColorScheme struct {
	Success     func(format string, a ...interface{}) string
	Error       func(format string, a ...interface{}) string
	Warning     func(format string, a ...interface{}) string
	Info        func(format string, a ...interface{}) string
	VerboseInfo func(format string, a ...interface{}) string
	ErrorDetail func(format string, a ...interface{}) string
	HelpText    func(format string, a ...interface{}) string
	Bold        func(format string, a ...interface{}) string
	Italic      func(format string, a ...interface{}) string
	Header      func(format string, a ...interface{}) string // Added for header styling
}

// PrintReport prints validation results.
func PrintReport(rules []model.CommitRule, commitInfo *model.CommitInfo, opts *PrintOptions) error {
	// Default options if none provided
	if opts == nil {
		opts = &PrintOptions{
			Verbose:        false,
			ShowHelp:       false,
			RuleToShowHelp: "",
			LightMode:      false,
		}
	}

	// Check for NO_COLOR environment variable
	noColor := false
	if val, exists := os.LookupEnv("NO_COLOR"); exists && val != "" {
		noColor = true
		color.NoColor = true
	}

	// Set up color scheme based on mode preference and accessibility
	colorScheme := getColorScheme(opts.LightMode, noColor)

	// Check if we should show help for a specific rule
	if opts.ShowHelp && opts.RuleToShowHelp != "" {
		found := false

		for _, rule := range rules {
			if strings.EqualFold(rule.Name(), opts.RuleToShowHelp) {
				ruleName := colorScheme.Bold(rule.Name())

				fmt.Printf("\n%s Rule Help:\n", ruleName)

				// Properly indent and style all help text lines
				helpLines := strings.Split(rule.Help(), "\n")
				for _, line := range helpLines {
					fmt.Printf("  %s\n", colorScheme.HelpText(line))
				}

				fmt.Println("")

				found = true

				break
			}
		}

		if !found {
			fmt.Printf("No help found for rule: %s\n", opts.RuleToShowHelp)
			fmt.Println("Available rules:")

			for _, rule := range rules {
				fmt.Printf("  - %s\n", colorScheme.Bold(rule.Name()))
			}
		}

		return nil
	}

	// Print header with commit information
	if commitInfo != nil && commitInfo.RawCommit != nil {
		printCommitHeader(commitInfo, colorScheme)
	}

	// Sort rules alphabetically by name
	sortedRules := make([]model.CommitRule, len(rules))
	copy(sortedRules, rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return strings.ToLower(sortedRules[i].Name()) < strings.ToLower(sortedRules[j].Name())
	})

	// Print validation results
	passedRules := 0
	totalRules := len(sortedRules)

	// Use Unicode symbols based on terminal capabilities
	passSymbol := colorScheme.Success("PASS")
	failSymbol := colorScheme.Error("FAIL")

	if canHandleUnicode() {
		passSymbol = colorScheme.Success("✓")
		failSymbol = colorScheme.Error("✗")
	}

	for _, rule := range sortedRules {
		ruleName := colorScheme.Bold(rule.Name())

		if len(rule.Errors()) == 0 {
			// Success
			passedRules++

			fmt.Printf("%s %s: ", passSymbol, ruleName)

			// Print basic result message
			fmt.Printf("%s\n", rule.Result())

			// In verbose mode, add detailed information
			if opts.Verbose {
				fmt.Printf("    %s\n", colorScheme.VerboseInfo(rule.VerboseResult()))
			}
		} else {
			// Error
			fmt.Printf("%s %s: ", failSymbol, ruleName)

			// Print basic error message in red
			fmt.Printf("%s\n", colorScheme.Error(rule.Result()))

			// In verbose mode or show help mode, add detailed error information
			if opts.Verbose || opts.ShowHelp {
				// Show detailed error message (verbose result)
				fmt.Printf("  %s\n", colorScheme.ErrorDetail(rule.VerboseResult()))
			}

			// If help mode, show full help text
			if opts.ShowHelp {
				fmt.Println("")

				// Properly indent and style all help text lines
				helpLines := strings.Split(rule.Help(), "\n")
				for _, line := range helpLines {
					fmt.Printf("  %s\n", colorScheme.HelpText(line))
				}

				fmt.Println("")
			} else if opts.Verbose {
				// In verbose mode, show a compact tip
				fmt.Printf("    %s Run with '--rulehelp=%s' for detailed instructions\n\n",
					colorScheme.Warning("TIP:"), strings.ToLower(rule.Name()))
			}
		}
	}

	// Print summary line
	if passedRules == totalRules {
		fmt.Printf("\n%s All rules passed (%d/%d)\n",
			colorScheme.Success("SUCCESS:"), passedRules, totalRules)
	} else {
		fmt.Printf("\n%s %d of %d rules passed\n",
			colorScheme.Warning("FAIL:"), passedRules, totalRules)

		// Add a help hint if we're not in verbose mode
		if !opts.Verbose {
			fmt.Printf("%s Run with '--verbose' and '--extra-verbose' for more details or '--rulehelp=<rule>' for specific rule help\n",
				colorScheme.Warning("TIP:"))
		}
	}

	fmt.Println()

	return nil
}

// printCommitHeader prints a header with commit SHA and message information.
func printCommitHeader(commitInfo *model.CommitInfo, colourScheme ColorScheme) {
	if commitInfo == nil || commitInfo.RawCommit == nil {
		return
	}

	// Get the short SHA (first 7 characters)
	shortSHA := commitInfo.RawCommit.Hash.String()[:7]

	// Print a section divider
	divider := strings.Repeat("=", 80)
	fmt.Println(colourScheme.Header(divider))

	// Print commit info header
	fmt.Printf("%s %s\n", colourScheme.Header("COMMIT-SHA:"), colourScheme.Bold(shortSHA))

	// Print subject line
	fmt.Printf("%s %s\n", colourScheme.Header("SUBJECT:"), commitInfo.Subject)

	// Print full message if body exists
	if commitInfo.Body != "" {
		fmt.Printf("%s\n%s\n", colourScheme.Header("MESSAGE:"), commitInfo.Body)
	}

	fmt.Println(colourScheme.Header(divider))
	fmt.Println() // Add a blank line before rule results
}

// getColorScheme returns appropriate color functions based on mode and accessibility.
func getColorScheme(lightMode, noColor bool) ColorScheme {
	if noColor {
		// Return a scheme with no colors
		return ColorScheme{
			Success:     fmt.Sprintf,
			Error:       fmt.Sprintf,
			Warning:     fmt.Sprintf,
			Info:        fmt.Sprintf,
			VerboseInfo: fmt.Sprintf,
			ErrorDetail: fmt.Sprintf,
			HelpText:    fmt.Sprintf,
			Bold:        fmt.Sprintf,
			Italic:      fmt.Sprintf,
			Header:      fmt.Sprintf,
		}
	}

	// Keep standard green/red for pass/fail indicators
	successColor := color.New(color.FgGreen, color.Bold).SprintfFunc()
	errorColor := color.New(color.FgRed, color.Bold).SprintfFunc()
	regularErrorColor := color.New(color.FgRed).SprintfFunc()

	if lightMode {
		// Light mode colors (higher contrast on light backgrounds)
		return ColorScheme{
			Success:     successColor,                                        // Keep green for success
			Error:       errorColor,                                          // Keep red for errors
			Warning:     color.New(color.FgBlack, color.Bold).SprintfFunc(),  // Black for warnings
			Info:        color.New(color.FgBlue).SprintfFunc(),               // Blue for info
			VerboseInfo: color.New(color.FgBlue).SprintfFunc(),               // Blue for verbose success info
			ErrorDetail: regularErrorColor,                                   // Regular red for error details
			HelpText:    color.New(color.FgBlue, color.Italic).SprintfFunc(), // Blue italic for help text
			Bold:        color.New(color.Bold).SprintfFunc(),
			Italic:      color.New(color.Italic).SprintfFunc(),
			Header:      color.New(color.FgBlue, color.Bold).SprintfFunc(), // Blue bold for header
		}
	}

	// Dark mode colors (default - better for most terminal backgrounds)
	return ColorScheme{
		Success:     successColor,                                        // Keep green for success
		Error:       errorColor,                                          // Keep red for errors
		Warning:     color.New(color.FgYellow, color.Bold).SprintfFunc(), // Yellow for warnings
		Info:        color.New(color.FgCyan).SprintfFunc(),               // Cyan for info
		VerboseInfo: color.New(color.FgHiCyan).SprintfFunc(),             // Bright cyan for verbose success
		ErrorDetail: regularErrorColor,                                   // Regular red for error details
		HelpText:    color.New(color.FgCyan, color.Italic).SprintfFunc(), // Cyan italic for help text
		Bold:        color.New(color.Bold).SprintfFunc(),
		Italic:      color.New(color.Italic).SprintfFunc(),
		Header:      color.New(color.FgMagenta, color.Bold).SprintfFunc(), // Magenta bold for header
	}
}

// canHandleUnicode checks if the terminal supports Unicode characters.
func canHandleUnicode() bool {
	// Check common environment variables for UTF-8 support
	for _, envVar := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		value := strings.ToLower(os.Getenv(envVar))
		if strings.Contains(value, "utf-8") || strings.Contains(value, "utf8") {
			return true
		}
	}

	return false
}
