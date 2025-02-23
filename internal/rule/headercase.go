// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rule

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// HeaderCase enforces the case of the first word in the header.
type HeaderCase struct {
	headerCase string
	RuleErrors []error
}

// Name returns the name of the check.
func (h *HeaderCase) Name() string {
	return "Header Case"
}

// Message returns the check message.
func (h *HeaderCase) Message() string {
	if len(h.RuleErrors) > 0 {
		return h.RuleErrors[0].Error()
	}

	return "Header case is valid"
}

// Errors returns any violations of the check.
func (h *HeaderCase) Errors() []error {
	return h.RuleErrors
}

// ValidateHeaderCase checks the header case based on the specified case choice.
func ValidateHeaderCase(isConventional bool, message, caseChoice string) interfaces.Rule {
	rule := &HeaderCase{headerCase: caseChoice}

	// Extract first word
	firstWord, err := extractFirstWord(isConventional, message)
	if err != nil {
		rule.RuleErrors = append(rule.RuleErrors, err)

		return rule
	}

	// Decode first rune
	first, _ := utf8.DecodeRuneInString(firstWord)
	if first == utf8.RuneError {
		rule.RuleErrors = append(rule.RuleErrors, errors.New("header does not start with valid UTF-8 text"))

		return rule
	}

	// Validate case
	var valid bool

	switch caseChoice {
	case "upper":
		valid = unicode.IsUpper(first)
	case "lower":
		valid = unicode.IsLower(first)
	default:
		rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("invalid configured case: %s", caseChoice))

		return rule
	}

	if !valid {
		rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("commit header case is not %s", caseChoice))
	}

	return rule
}
