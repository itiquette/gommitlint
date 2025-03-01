// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "github.com/itiquette/gommitlint/internal/interfaces"

type Report struct {
	rules []interfaces.CommitRule
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

func (r *Report) Rules() []interfaces.CommitRule {
	return r.rules
}

func (r *Report) AddRule(c interfaces.CommitRule) {
	r.rules = append(r.rules, c)
}
