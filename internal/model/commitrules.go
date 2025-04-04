// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package model

type CommitRule interface {
	Name() string
	Result() string
	Errors() []error
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
