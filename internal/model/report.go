// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package model

import "github.com/itiquette/gommitlint/internal/interfaces"

// Report reports the compliance checks.
type Report struct {
	rules []interfaces.Rule
}

// Valid checks if a report is valid.
func (r *Report) Valid() bool {
	for _, rule := range r.rules {
		if len(rule.Errors()) != 0 {
			return false
		}
	}

	return true
}

// Rules returns the rules executed by a validator run.
func (r *Report) Rules() []interfaces.Rule {
	return r.rules
}

// AddRule adds a rule to the report.
func (r *Report) AddRule(c interfaces.Rule) {
	r.rules = append(r.rules, c)
}
