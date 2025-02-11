// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

//nolint:testpackage
package rules_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rules"
)

func TestValidateDCO(t *testing.T) {
	type testDesc struct {
		Name          string
		CommitMessage string
		ExpectValid   bool
	}

	for _, test := range []testDesc{
		{
			Name:          "Valid DCO",
			CommitMessage: "something nice\n\nSigned-off-by: Foo Bar <foobar@example.org>\n\n",
			ExpectValid:   true,
		},
		{
			Name:          "Valid DCO with CRLF",
			CommitMessage: "something nice\r\n\r\nSigned-off-by: Foo Bar <foobar@example.org>\r\n\r\n",
			ExpectValid:   true,
		},
		{
			Name:          "No DCO",
			CommitMessage: "something nice\n\nnot signed\n",
			ExpectValid:   false,
		},
	} {
		t.Run(test.Name, func(testing *testing.T) {
			var report model.Report

			commit := configuration.Gommit{Message: test.CommitMessage}
			report.AddCheck(rules.ValidateDCO(commit.Message))

			if test.ExpectValid {
				if !report.Valid() {
					testing.Error("Report is invalid with valid DCP")
				}
			} else {
				if report.Valid() {
					testing.Error("Report is valid with invalid DCO")
				}
			}
		})
	}
}
