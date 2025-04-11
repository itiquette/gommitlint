// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/model"
)

// validationStatus represents the status of a validation rule.
type validationStatus string

const (
	statusPass   validationStatus = "✓ PASS"
	statusFailed validationStatus = "✗ FAILED"
)

// CommitInfo contains information about the current commit.
type CommitInfo struct {
	Subject       string
	SHA           string
	CompareBranch string
	Author        string
	Date          time.Time
}

// PrintReport prints the validation report to stdout.
func PrintReport(rules []model.CommitRule, commitInfo *CommitInfo) error {
	return PrintReportTo(os.Stdout, rules, commitInfo)
}

// PrintReportTo prints the validation report to the specified writer.
func PrintReportTo(writer io.Writer, rules []model.CommitRule, _ *CommitInfo) error {
	// Print commit info box if available
	// if commitInfo != nil {
	// 	if err := printCommitInfoBox(writer, commitInfo); err != nil {
	// 		return fmt.Errorf("failed to print commit info: %w", err)
	// 	}
	// 	fmt.Fprintln(writer) // Add spacing
	// }
	// Create a new tabwriter for aligned columns
	tabwriter := tabwriter.NewWriter(writer, 0, 0, 4, ' ', 0)
	defer tabwriter.Flush()

	// Print all rules
	var failed bool

	for _, rule := range rules {
		if errs := rule.Errors(); len(errs) > 0 {
			for _, err := range errs {
				printRule(tabwriter, rule.Name(), statusFailed, err.Error(), color.FgRed)
			}

			failed = true
		} else {
			printRule(tabwriter, rule.Name(), statusPass, rule.Result(), color.FgGreen)
		}
	}

	// Count results
	passed, total := countResults(rules)
	passRate := float64(passed) / float64(total) * 100

	// Determine color based on pass rate
	summaryColor := color.FgRed
	if passRate == 100 {
		summaryColor = color.FgGreen
	} else if passRate >= 80 {
		summaryColor = color.FgYellow
	}

	// Print summary
	summary := fmt.Sprintf("%d/%d", passed, total)
	result := fmt.Sprintf("%.1f%% passed", passRate)
	printRule(tabwriter, "SUMMARY", "", summary+" "+result, summaryColor)

	if failed {
		return errors.New("one or more rules failed")
	}

	return nil
}

// printRule outputs a single rule line with proper formatting.
func printRule(writer io.Writer, name string, status validationStatus, result string, statusColor color.Attribute) {
	// Apply color to status text
	statusText := status

	if status != "" {
		colorFunc := color.New(statusColor, color.Bold).SprintFunc()
		statusText = validationStatus(colorFunc(string(status)))
	}

	fmt.Fprintf(writer, "%s\t%s\t%s\n", name, statusText, result)
}

// countResults counts the passed and total rules.
func countResults(rules []model.CommitRule) (int, int) {
	total := len(rules)
	passed := 0

	for _, rule := range rules {
		if len(rule.Errors()) == 0 {
			passed++
		}
	}

	return passed, total
}

// // printCommitInfoBox renders a box with commit information.
// func printCommitInfoBox(writer io.Writer, info *CommitInfo) error {
// 	// Create new table writer
// 	table := tablewriter.NewWriter(writer)

// 	// Set table style
// 	table.SetBorder(true)
// 	table.SetRowLine(false)
// 	table.SetColumnSeparator("│")
// 	table.SetRowSeparator("─")
// 	table.SetCenterSeparator("┼")

// 	// Set header colors
// 	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
// 	headerLabels := []string{
// 		cyan("COMMIT SHA"),
// 		cyan("SUBJECT"),
// 		cyan("COMPARE BRANCH"),
// 		cyan("AUTHOR"),
// 		cyan("DATE"),
// 	}
// 	table.SetHeader(headerLabels)

// 	// Format date
// 	dateStr := ""
// 	if !info.Date.IsZero() {
// 		dateStr = info.Date.Format("2006-01-02 15:04:05")
// 	}

// 	// Add a single row with all info
// 	table.Append([]string{
// 		info.SHA,
// 		info.Subject,
// 		info.CompareBranch,
// 		info.Author,
// 		dateStr,
// 	})

// 	// Render table
// 	table.Render()

// 	return nil
// }
