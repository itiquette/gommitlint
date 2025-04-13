// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package model

type CommitRule interface {
	// Name returns the name of the rule.
	Name() string

	// Result returns a concise human-readable result message.
	Result() string

	// VerboseResult returns a more detailed explanation for verbose mode.
	// If not implemented, it falls back to Result().
	VerboseResult() string

	// Errors returns any violations of the rule.
	Errors() []*ValidationError

	// Help returns a comprehensive guide on how to fix the rule violation.
	Help() string
}

type CommitRules struct {
	rules []CommitRule
}

func NewCommitRules() *CommitRules {
	return &CommitRules{
		rules: make([]CommitRule, 0, 50),
	}
}

func (r *CommitRules) All() []CommitRule {
	return r.rules
}

func (r *CommitRules) Add(c CommitRule) {
	r.rules = append(r.rules, c)
}
