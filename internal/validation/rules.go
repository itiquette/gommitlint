// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/itiquette/gommitlint/internal/rule/signedidentityrule"
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

	if *v.config.IgnoreMergeCommits && commitInfo.IsMergeCommit {
		//fmt.Printf("Ignoring merge commit")
		return
	}

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
			v.config.Subject.Jira = &configuration.JiraRule{Required: false}
		}
	}

	// Signature defaults
	if v.config.SignOffRequired == nil {
		v.config.SignOffRequired = boolPtr(DefaultSignOffRequired)
	}

	if v.config.Signature == nil {
		v.config.Signature = &configuration.SignatureRule{Required: true}
	}

	// Conventional commit defaults
	if v.config.ConventionalCommit == nil {
		v.config.ConventionalCommit = &configuration.ConventionalRule{
			Required: true,
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
		v.config.Body = &configuration.BodyRule{Required: false}
	}

	if v.config.NCommitsAhead == nil {
		v.config.NCommitsAhead = boolPtr(DefaultOneCommitMax)
	}

	if v.config.IgnoreMergeCommits == nil {
		v.config.IgnoreMergeCommits = boolPtr(true)
	}
}

func (v *Validator) checkSubjectRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	if v.config.Subject == nil {
		return
	}

	subject := v.config.Subject
	isConventional := v.config.ConventionalCommit.Required

	subjectLengthRule := rule.ValidateSubjectLength(commitInfo.Subject, subject.MaxLength)
	report.Add(subjectLengthRule)

	if *subject.Imperative {
		imperativeRule := rule.ValidateImperative(commitInfo.Subject, isConventional)
		report.Add(&imperativeRule)
	}

	subjectCaseRule := rule.ValidateSubjectCase(commitInfo.Subject, subject.Case, isConventional)
	report.Add(subjectCaseRule)

	subjectSuffixRule := rule.ValidateSubjectSuffix(commitInfo.Subject, subject.InvalidSuffixes)
	report.Add(subjectSuffixRule)

	if subject.Jira.Required {
		jiraReferenceRule := rule.ValidateJiraReference(commitInfo.Subject, commitInfo.Body, subject.Jira, isConventional)
		report.Add(jiraReferenceRule)
	}
}

func (v *Validator) checkSignatureRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	if *v.config.SignOffRequired {
		signOffRule := rule.ValidateSignOff(commitInfo.Body)
		report.Add(signOffRule)
	}

	if v.config.Signature.Required {
		signatureRule := rule.ValidateSignature(commitInfo.Signature)
		report.Add(signatureRule)

		if v.config.Signature.Identity != nil {
			signedIdentityRule := signedidentityrule.VerifySignatureIdentity(commitInfo.RawCommit, commitInfo.Signature, v.config.Signature.Identity.PublicKeyURI)
			report.Add(signedIdentityRule)
		}
	}
}

func (v *Validator) checkConventionalRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	if v.config.ConventionalCommit.Required {
		conv := v.config.ConventionalCommit
		ccRule := rule.ValidateConventionalCommit(commitInfo.Subject, conv.Types, conv.Scopes, conv.MaxDescriptionLength)
		report.Add(ccRule)
	}
}

func (v *Validator) checkAdditionalRules(report *model.CommitRules, commitInfo model.CommitInfo) {
	spellRule := rule.ValidateSpelling(commitInfo.Message, v.config.SpellCheck.Locale)
	report.Add(spellRule)

	if *v.config.NCommitsAhead {
		commitsAhead := rule.ValidateNumberOfCommits(v.repo, v.options.CommitRef)
		report.Add(commitsAhead)
	}

	if v.config.Body.Required {
		commitBodyRule := rule.ValidateCommitBody(commitInfo.Message)
		report.Add(commitBodyRule)
	}
}
