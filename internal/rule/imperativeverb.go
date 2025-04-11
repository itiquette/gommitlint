// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"errors"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/model"
	"github.com/kljensen/snowball"
)

// firstWordRegex is the regular expression used to find the first word in a commit.
var firstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

// ImperativeVerb enforces that commit messages begin with a verb in the imperative mood.
//
// This rule helps maintain consistent and actionable commit messages by ensuring they
// follow the Git convention of using imperative verbs. Imperative mood describes the
// action that the commit will perform when applied, making commit messages more
// readable and aligned with Git's own commit messages.
//
// The rule validates that the first word of the commit message (or the first word after
// the conventional commit prefix) is an imperative verb and not:
//   - A past tense verb (e.g., "Added", "Fixed")
//   - A gerund (e.g., "Adding", "Fixing")
//   - A third-person present verb (e.g., "Adds", "Fixes")
//   - A non-verb like articles or pronouns (e.g., "The", "A", "This")
//
// Examples:
//
//   - For standard commits:
//     "Add feature" would pass
//     "Added feature" would fail (past tense)
//     "Adding feature" would fail (gerund)
//     "Adds feature" would fail (third person)
//     "The feature" would fail (article)
//
//   - For conventional commits:
//     "feat(auth): Add login form" would pass
//     "fix: Resolve memory leak" would pass
//     "feat(ui): Added button" would fail (past tense)
//     "chore: Updating dependencies" would fail (gerund)
//
// The rule uses the Snowball stemming algorithm to detect non-imperative forms
// and has special handling for words that naturally end in "ed" or "s".
type ImperativeVerb struct {
	errors []*model.ValidationError
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
func (rule ImperativeVerb) Errors() []*model.ValidationError {
	return rule.errors
}

// SetErrors sets the errors for this rule (for testing).
func (rule *ImperativeVerb) SetErrors(errs []*model.ValidationError) {
	rule.errors = errs
}

// Help returns a description of how to fix the rule violation.
func (rule ImperativeVerb) Help() string {
	if len(rule.errors) == 0 {
		return "No errors to fix"
	}

	// Check error code
	if len(rule.errors) > 0 {
		switch rule.errors[0].Code {
		case "invalid_format":
			return `Format your commit message according to the Conventional Commits specification.
Example: feat(auth): Add login feature
The correct format is: type(scope): subject
- type: feat, fix, docs, etc.
- scope: optional context (in parentheses)
- subject: description of the change`

		case "empty_message":
			return "Provide a non-empty commit message with a verb in the imperative mood."

		case "missing_subject":
			return "Add a description after the type and colon in your conventional commit message."

		case "no_first_word":
			return "Start your commit message with a word (letters or numbers). Remove any leading special characters."

		case "non_verb":
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

		case "past_tense":
			return `Avoid using past tense verbs at the start of commit messages.
Instead of "Added feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`

		case "gerund":
			return `Avoid using gerund (-ing) forms at the start of commit messages.
Instead of "Adding feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`

		case "third_person":
			return `Avoid using third-person present verbs at the start of commit messages.
Instead of "Adds feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`
		}
	}

	// Default help using message content if available
	errMsg := rule.errors[0].Message

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

// addError adds a structured validation error.
func (rule *ImperativeVerb) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("ImperativeVerb", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	rule.errors = append(rule.errors, err)
}

// ValidateImperative validates that the first word of a commit message is in imperative form.
//
// Parameters:
//   - subject: The commit subject line to validate
//   - isConventional: Whether to parse as a conventional commit
//
// For conventional commits (format: "type(scope): subject"), it validates the
// first word after the prefix. For example, in "feat(auth): Add login", it validates "Add".
//
// For standard commits, it validates the first word of the subject.
// For example, in "Fix memory leak", it validates "Fix".
//
// The validation uses the Snowball stemming algorithm to check for non-imperative
// forms and has special handling for words that naturally end in "ed" or "s" but
// are already in imperative form (e.g., "proceed", "focus").
//
// Returns:
//   - An ImperativeVerb instance with validation results
func ValidateImperative(subject string, isConventional bool) ImperativeVerb {
	rule := ImperativeVerb{}

	// Check for empty message first
	if subject == "" {
		rule.addError(
			"empty_message",
			"empty message",
			nil,
		)

		return rule
	}

	// Check for conventional commit format issues
	if isConventional {
		// Check for empty subject after colon
		if regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ]$`).MatchString(subject) {
			rule.addError(
				"missing_subject",
				"missing subject after type",
				map[string]string{
					"subject": subject,
				},
			)

			return rule
		}

		// Check for invalid format
		if !regexp.MustCompile(`^([a-z]+)(?:\(([\w,/-]+)\))?(!)?:[ ]`).MatchString(subject) {
			rule.addError(
				"invalid_format",
				"invalid conventional commit format",
				map[string]string{
					"subject": subject,
				},
			)

			return rule
		}

		// Check for invalid type (uppercase type or without space after colon)
		if !SubjectRegex.MatchString(subject) {
			rule.addError(
				"invalid_format",
				"invalid conventional commit format",
				map[string]string{
					"subject": subject,
				},
			)

			return rule
		}
	}

	// Extract first word
	word, err := extractFirstWord(isConventional, subject)
	if err != nil {
		// Map generic errors to specific error codes
		if strings.Contains(err.Error(), "invalid conventional commit format") {
			rule.addError(
				"invalid_format",
				err.Error(),
				map[string]string{
					"subject": subject,
				},
			)
		} else if strings.Contains(err.Error(), "missing subject after type") {
			rule.addError(
				"missing_subject",
				err.Error(),
				map[string]string{
					"subject": subject,
				},
			)
		} else if strings.Contains(err.Error(), "no valid first word found") {
			rule.addError(
				"no_first_word",
				err.Error(),
				map[string]string{
					"subject": subject,
				},
			)
		} else {
			// Generic error fallback
			rule.addError(
				"validation_error",
				err.Error(),
				map[string]string{
					"subject": subject,
				},
			)
		}

		return rule
	}

	// Validate if the word is in imperative form
	validateIsImperative(word, &rule)

	return rule
}

// extractFirstWord extracts the first word from the commit message.
//
// Parameters:
//   - isConventional: Whether to parse as a conventional commit
//   - subject: The commit subject line
//
// For conventional commits, it extracts the first word after the "type(scope): " prefix.
// For standard commits, it extracts the first word of the subject.
//
// Returns:
//   - The first word to validate
//   - Any error encountered during extraction
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
//
// Parameters:
//   - word: The word to validate
//   - rule: The ImperativeVerb rule to populate with errors
//
// The function first checks if the word is a non-verb (like articles or pronouns).
// Then it uses the Snowball stemming algorithm to detect past tense, gerund, and
// third-person forms. Special handling is provided for words that naturally end in
// "ed" or "s" but are already in imperative form.
func validateIsImperative(word string, rule *ImperativeVerb) {
	wordLower := strings.ToLower(word)

	// Check if the word is a non-imperative starter
	nonImperativeStarters := map[string]bool{
		"i": true, "we": true, "they": true, "he": true, "she": true, "it": true,
		"the": true, "a": true, "an": true, "this": true, "that": true,
		"these": true, "those": true, "my": true, "your": true, "our": true,
	}

	if nonImperativeStarters[wordLower] {
		rule.addError(
			"non_verb",
			"first word of commit must be an imperative verb: \""+word+"\" is not a verb",
			map[string]string{
				"word": word,
				"type": "non_verb",
			},
		)

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
		rule.addError(
			"past_tense",
			"first word of commit must be an imperative verb: \""+word+"\" appears to be past tense",
			map[string]string{
				"word": word,
				"type": "past_tense",
			},
		)

		return
	}

	// Gerunds end in "ing"
	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		rule.addError(
			"gerund",
			"first word of commit must be an imperative verb: \""+word+"\" appears to be a gerund",
			map[string]string{
				"word": word,
				"type": "gerund",
			},
		)

		return
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !isBaseFormWithSEnding(wordLower) {
		rule.addError(
			"third_person",
			"first word of commit must be an imperative verb: \""+word+"\" appears to be 3rd person present",
			map[string]string{
				"word": word,
				"type": "third_person",
			},
		)

		return
	}
}

// validateWithSimpleRules provides a fallback if stemming fails.
//
// Parameters:
//   - wordLower: The lowercase version of the word to validate
//   - originalWord: The original word (preserving case)
//   - rule: The ImperativeVerb rule to populate with errors
//
// This function performs simple suffix-based checks to detect non-imperative forms
// when the Snowball stemming algorithm fails.
func validateWithSimpleRules(wordLower, originalWord string, rule *ImperativeVerb) {
	// Simple pattern checks for non-imperative forms
	if strings.HasSuffix(wordLower, "ed") && !isBaseFormWithEDEnding(wordLower) {
		rule.addError(
			"past_tense",
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be past tense",
			map[string]string{
				"word": originalWord,
				"type": "past_tense",
			},
		)

		return
	}

	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		rule.addError(
			"gerund",
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be a gerund",
			map[string]string{
				"word": originalWord,
				"type": "gerund",
			},
		)

		return
	}

	if strings.HasSuffix(wordLower, "s") && !isBaseFormWithSEnding(wordLower) && len(wordLower) > 2 {
		rule.addError(
			"third_person",
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be 3rd person present",
			map[string]string{
				"word": originalWord,
				"type": "third_person",
			},
		)

		return
	}
}

// isBaseFormWithEDEnding checks if a word ending in "ed" is actually a base form.
//
// Parameters:
//   - word: The word to check
//
// Some verbs naturally end in "ed" in their imperative form, such as "proceed"
// or "exceed". This function checks against a list of such words.
//
// Returns:
//   - true if the word is already in imperative form despite ending in "ed"
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
//
// Parameters:
//   - word: The word to check
//
// Some verbs naturally end in "s" in their imperative form, such as "focus"
// or "process". This function checks against a list of such words.
//
// Returns:
//   - true if the word is already in imperative form despite ending in "s"
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
//
// Parameters:
//   - msg: The commit message to parse
//
// Returns:
//   - A slice of strings containing the matched groups from the subject line
func parseSubject(msg string) []string {
	subject := strings.Split(msg, "\n")[0]
	groups := SubjectRegex.FindStringSubmatch(subject)

	return groups
}
