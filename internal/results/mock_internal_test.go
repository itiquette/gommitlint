// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package results_test

import (
	"errors"
	"fmt"
	"io"

	"github.com/itiquette/gommitlint/internal"
	"github.com/itiquette/gommitlint/internal/model"
)

// Ensure testWriter is initialized before tests run.
var testWriter io.Writer

// mockPrintReport is used to replace internal.PrintReport for testing purposes.
func mockPrintReport(rules []model.CommitRule, commitInfo *model.CommitInfo, opts *internal.PrintOptions) error {
	// Get the writer from the test context
	writer := testWriter
	if writer == nil {
		return errors.New("test writer not set")
	}

	// Generate predictable output for test assertions
	if len(rules) == 0 {
		fmt.Fprintln(writer, "No rules to validate")

		return nil
	}

	// Print commit info if available
	if commitInfo != nil && commitInfo.RawCommit != nil {
		fmt.Fprintf(writer, "COMMIT-SHA: %s\n", commitInfo.RawCommit.Hash.String()[:7])
		fmt.Fprintf(writer, "SUBJECT: %s\n\n", commitInfo.Subject)
	}

	// Print each rule with its results
	for _, rule := range rules {
		fmt.Fprintf(writer, "%s: %s\n", rule.Name(), rule.Result())

		if opts.Verbose {
			fmt.Fprintf(writer, "    %s\n", rule.VerboseResult())
		}

		if len(rule.Errors()) > 0 {
			fmt.Fprintln(writer, "    FAIL: Rule validation failed")

			if opts.ShowHelp {
				fmt.Fprintf(writer, "    %s\n", rule.Help())
			} else if opts.Verbose {
				fmt.Fprintln(writer, "    TIP: Run with '--rulehelp=' for detailed instructions")
			}
		}
	}

	// Print overall result
	allPassed := true

	for _, rule := range rules {
		if len(rule.Errors()) > 0 {
			allPassed = false

			break
		}
	}

	if allPassed {
		fmt.Fprintf(writer, "\nSUCCESS: All rules passed (%d/%d)\n", len(rules), len(rules))
	} else {
		passedCount := 0

		for _, rule := range rules {
			if len(rule.Errors()) == 0 {
				passedCount++
			}
		}

		fmt.Fprintf(writer, "\nFAIL: %d of %d rules passed\n", passedCount, len(rules))
	}

	return nil
}

// Replace the internal.PrintReport with our mock for testing.
func setupTestPrintReport() func() {
	// Store the original function
	originalFunc := internal.PrintReport

	// Replace with our mock
	internal.PrintReport = mockPrintReport

	// Return a cleanup function
	return func() {
		internal.PrintReport = originalFunc
	}
}
