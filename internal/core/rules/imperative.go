// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
// Package rules provides validation rules for Git commits.
package rules

import (
	"errors"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/kljensen/snowball"
)

// firstWordRegex is the regular expression used to find the first word in a commit.
var firstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

// conventionalCommitRegex is used to validate and parse conventional commit messages.
var conventionalCommitRegex = regexp.MustCompile(`^([a-z]+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// ImperativeVerbRule validates that commit messages begin with a verb in the imperative mood.
//
// This rule helps maintain consistent and actionable commit messages by ensuring they
// follow the Git convention of using imperative verbs. Imperative mood describes the
// action that the commit will perform when applied, making commit messages more
// readable and aligned with Git's own commit messages.
type ImperativeVerbRule struct {
	isConventional              bool
	customNonImperativeStarters map[string]bool
	baseFormsEndingWithED       map[string]bool
	baseFormsEndingWithS        map[string]bool
	errors                      []appErrors.ValidationError
	firstWord                   string // Store the first word for verbose output
	verbType                    string // Store the type of verb issue for verbose output
}

// ImperativeVerbOption is a function that modifies an ImperativeVerbRule.
type ImperativeVerbOption func(ImperativeVerbRule) ImperativeVerbRule

// WithImperativeConventionalCommit configures whether to treat the commit as a conventional commit.
func WithImperativeConventionalCommit(isConventional bool) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		rule.isConventional = isConventional

		return rule
	}
}

// WithCustomNonImperativeStarters adds custom words to consider as non-imperative starters.
func WithCustomNonImperativeStarters(words map[string]bool) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		newStarters := make(map[string]bool)
		// Copy existing starters
		for k, v := range rule.customNonImperativeStarters {
			newStarters[k] = v
		}
		// Add new starters
		for word, val := range words {
			newStarters[strings.ToLower(word)] = val
		}

		rule.customNonImperativeStarters = newStarters

		return rule
	}
}

// WithAdditionalBaseFormsEndingWithED adds custom words that naturally end with "ed"
// but are already in imperative form.
func WithAdditionalBaseFormsEndingWithED(words map[string]bool) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		newForms := make(map[string]bool)
		// Copy existing forms
		for k, v := range rule.baseFormsEndingWithED {
			newForms[k] = v
		}
		// Add new forms
		for word, val := range words {
			newForms[strings.ToLower(word)] = val
		}

		rule.baseFormsEndingWithED = newForms

		return rule
	}
}

// WithAdditionalBaseFormsEndingWithS adds custom words that naturally end with "s"
// but are already in imperative form.
func WithAdditionalBaseFormsEndingWithS(words map[string]bool) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		newForms := make(map[string]bool)
		// Copy existing forms
		for k, v := range rule.baseFormsEndingWithS {
			newForms[k] = v
		}
		// Add new forms
		for word, val := range words {
			newForms[strings.ToLower(word)] = val
		}

		rule.baseFormsEndingWithS = newForms

		return rule
	}
}

// NewImperativeVerbRule creates a new ImperativeVerbRule.
func NewImperativeVerbRule(isConventional bool, options ...ImperativeVerbOption) ImperativeVerbRule {
	rule := ImperativeVerbRule{
		isConventional:              isConventional,
		customNonImperativeStarters: make(map[string]bool),
		baseFormsEndingWithED: map[string]bool{
			"shed":    true,
			"embed":   true,
			"speed":   true,
			"proceed": true,
			"exceed":  true,
			"succeed": true,
			"feed":    true,
			"need":    true,
			"breed":   true,
		},
		baseFormsEndingWithS: map[string]bool{
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
		},
		errors: make([]appErrors.ValidationError, 0),
	}
	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r ImperativeVerbRule) Name() string {
	return "ImperativeVerb"
}

// Validate performs validation against a commit and returns any errors.
func (r ImperativeVerbRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// This returns a new validated rule state rather than modifying the existing one
	newRule := r.validateCommit(commit)

	return newRule.errors
}

// validateCommit performs the actual validation and returns a new rule with updated state.
func (r ImperativeVerbRule) validateCommit(commit domain.CommitInfo) ImperativeVerbRule {
	// Create a new rule with reset state for validation
	newRule := r
	newRule.errors = make([]appErrors.ValidationError, 0)
	newRule.firstWord = ""
	newRule.verbType = ""

	subject := commit.Subject

	// Check for empty message
	if subject == "" {
		return newRule.addError(
			appErrors.ErrMissingSubject,
			"Commit message is empty",
			map[string]string{
				"word": "(empty message)",
			},
		)
	}

	// Check format
	newRule = newRule.validateFormat(subject)
	if len(newRule.errors) > 0 {
		return newRule
	}

	// Extract and validate first word
	return newRule.validateFirstWord(subject)
}

// validateFormat checks if the commit message follows the expected format.
func (r ImperativeVerbRule) validateFormat(subject string) ImperativeVerbRule {
	newRule := r

	if r.isConventional {
		// Check for conventional commit format
		if !conventionalCommitRegex.MatchString(subject) {
			// If it doesn't match the pattern fully, figure out specific issue
			if regexp.MustCompile(`^([a-z]+)(?:\(([\w,/-]+)\))?(!)?:[ ]$`).MatchString(subject) {
				// Empty subject after colon
				return newRule.addError(
					appErrors.ErrMissingSubject,
					"Missing subject after type in conventional commit",
					map[string]string{
						"subject": subject,
						"word":    "(missing subject)",
					},
				)
			}
			// General format issue
			return newRule.addError(
				appErrors.ErrInvalidFormat,
				"Invalid conventional commit format",
				map[string]string{
					"subject": subject,
					"word":    "(invalid format)",
				},
			)
		}
	}
	// For non-conventional commits, check if it starts with a valid character
	if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(strings.TrimSpace(subject)) {
		return newRule.addError(
			appErrors.ErrInvalidFormat,
			"Invalid commit message format",
			map[string]string{
				"subject": subject,
				"word":    "(invalid format)",
			},
		)
	}

	return newRule
}

// validateFirstWord extracts and validates the first word from the commit message.
func (r ImperativeVerbRule) validateFirstWord(subject string) ImperativeVerbRule {
	newRule := r

	// Extract first word
	word, err := r.extractFirstWord(subject)
	if err != nil {
		// Map generic errors to specific error codes
		if strings.Contains(err.Error(), "invalid conventional commit format") {
			return newRule.addError(
				appErrors.ErrInvalidFormat,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(invalid format)",
				},
			)
		} else if strings.Contains(err.Error(), "missing subject after type") {
			return newRule.addError(
				appErrors.ErrMissingSubject,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(missing subject)",
				},
			)
		} else if strings.Contains(err.Error(), "no valid first word found") {
			return newRule.addError(
				appErrors.ErrNoFirstWord,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(no valid first word)",
				},
			)
		}
		// Generic error fallback
		return newRule.addError(
			appErrors.ErrUnknown,
			err.Error(),
			map[string]string{
				"subject": subject,
				"word":    "(validation error)",
			},
		)
	}

	// Store the first word for verbose output
	newRule.firstWord = word

	// Validate if the word is in imperative form
	return newRule.validateIsImperative(word)
}

// Result returns a concise result message.
func (r ImperativeVerbRule) Result() string {
	if len(r.errors) > 0 {
		return "Non-imperative verb detected"
	}

	return "Commit begins with imperative verb"
}

// VerboseResult returns a detailed result message.
func (r ImperativeVerbRule) VerboseResult() string {
	if len(r.errors) > 0 {
		// Return a more detailed error message based on error code
		switch appErrors.ValidationErrorCode(r.errors[0].Code) { //nolint:exhaustive
		case appErrors.ErrInvalidFormat:
			if strings.Contains(r.errors[0].Message, "conventional") {
				return "Invalid conventional commit format. Must follow 'type(scope): description' pattern."
			}

			return "Invalid commit format. Commit message should start with an imperative verb."
		case appErrors.ErrEmptyMessage:
			return "Commit message is empty. Cannot validate imperative verb."
		case appErrors.ErrMissingSubject:
			return "Missing subject after conventional commit type. Nothing to validate."
		case appErrors.ErrNoFirstWord:
			return "No valid first word found in commit message."
		case appErrors.ErrNonVerb:
			return "'" + r.firstWord + "' is a non-verb word (article, pronoun, etc.). Use an action verb instead."
		case appErrors.ErrPastTense:
			return "'" + r.firstWord + "' is in past tense. Use present imperative form instead (e.g., 'Add' not 'Added')."
		case appErrors.ErrGerund:
			return "'" + r.firstWord + "' is a gerund (-ing form). Use present imperative form instead (e.g., 'Add' not 'Adding')."
		case appErrors.ErrThirdPerson:
			return "'" + r.firstWord + "' is in third person form. Use present imperative form instead (e.g., 'Add' not 'Adds')."
		default:
			// If we have a first word but no specific code match, include it in the error
			if r.firstWord != "" {
				return "Word '" + r.firstWord + "' detected: " + r.errors[0].Error()
			}

			return r.errors[0].Error()
		}
	}

	return "Commit begins with proper imperative verb '" + r.firstWord + "'"
}

// Help returns guidance on how to fix rule violations.
func (r ImperativeVerbRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}
	// Check error code
	if len(r.errors) > 0 {
		switch appErrors.ValidationErrorCode(r.errors[0].Code) { //nolint:exhaustive
		case appErrors.ErrInvalidFormat:
			if strings.Contains(r.errors[0].Message, "conventional") {
				return `Format your commit message according to the Conventional Commits specification.
Example: feat(auth): Add login feature
The correct format is: type(scope): subject
- type: feat, fix, docs, etc.
- scope: optional context (in parentheses)
- subject: description of the change`
			}

			return "Make sure your commit message starts with an imperative verb (e.g., 'Add', 'Fix', 'Update')."
		case appErrors.ErrEmptyMessage:
			return "Provide a non-empty commit message with a verb in the imperative mood."
		case appErrors.ErrMissingSubject:
			return "Add a description after the type and colon in your conventional commit message."
		case appErrors.ErrNoFirstWord:
			return "Start your commit message with a word (letters or numbers). Remove any leading special characters."
		case appErrors.ErrNonVerb:
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
		case appErrors.ErrPastTense:
			return `Avoid using past tense verbs at the start of commit messages.
Instead of "Added feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`
		case appErrors.ErrGerund:
			return `Avoid using gerund (-ing) forms at the start of commit messages.
Instead of "Adding feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`
		case appErrors.ErrThirdPerson:
			return `Avoid using third-person present verbs at the start of commit messages.
Instead of "Adds feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`
		}
	}
	// Default help based on message content
	errMsg := r.errors[0].Message
	if strings.Contains(errMsg, "invalid conventional commit format") {
		return `Format your commit message according to the Conventional Commits specification.
Example: feat(auth): Add login feature
The correct format is: type(scope): subject
- type: feat, fix, docs, etc.
- scope: optional context (in parentheses)
- subject: description of the change`
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

// Errors returns all validation errors.
func (r ImperativeVerbRule) Errors() []appErrors.ValidationError {
	return r.errors
}

// extractFirstWord extracts the first word from the commit message.
func (r ImperativeVerbRule) extractFirstWord(subject string) (string, error) {
	if r.isConventional {
		matches := conventionalCommitRegex.FindStringSubmatch(subject)
		// Validate conventional commit format
		if len(matches) != 5 {
			return "", errors.New("invalid conventional commit format")
		}

		msg := matches[4]
		if msg == "" {
			return "", errors.New("missing subject after type")
		}

		matches = firstWordRegex.FindStringSubmatch(msg)
		if len(matches) == 0 {
			return "", errors.New("no valid first word found")
		}

		return matches[1], nil
	}

	matches := firstWordRegex.FindStringSubmatch(subject)
	if len(matches) == 0 {
		return "", errors.New("no valid first word found")
	}

	return matches[1], nil
}

// validateIsImperative checks if a word is in imperative form using snowball stemming.
func (r ImperativeVerbRule) validateIsImperative(word string) ImperativeVerbRule {
	newRule := r
	wordLower := strings.ToLower(word)

	// Default non-imperative starters
	nonImperativeStarters := map[string]bool{
		"i": true, "we": true, "they": true, "he": true, "she": true, "it": true,
		"the": true, "a": true, "an": true, "this": true, "that": true,
		"these": true, "those": true, "my": true, "your": true, "our": true,
	}

	// Add custom non-imperative starters if provided
	for word, val := range r.customNonImperativeStarters {
		nonImperativeStarters[strings.ToLower(word)] = val
	}

	if nonImperativeStarters[wordLower] {
		return newRule.addError(
			appErrors.ErrNonVerb,
			"first word of commit must be an imperative verb: \""+word+"\" is not a verb",
			map[string]string{
				"word": word,
				"type": "non_verb",
			},
		)
	}

	// Use snowball stemmer to get the base form
	stem, err := snowball.Stem(wordLower, "english", true)
	if err != nil {
		// If stemming fails, fall back to simpler checks
		return newRule.validateWithSimpleRules(wordLower, word)
	}

	// Check for specific non-imperative forms
	// Past tense verbs often end in "ed" and their stem is different
	if strings.HasSuffix(wordLower, "ed") && stem != wordLower && !newRule.isBaseFormWithEDEnding(wordLower) {
		return newRule.addError(
			appErrors.ErrPastTense,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be past tense",
			map[string]string{
				"word": word,
				"type": "past_tense",
			},
		)
	}

	// Gerunds end in "ing"
	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		return newRule.addError(
			appErrors.ErrGerund,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be a gerund",
			map[string]string{
				"word": word,
				"type": "gerund",
			},
		)
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !newRule.isBaseFormWithSEnding(wordLower) {
		return newRule.addError(
			appErrors.ErrThirdPerson,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be 3rd person present",
			map[string]string{
				"word": word,
				"type": "third_person",
			},
		)
	}

	return newRule
}

// validateWithSimpleRules provides a fallback if stemming fails.
func (r ImperativeVerbRule) validateWithSimpleRules(wordLower, originalWord string) ImperativeVerbRule {
	newRule := r

	// Simple pattern checks for non-imperative forms
	if strings.HasSuffix(wordLower, "ed") && !newRule.isBaseFormWithEDEnding(wordLower) {
		return newRule.addError(
			appErrors.ErrPastTense,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be past tense",
			map[string]string{
				"word": originalWord,
				"type": "past_tense",
			},
		)
	}

	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		return newRule.addError(
			appErrors.ErrGerund,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be a gerund",
			map[string]string{
				"word": originalWord,
				"type": "gerund",
			},
		)
	}

	if strings.HasSuffix(wordLower, "s") && !newRule.isBaseFormWithSEnding(wordLower) && len(wordLower) > 2 {
		return newRule.addError(
			appErrors.ErrThirdPerson,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be 3rd person present",
			map[string]string{
				"word": originalWord,
				"type": "third_person",
			},
		)
	}

	return newRule
}

// isBaseFormWithEDEnding checks if a word ending in "ed" is actually a base form.
func (r ImperativeVerbRule) isBaseFormWithEDEnding(word string) bool {
	return r.baseFormsEndingWithED[word]
}

// isBaseFormWithSEnding checks if a word ending in "s" is actually a base form.
func (r ImperativeVerbRule) isBaseFormWithSEnding(word string) bool {
	return r.baseFormsEndingWithS[word]
}

// addError adds a structured validation error.
func (r ImperativeVerbRule) addError(code appErrors.ValidationErrorCode, message string, context map[string]string) ImperativeVerbRule {
	// Create a validation error with context in one step
	err := appErrors.New(r.Name(), code, message, appErrors.WithContextMap(context))

	// Create a new rule instance
	newRule := r

	// Store verb type for verbose output if provided
	if context != nil {
		if wordType, ok := context["type"]; ok {
			newRule.verbType = wordType
		}

		if word, ok := context["word"]; ok {
			newRule.firstWord = word
		} else if code == appErrors.ErrInvalidFormat {
			newRule.firstWord = "(invalid format)"
		} else if code == appErrors.ErrMissingSubject {
			newRule.firstWord = "(missing subject)"
		} else if code == appErrors.ErrEmptyMessage {
			newRule.firstWord = "(empty message)"
		} else if code == appErrors.ErrNoFirstWord {
			newRule.firstWord = "(no valid first word)"
		}
	}

	// Add error to the new rule's slice
	newRule.errors = append(newRule.errors, err)

	return newRule
}
