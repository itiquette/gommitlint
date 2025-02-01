// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/itiquette/gommitlint/internal/interfaces"

// Report reports the compliance checks.
type Report struct {
	checks []interfaces.Check
}

// Valid checks if a report is valid.
func (r *Report) Valid() bool {
	for _, check := range r.checks {
		if len(check.Errors()) != 0 {
			return false
		}
	}

	return true
}

// Checks returns the checks executed by a validator run.
func (r *Report) Checks() []interfaces.Check {
	return r.checks
}

// AddCheck adds a check to the report.
func (r *Report) AddCheck(c interfaces.Check) {
	r.checks = append(r.checks, c)
}
