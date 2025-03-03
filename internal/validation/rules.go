// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
)

const (
	DefaultSubjectDescriptionCase = "lower"
	DefaultSubjectInvalidSuffixes = ".! ?"
	DefaultSpellCheckLocale       = "UK"
)

var (
	DefaultSubjectImperativeRequired = true
	DefaultSignOffRequired           = true
	DefaultOneCommitMax              = true
)

var DefaultConventionalTypes = []string{
	"build", "chore", "ci", "docs", "feat", "fix",
	"perf", "refactor", "revert", "style", "test",
}

// boolPtr returns a pointer to the given boolean value.
func boolPtr(b bool) *bool {
	return &b
}

func (v *Validator) checkValidity(commitRules *model.CommitRules, commitInfo model.CommitInfo) {
	v.ensureDefaultValues()
	v.checkSubjectRules(commitRules, commitInfo)
	v.checkSignatureRules(commitRules, commitInfo)
	v.checkConventionalRules(commitRules, commitInfo)
	v.checkAdditionalRules(commitRules, commitInfo)
}

// ensureDefaultValues ensures all configuration values have appropriate defaults.
func (v *Validator) ensureDefaultValues() {
	// Subject defaults
	if v.config.Subject != nil {
		if v.config.Subject.Imperative == nil {
			v.config.Subject.Imperative = boolPtr(DefaultSubjectImperativeRequired)
		}

		if v.config.Subject.Case == "" {
			v.config.Subject.Case = DefaultSubjectDescriptionCase
		}

		if v.config.Subject.InvalidSuffixes == "" {
			v.config.Subject.InvalidSuffixes = DefaultSubjectInvalidSuffixes
		}

		if v.config.Subject.Jira == nil {
			v.config.Subject.Jira = &configuration.JiraRule{IsRequired: false}
		}
	}

	// Signature defaults
	if v.config.IsSignOffRequired == nil {
		v.config.IsSignOffRequired = boolPtr(DefaultSignOffRequired)
	}

	if v.config.Signature == nil {
		v.config.Signature = &configuration.SignatureRule{IsRequired: true}
	}

	// Conventional commit defaults
	if v.config.ConventionalCommit == nil {
		v.config.ConventionalCommit = &configuration.ConventionalRule{
			IsRequired: true,
		}
	}

	if len(v.config.ConventionalCommit.Types) == 0 {
		v.config.ConventionalCommit.Types = DefaultConventionalTypes
	}

	// Additional rules defaults
	if v.config.SpellCheck == nil {
		v.config.SpellCheck = &configuration.SpellingRule{Locale: DefaultSpellCheckLocale}
	}

	if v.config.Body == nil {
		v.config.Body = &configuration.BodyRule{IsRequired: false}
	}

	if v.config.IsNCommitMax == nil {
		v.config.IsNCommitMax = boolPtr(DefaultOneCommitMax)
	}
}

func (v *Validator) checkSubjectRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	if v.config.Subject == nil {
		return
	}

	subject := v.config.Subject
	isConventional := v.config.ConventionalCommit != nil

	report.Add(rule.ValidateSubjectLengthRule(commitInfo.Subject, subject.MaxLength))

	if *subject.Imperative {
		report.Add(rule.ValidateImperativeRule(commitInfo.Subject, isConventional))
	}

	report.Add(rule.ValidateSubjectCaseRule(commitInfo.Subject, subject.Case, isConventional))
	report.Add(rule.ValidateSubjectSuffix(commitInfo.Subject, subject.InvalidSuffixes))

	if subject.Jira.IsRequired {
		report.Add(rule.ValidateJira(commitInfo.Subject, subject.Jira.Keys, isConventional))
	}
}

func (v *Validator) checkSignatureRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	if *v.config.IsSignOffRequired {
		report.Add(rule.ValidateSignOffRule(commitInfo.Body))
	}

	if v.config.Signature.IsRequired {
		report.Add(rule.ValidateSignatureRule(commitInfo.Signature))

		if v.config.Signature.Identity != nil {
			report.Add(rule.ValidateGPGIdentity(commitInfo.Signature, commitInfo.RawCommit, v.config.Signature.Identity.PublicKeyURI))
		}
	}
}

func (v *Validator) checkConventionalRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	if v.config.ConventionalCommit.IsRequired {
		conv := v.config.ConventionalCommit
		report.Add(rule.ValidateConventionalCommit(commitInfo.Subject, conv.Types, conv.Scopes, conv.MaxDescriptionLength))
	}
}

func (v *Validator) checkAdditionalRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	report.Add(rule.ValidateSpellingRule(commitInfo.Message, v.config.SpellCheck.Locale))

	if *v.config.IsNCommitMax {
		report.Add(rule.ValidateNumberOfCommits(v.repo, v.options.CommitRef))
	}

	if v.config.Body.IsRequired {
		report.Add(rule.ValidateCommitBody(commitInfo.Message))
	}
}
