// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
)

// Default values as simple constants.
const (
	DefaultSubjectDescriptionCase = "lower"
	DefaultSubjectInvalidSuffixes = ".! ?"
	DefaultSpellCheckLocale       = "UK"
)

// Default boolean values.
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

// checkValidity validates the commit message against all configured rules.
func (v *Validator) checkValidity(report *model.Report, commitInfo model.CommitInfo) {
	v.ensureDefaultValues()
	v.checkSubjectRules(report, commitInfo)
	v.checkSignatureRules(report, commitInfo)
	v.checkConventionalRules(report, commitInfo)
	v.checkAdditionalRules(report, commitInfo)
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

func (v *Validator) checkSubjectRules(report *model.Report, commitInfo model.CommitInfo) {
	if v.config.Subject == nil {
		return
	}

	subject := v.config.Subject
	isConventional := v.config.ConventionalCommit != nil

	report.AddRule(rule.ValidateSubjectLength(commitInfo.Subject, subject.MaxLength))

	if *subject.Imperative {
		report.AddRule(rule.ValidateImperative(commitInfo.Subject, isConventional))
	}

	report.AddRule(rule.ValidateSubjectCase(commitInfo.Subject, subject.Case, isConventional))
	report.AddRule(rule.ValidateSubjectSuffix(commitInfo.Subject, subject.InvalidSuffixes))

	if subject.Jira.IsRequired {
		report.AddRule(rule.ValidateJira(commitInfo.Subject, subject.Jira.Keys, isConventional))
	}
}

func (v *Validator) checkSignatureRules(report *model.Report, commitInfo model.CommitInfo) {
	if *v.config.IsSignOffRequired {
		report.AddRule(rule.ValidateSignOff(commitInfo.Body))
	}

	if v.config.Signature.IsRequired {
		report.AddRule(rule.ValidateSignature(commitInfo.Signature))

		if v.config.Signature.Identity != nil {
			report.AddRule(rule.ValidateGPGIdentity(commitInfo.Signature, commitInfo.RawCommit, v.config.Signature.Identity.PublicKeyURI))
		}
	}
}

func (v *Validator) checkConventionalRules(report *model.Report, commitInfo model.CommitInfo) {
	if v.config.ConventionalCommit.IsRequired {
		conv := v.config.ConventionalCommit
		report.AddRule(rule.ValidateConventionalCommit(commitInfo.Subject, conv.Types, conv.Scopes, conv.MaxDescriptionLength))
	}
}

func (v *Validator) checkAdditionalRules(report *model.Report, commitInfo model.CommitInfo) {
	report.AddRule(rule.ValidateSpelling(commitInfo.Message, v.config.SpellCheck.Locale))

	if *v.config.IsNCommitMax {
		report.AddRule(rule.ValidateNumberOfCommits(v.git, v.options.CommitRef))
	}

	if v.config.Body.IsRequired {
		report.AddRule(rule.ValidateCommitBody(commitInfo.Message))
	}
}
