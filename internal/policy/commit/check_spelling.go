// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"

	"github.com/janderssonse/gommitlint/internal/policy"
)

// SpellCheck represents to spell check policy.
type SpellCheck struct {
	Locale string `mapstructure:"locale"`
}

// SpellingCheck enforces correct spelling.
type SpellingCheck struct {
	errors []error
}

// Name returns the name of the check.
func (h SpellingCheck) Name() string {
	return "Spellcheck"
}

// Message returns to check message.
func (h SpellingCheck) Message() string {
	return fmt.Sprintf("Commit contains %d misspellings", len(h.errors))
}

// Errors returns any violations of the check.
func (h SpellingCheck) Errors() []error {
	return h.errors
}

// ValidateSpelling checks the spelling.
func (commit Commit) ValidateSpelling() policy.Check { //nolint:ireturn
	check := &SpellingCheck{}

	replacer := misspell.Replacer{
		Replacements: misspell.DictMain,
	}

	switch strings.ToUpper(commit.SpellCheck.Locale) {
	case "":
	case "US":
		replacer.AddRuleList(misspell.DictAmerican)
	case "UK", "GB":
		replacer.AddRuleList(misspell.DictBritish)
	case "NZ", "AU", "CA":
		check.errors = append(check.errors, fmt.Errorf("unknown locale: %q", commit.SpellCheck.Locale))
	}

	replacer.Compile()

	_, diffs := replacer.Replace(commit.msg)

	for _, diff := range diffs {
		check.errors = append(check.errors, fmt.Errorf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected))
	}

	return check
}
