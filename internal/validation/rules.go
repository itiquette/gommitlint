// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/itiquette/gommitlint/internal/rule/signedidentityrule"
)

// No longer needed since we're using defaults package

func (v *Validator) checkValidity(ctx context.Context, commitRules *model.CommitRules, commitInfo model.CommitInfo) {
	// Check for cancellation
	if ctx.Err() != nil {
		return
	}

	// Skip merge commits if configured to do so
	if *v.config.IgnoreMergeCommits && commitInfo.IsMergeCommit {
		return
	}

	v.checkSubjectRules(ctx, commitRules, commitInfo)
	v.checkSignatureRules(ctx, commitRules, commitInfo)
	v.checkConventionalRules(ctx, commitRules, commitInfo)
	v.checkAdditionalRules(ctx, commitRules, commitInfo)
}

func (v *Validator) checkSubjectRules(ctx context.Context, report *model.CommitRules, commitInfo model.CommitInfo) {
	// Check for cancellation
	if ctx.Err() != nil {
		return
	}

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

func (v *Validator) checkSignatureRules(ctx context.Context, report *model.CommitRules, commitInfo model.CommitInfo) {
	// Check for cancellation
	if ctx.Err() != nil {
		return
	}

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

func (v *Validator) checkConventionalRules(ctx context.Context, report *model.CommitRules, commitInfo model.CommitInfo) {
	// Check for cancellation
	if ctx.Err() != nil {
		return
	}

	if v.config.ConventionalCommit.Required {
		conv := v.config.ConventionalCommit
		ccRule := rule.ValidateConventionalCommit(commitInfo.Subject, conv.Types, conv.Scopes, conv.MaxDescriptionLength)
		report.Add(ccRule)
	}
}

func (v *Validator) checkAdditionalRules(ctx context.Context, report *model.CommitRules, commitInfo model.CommitInfo) {
	// Check for cancellation
	if ctx.Err() != nil {
		return
	}

	// Only run spell checking if it's enabled
	if v.config.SpellCheck.Enabled {
		spellRule := rule.ValidateSpelling(commitInfo.Message, v.config.SpellCheck.Locale)
		report.Add(spellRule)
	}

	if *v.config.NCommitsAhead {
		commitsAhead := rule.ValidateNumberOfCommits(v.repo, v.options.CommitRef)
		report.Add(commitsAhead)
	}

	if v.config.Body.Required {
		commitBodyRule := rule.ValidateCommitBody(commitInfo.Message)
		report.Add(commitBodyRule)
	}
}
