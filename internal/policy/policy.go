// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

// Package policy provides base policy definitions.
package policy

// Report reports the compliance of a policy.
type Report struct {
	checks []Check
}

// Check defines a policy check.
type Check interface {
	Name() string
	Message() string
	Errors() []error
}

// Policy is an interface that policies must implement.
type Policy interface {
	Compliance(options *Options) (*Report, error)
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

// Checks returns the checks executed by a policy.
func (r *Report) Checks() []Check {
	return r.checks
}

// AddCheck adds a check to the policy report.
func (r *Report) AddCheck(c Check) {
	r.checks = append(r.checks, c)
}
