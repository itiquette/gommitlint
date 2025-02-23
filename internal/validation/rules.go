// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rule"
)

// checkValidity validates the commit message against all configured rules.
func (v *Validator) checkValidity(report *model.Report) {
	v.checkHeaderRules(report)
	v.checkSignatureRules(report)
	v.checkConventionalRules(report)
	v.checkAdditionalRules(report)
}

func (v *Validator) checkHeaderRules(report *model.Report) {
	if v.config.Header == nil {
		return
	}

	header := v.config.Header
	isConventional := v.config.Conventional != nil

	if header.Length != 0 {
		report.AddRule(rule.ValidateHeaderLength(v.config.Message, header.Length))
	}

	if header.Imperative {
		report.AddRule(rule.ValidateImperative(isConventional, v.config.Message))
	}

	if header.Case != "" {
		report.AddRule(rule.ValidateHeaderCase(isConventional, v.config.Message, header.Case))
	}

	if header.InvalidSuffix != "" {
		report.AddRule(rule.ValidateHeaderSuffix(v.config.HeaderFromMsg(), header.InvalidSuffix))
	}

	if header.Jira != nil {
		report.AddRule(rule.ValidateJira(v.config.Message, header.Jira.Keys, isConventional))
	}
}

func (v *Validator) checkSignatureRules(report *model.Report) {
	if v.config.DCO {
		report.AddRule(rule.ValidateSignOff(v.config.Message))
	}

	if v.config.GPG != nil && v.config.GPG.Required {
		report.AddRule(rule.ValidateSignature(v.git))

		if v.config.GPG.Identity != nil {
			report.AddRule(rule.ValidateGPGIdentity(v.config.Signature, v.config.RawCommit, v.config.GPG.Identity.PublicKeyURI))
		}
	}
}

func (v *Validator) checkConventionalRules(report *model.Report) {
	if v.config.Conventional != nil {
		conv := v.config.Conventional
		report.AddRule(rule.ValidateConventionalCommit(v.config.Message, conv.Types, conv.Scopes, conv.DescriptionLength))
	}
}

func (v *Validator) checkAdditionalRules(report *model.Report) {
	if v.config.SpellCheck != nil {
		report.AddRule(rule.ValidateSpelling(v.config.Message, v.config.SpellCheck.Locale))
	}

	if v.config.MaximumOfOneCommit {
		report.AddRule(rule.ValidateNumberOfCommits(v.git, v.options.CommitRef))
	}

	if v.config.Body != nil && v.config.Body.Required {
		report.AddRule(rule.ValidateCommitBody(v.config.Message))
	}
}
