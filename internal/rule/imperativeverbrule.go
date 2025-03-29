// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
	"github.com/pkg/errors"
)

// ImperativeVerbRule enforces that the first word of a commit message subject is an imperative verb.
type ImperativeVerbRule struct {
	errors []error
}

// Name returns the name of the rule.
func (i *ImperativeVerbRule) Name() string {
	return "ImperativeVerbRule"
}

// Result returns the validation result.
func (i *ImperativeVerbRule) Result() string {
	if len(i.errors) > 0 {
		return i.errors[0].Error()
	}

	return "Commit begins with imperative verb"
}

func (i *ImperativeVerbRule) Errors() []error {
	return i.errors
}

func (i *ImperativeVerbRule) SetErrors(err []error) {
	i.errors = err
}

func ValidateImperativeRule(subject string, isConventional bool) *ImperativeVerbRule {
	rule := &ImperativeVerbRule{}

	// Extract first word
	word, err := extractFirstWord(isConventional, subject)
	if err != nil {
		rule.errors = append(rule.errors, err)

		return rule
	}

	// Validate if the word is in imperative form
	if err := validateIsImperative(word); err != nil {
		rule.errors = append(rule.errors, err)
	}

	return rule
}

// extractFirstWord extracts the first word from the commit message.
func extractFirstWord(isConventional bool, subject string) (string, error) {
	var msg string

	if isConventional {
		groups := parseSubject(subject)
		if len(groups) != 5 {
			return "", errors.New("invalid conventional commit format")
		}

		msg = groups[4]
	} else {
		msg = subject
	}

	if msg == "" {
		return "", errors.New("empty message")
	}

	matches := firstWordRegex.FindStringSubmatch(msg)
	if len(matches) == 0 {
		return "", errors.New("no valid first word found")
	}

	return matches[0], nil
}

// validateIsImperative checks if a word is in imperative form using snowball stemming.
func validateIsImperative(word string) error {
	wordLower := strings.ToLower(word)

	// Check if the word is a non-imperative starter
	nonImperativeStarters := map[string]bool{
		"i": true, "we": true, "they": true, "he": true, "she": true, "it": true,
		"the": true, "a": true, "an": true, "this": true, "that": true,
		"these": true, "those": true, "my": true, "your": true, "our": true,
	}

	if nonImperativeStarters[wordLower] {
		return fmt.Errorf("first word of commit must be an imperative verb: %q is not a verb", word)
	}

	// Use snowball stemmer to get the base form
	stem, err := snowball.Stem(wordLower, "english", true)
	if err != nil {
		// If stemming fails, fall back to simpler checks
		return validateWithSimpleRules(wordLower)
	}

	// Check for specific non-imperative forms

	// Past tense verbs often end in "ed" and their stem is different
	if strings.HasSuffix(wordLower, "ed") && stem != wordLower && !isBaseFormWithEDEnding(wordLower) {
		return fmt.Errorf("first word of commit must be an imperative verb: %q appears to be past tense", word)
	}

	// Gerunds end in "ing"
	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		return fmt.Errorf("first word of commit must be an imperative verb: %q appears to be a gerund", word)
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !isBaseFormWithSEnding(wordLower) {
		return fmt.Errorf("first word of commit must be an imperative verb: %q appears to be 3rd person present", word)
	}

	return nil
}

// validateWithSimpleRules provides a fallback if stemming fails.
func validateWithSimpleRules(word string) error {
	// Simple pattern checks for non-imperative forms
	if strings.HasSuffix(word, "ed") && !isBaseFormWithEDEnding(word) {
		return fmt.Errorf("first word of commit must be an imperative verb: %q appears to be past tense", word)
	}

	if strings.HasSuffix(word, "ing") && len(word) > 4 {
		return fmt.Errorf("first word of commit must be an imperative verb: %q appears to be a gerund", word)
	}

	if strings.HasSuffix(word, "s") && !isBaseFormWithSEnding(word) && len(word) > 2 {
		return fmt.Errorf("first word of commit must be an imperative verb: %q appears to be 3rd person present", word)
	}

	return nil
}

// isBaseFormWithEDEnding checks if a word ending in "ed" is actually a base form.
func isBaseFormWithEDEnding(word string) bool {
	baseFormsEndingWithED := map[string]bool{
		"shed":    true,
		"embed":   true,
		"speed":   true,
		"proceed": true,
		"exceed":  true,
		"succeed": true,
		"feed":    true,
		"need":    true,
		"breed":   true,
	}

	return baseFormsEndingWithED[word]
}

// isBaseFormWithSEnding checks if a word ending in "s" is actually a base form.
func isBaseFormWithSEnding(word string) bool {
	baseFormsEndingWithS := map[string]bool{
		"focus":   true,
		"process": true,
		"pass":    true,
		"address": true,
		"express": true,
		"dismiss": true,
		"access":  true,
		"press":   true,
		"cross":   true,
		"miss":    true,
		"toss":    true,
		"guess":   true,
		"dress":   true,
		"bless":   true,
		"stress":  true,
	}

	return baseFormsEndingWithS[word]
}

// firstWordRegex is the regular expression used to find the first word in a commit.
var firstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

// parseSubject parses a conventional commit subject line.
func parseSubject(msg string) []string {
	subject := strings.Split(msg, "\n")[0]
	groups := SubjectRegex.FindStringSubmatch(subject)

	return groups
}
