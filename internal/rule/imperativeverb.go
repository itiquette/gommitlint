// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"
)

// firstWordRegex is the regular expression used to find the first word in a commit.
var firstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

type ImperativeVerb struct {
	errors []error
}

// Name returns the name of the rule.
func (rule *ImperativeVerb) Name() string {
	return "ImperativeVerb"
}

// Result returns the validation result.
func (rule ImperativeVerb) Result() string {
	if len(rule.errors) > 0 {
		return rule.errors[0].Error()
	}

	return "Commit begins with imperative verb"
}

// Errors returns any violations of the rule.
func (rule ImperativeVerb) Errors() []error {
	return rule.errors
}

// SetErrors sets the errors for this rule.
func (rule *ImperativeVerb) SetErrors(err []error) {
	rule.errors = err
}

// Help returns a description of how to fix the rule violation.
func (rule ImperativeVerb) Help() string {
	if len(rule.errors) == 0 {
		return "No errors to fix"
	}

	errMsg := rule.errors[0].Error()

	if strings.Contains(errMsg, "invalid conventional commit format") {
		return `Format your commit message according to the Conventional Commits specification.
Example: feat(auth): Add login feature

The correct format is: type(scope): subject
- type: feat, fix, docs, etc.
- scope: optional context (in parentheses)
- subject: description of the change`
	}

	if strings.Contains(errMsg, "empty message") {
		return "Provide a non-empty commit message with a verb in the imperative mood."
	}

	if strings.Contains(errMsg, "missing subject after type") {
		return "Add a description after the type and colon in your conventional commit message."
	}

	if strings.Contains(errMsg, "no valid first word found") {
		return "Start your commit message with a word (letters or numbers). Remove any leading special characters."
	}

	if strings.Contains(errMsg, "first word of commit must be an imperative verb") {
		return `Use the imperative mood for the first word in your commit message.

Examples of imperative verbs:
- Add, Fix, Update, Remove, Change, Refactor, Implement

Avoid:
- Past tense: Added, Fixed, Updated
- Gerund: Adding, Fixing, Updating
- 3rd person: Adds, Fixes, Updates
- Articles or pronouns: The, A, This, I, We

The imperative form is preferred because it completes the sentence:
"If applied, this commit will [your commit message]"`
	}

	// Default help
	return "Use the imperative mood for the first word in your commit message (e.g., 'Add feature' not 'Added feature')."
}

// addErrorf adds an error to the rule's errors slice.
func (rule *ImperativeVerb) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}

// ValidateImperative validates that the first word is in imperative form.
func ValidateImperative(subject string, isConventional bool) ImperativeVerb {
	rule := ImperativeVerb{}

	// Check for empty message first
	if subject == "" {
		rule.addErrorf("empty message")

		return rule
	}

	// Check for conventional commit format issues
	if isConventional {
		// Check for empty subject after colon
		if regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ]$`).MatchString(subject) {
			rule.addErrorf("missing subject after type")

			return rule
		}

		// Check for invalid format
		if !regexp.MustCompile(`^([a-z]+)(?:\(([\w,/-]+)\))?(!)?:[ ]`).MatchString(subject) {
			rule.addErrorf("invalid conventional commit format")

			return rule
		}

		// Check for invalid type (uppercase type or without space after colon)
		if !SubjectRegex.MatchString(subject) {
			rule.addErrorf("invalid conventional commit format")

			return rule
		}
	}

	// Extract first word
	word, err := extractFirstWord(isConventional, subject)
	if err != nil {
		rule.addErrorf("%v", err)

		return rule
	}

	// Validate if the word is in imperative form
	validateIsImperative(word, &rule)

	return rule
}

// extractFirstWord extracts the first word from the commit message.
func extractFirstWord(isConventional bool, subject string) (string, error) {
	if isConventional {
		groups := parseSubject(subject)
		// Validate conventional commit format
		if len(groups) != 5 {
			return "", errors.New("invalid conventional commit format")
		}

		msg := groups[4]
		if msg == "" {
			return "", errors.New("missing subject after type")
		}

		matches := firstWordRegex.FindStringSubmatch(msg)
		if len(matches) == 0 {
			return "", errors.New("no valid first word found")
		}

		return matches[0], nil
	}

	matches := firstWordRegex.FindStringSubmatch(subject)
	if len(matches) == 0 {
		return "", errors.New("no valid first word found")
	}

	return matches[0], nil
}

// validateIsImperative checks if a word is in imperative form using snowball stemming.
func validateIsImperative(word string, rule *ImperativeVerb) {
	wordLower := strings.ToLower(word)

	// Check if the word is a non-imperative starter
	nonImperativeStarters := map[string]bool{
		"i": true, "we": true, "they": true, "he": true, "she": true, "it": true,
		"the": true, "a": true, "an": true, "this": true, "that": true,
		"these": true, "those": true, "my": true, "your": true, "our": true,
	}

	if nonImperativeStarters[wordLower] {
		rule.addErrorf("first word of commit must be an imperative verb: %q is not a verb", word)

		return
	}

	// Use snowball stemmer to get the base form
	stem, err := snowball.Stem(wordLower, "english", true)
	if err != nil {
		// If stemming fails, fall back to simpler checks
		validateWithSimpleRules(wordLower, word, rule)

		return
	}

	// Check for specific non-imperative forms

	// Past tense verbs often end in "ed" and their stem is different
	if strings.HasSuffix(wordLower, "ed") && stem != wordLower && !isBaseFormWithEDEnding(wordLower) {
		rule.addErrorf("first word of commit must be an imperative verb: %q appears to be past tense", word)

		return
	}

	// Gerunds end in "ing"
	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		rule.addErrorf("first word of commit must be an imperative verb: %q appears to be a gerund", word)

		return
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !isBaseFormWithSEnding(wordLower) {
		rule.addErrorf("first word of commit must be an imperative verb: %q appears to be 3rd person present", word)

		return
	}
}

// validateWithSimpleRules provides a fallback if stemming fails.
func validateWithSimpleRules(wordLower, originalWord string, rule *ImperativeVerb) {
	// Simple pattern checks for non-imperative forms
	if strings.HasSuffix(wordLower, "ed") && !isBaseFormWithEDEnding(wordLower) {
		rule.addErrorf("first word of commit must be an imperative verb: %q appears to be past tense", originalWord)

		return
	}

	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		rule.addErrorf("first word of commit must be an imperative verb: %q appears to be a gerund", originalWord)

		return
	}

	if strings.HasSuffix(wordLower, "s") && !isBaseFormWithSEnding(wordLower) && len(wordLower) > 2 {
		rule.addErrorf("first word of commit must be an imperative verb: %q appears to be 3rd person present", originalWord)

		return
	}
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

// parseSubject parses a conventional commit subject line.
func parseSubject(msg string) []string {
	subject := strings.Split(msg, "\n")[0]
	groups := SubjectRegex.FindStringSubmatch(subject)

	return groups
}
