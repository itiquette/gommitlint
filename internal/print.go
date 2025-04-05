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

	"github.com/itiquette/gommitlint/internal/model"
)

type validationStatus string

const (
	statusPass   validationStatus = "PASS"
	statusFailed validationStatus = "FAILED"
)

// Constants for report formatting.
const (
	reportHeader   = "RULE\tSTATUS\tRESULT\t"
	tabPadding     = 8
	tabWriterFlags = 0
)

// PrintReport prints the results of all rule validations.
func PrintReport(rules []model.CommitRule) error {
	return PrintReportTo(os.Stdout, rules)
}

// PrintReportTo prints the report to the specified writer.
func PrintReportTo(writer io.Writer, rules []model.CommitRule) error {
	tabWriter := tabwriter.NewWriter(writer, 0, 0, tabPadding, ' ', tabWriterFlags)
	defer tabWriter.Flush()

	if err := printSubject(tabWriter); err != nil {
		return fmt.Errorf("failed to print subject: %w", err)
	}

	if err := printRules(tabWriter, rules); err != nil {
		return err
	}

	return nil
}

func printSubject(writer io.Writer) error {
	_, err := fmt.Fprintln(writer, reportHeader)

	return err
}

func printRules(writer io.Writer, rules []model.CommitRule) error {
	var failed bool

	for _, rule := range rules {
		if errs := rule.Errors(); len(errs) > 0 {
			if err := printFailedRule(writer, rule, errs); err != nil {
				return err
			}

			failed = true
		} else {
			if err := printPassedRule(writer, rule); err != nil {
				return err
			}
		}
	}

	if failed {
		return errors.New("one or more rules failed")
	}

	return nil
}

func printFailedRule(writer io.Writer, rule model.CommitRule, errs []error) error {
	for _, err := range errs {
		if _, err := fmt.Fprintf(writer, "%s\t%s\t%v\t\n", rule.Name(), statusFailed, err); err != nil {
			return fmt.Errorf("failed to print failed rule: %w", err)
		}
	}

	return nil
}

func printPassedRule(writer io.Writer, rule model.CommitRule) error {
	_, err := fmt.Fprintf(writer, "%s\t%s\t%s\t\n", rule.Name(), statusPass, rule.Result())
	if err != nil {
		return fmt.Errorf("failed to print passed rule: %w", err)
	}

	return nil
}
