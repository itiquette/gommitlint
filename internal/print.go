// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package internal

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/itiquette/gommitlint/internal/model"
	"github.com/pkg/errors"
)

// reportStatus represents the status of a rule check.
type reportStatus string

const (
	statusPass   reportStatus = "PASS"
	statusFailed reportStatus = "FAILED"
)

// Constants for report formatting.
const (
	reportHeader   = "RULE\tSTATUS\tRESULT\t"
	tabPadding     = 8
	tabWriterFlags = 0
)

// PrintReport prints the results of all rule checks and returns an error if any checks failed.
// It formats the output in a tabulated form using a tabwriter.
func PrintReport(checks []model.CommitRule) error {
	return PrintReportTo(os.Stdout, checks)
}

// PrintReportTo prints the report to the specified writer, allowing for better testing
// and flexibility in output destination.
func PrintReportTo(writer io.Writer, checks []model.CommitRule) error {
	tabWriter := tabwriter.NewWriter(writer, 0, 0, tabPadding, ' ', tabWriterFlags)
	defer tabWriter.Flush()

	if err := printSubject(tabWriter); err != nil {
		return errors.Wrap(err, "failed to print subject")
	}

	if err := printChecks(tabWriter, checks); err != nil {
		return err
	}

	return nil
}

func printSubject(writer io.Writer) error {
	_, err := fmt.Fprintln(writer, reportHeader)

	return err
}

func printChecks(writer io.Writer, checks []model.CommitRule) error {
	var failed bool

	for _, check := range checks {
		if errs := check.Errors(); len(errs) > 0 {
			if err := printFailedCheck(writer, check, errs); err != nil {
				return err
			}

			failed = true
		} else {
			if err := printPassedCheck(writer, check); err != nil {
				return err
			}
		}
	}

	if failed {
		return errors.New("one or more rules failed")
	}

	return nil
}

func printFailedCheck(writer io.Writer, check model.CommitRule, errs []error) error {
	for _, err := range errs {
		if _, err := fmt.Fprintf(writer, "%s\t%s\t%v\t\n", check.Name(), statusFailed, err); err != nil {
			return errors.Wrap(err, "failed to print failed check")
		}
	}

	return nil
}

func printPassedCheck(writer io.Writer, check model.CommitRule) error {
	_, err := fmt.Fprintf(writer, "%s\t%s\t%s\t\n", check.Name(), statusPass, check.Result())

	return errors.Wrap(err, "failed to print passed check")
}
