// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package factories

import (
	"context"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
)

// SimpleRuleFactory creates rules with minimal configuration.
type SimpleRuleFactory struct {
}

// NewSimpleRuleFactory creates a new simple rule factory.
func NewSimpleRuleFactory() *SimpleRuleFactory {
	return &SimpleRuleFactory{}
}

// CreateBasicRules creates all standard validation rules.
func (f *SimpleRuleFactory) CreateBasicRules() map[string]func(context.Context) domain.Rule {
	return map[string]func(context.Context) domain.Rule{
		"SubjectLength": func(_ context.Context) domain.Rule {
			return rules.NewSubjectLengthRule()
		},
		"SubjectCase": func(_ context.Context) domain.Rule {
			return rules.NewSubjectCaseRule()
		},
		"SubjectSuffix": func(_ context.Context) domain.Rule {
			return rules.NewSubjectSuffixRule()
		},
		"ConventionalCommit": func(_ context.Context) domain.Rule {
			return rules.NewConventionalCommitRule()
		},
		"ImperativeVerb": func(_ context.Context) domain.Rule {
			return rules.NewImperativeVerbRule()
		},
		"CommitBody": func(_ context.Context) domain.Rule {
			return rules.NewCommitBodyRule()
		},
		"JiraReference": func(_ context.Context) domain.Rule {
			return rules.NewJiraReferenceRule()
		},
		"SignOff": func(_ context.Context) domain.Rule {
			return rules.NewSignOffRule()
		},
		"Signature": func(_ context.Context) domain.Rule {
			return rules.NewSignatureRule()
		},
		"Identity": func(_ context.Context) domain.Rule {
			return rules.NewIdentityRule()
		},
		"Spell": func(_ context.Context) domain.Rule {
			return rules.NewSpellRule()
		},
	}
}

// CreateAnalyzerRules creates rules that require a commit analyzer.
func (f *SimpleRuleFactory) CreateAnalyzerRules(analyzer domain.CommitAnalyzer) map[string]func(context.Context) domain.Rule {
	return map[string]func(context.Context) domain.Rule{
		"BranchAhead": func(_ context.Context) domain.Rule {
			return rules.NewBranchAheadRule(rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return analyzer
			}))
		},
	}
}
