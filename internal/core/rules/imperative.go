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
	errors                      []*domain.ValidationError
	firstWord                   string // Store the first word for verbose output
	verbType                    string // Store the type of verb issue for verbose output
}

// ImperativeVerbOption is a function that modifies an ImperativeVerbRule.
type ImperativeVerbOption func(*ImperativeVerbRule)

// WithImperativeConventionalCommit configures whether to treat the commit as a conventional commit.
func WithImperativeConventionalCommit(isConventional bool) ImperativeVerbOption {
	return func(rule *ImperativeVerbRule) {
		rule.isConventional = isConventional
	}
}

// WithCustomNonImperativeStarters adds custom words to consider as non-imperative starters.
func WithCustomNonImperativeStarters(words map[string]bool) ImperativeVerbOption {
	return func(rule *ImperativeVerbRule) {
		for word, val := range words {
			rule.customNonImperativeStarters[strings.ToLower(word)] = val
		}
	}
}

// WithAdditionalBaseFormsEndingWithED adds custom words that naturally end with "ed"
// but are already in imperative form.
func WithAdditionalBaseFormsEndingWithED(words map[string]bool) ImperativeVerbOption {
	return func(rule *ImperativeVerbRule) {
		for word, val := range words {
			rule.baseFormsEndingWithED[strings.ToLower(word)] = val
		}
	}
}

// WithAdditionalBaseFormsEndingWithS adds custom words that naturally end with "s"
// but are already in imperative form.
func WithAdditionalBaseFormsEndingWithS(words map[string]bool) ImperativeVerbOption {
	return func(rule *ImperativeVerbRule) {
		for word, val := range words {
			rule.baseFormsEndingWithS[strings.ToLower(word)] = val
		}
	}
}

// NewImperativeVerbRule creates a new ImperativeVerbRule.
func NewImperativeVerbRule(isConventional bool, options ...ImperativeVerbOption) *ImperativeVerbRule {
	rule := &ImperativeVerbRule{
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
		errors: make([]*domain.ValidationError, 0),
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r *ImperativeVerbRule) Name() string {
	return "ImperativeVerb"
}

// Validate validates that the first word of a commit message is in imperative form.
func (r *ImperativeVerbRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors and state
	r.errors = make([]*domain.ValidationError, 0)
	r.firstWord = ""
	r.verbType = ""

	subject := commit.Subject

	// Check for empty message
	if !r.validateNonEmptyMessage(subject) {
		return r.errors
	}

	// Check format based on conventional commit setting
	if !r.validateFormat(subject) {
		return r.errors
	}

	// Extract and validate the first word
	if !r.validateFirstWord(subject) {
		return r.errors
	}

	return r.errors
}

// validateNonEmptyMessage checks if the commit message is empty.
func (r *ImperativeVerbRule) validateNonEmptyMessage(subject string) bool {
	if subject == "" {
		r.addError(
			domain.ValidationErrorEmptyMessage,
			"empty message",
			map[string]string{
				"word": "(empty message)",
			},
		)

		return false
	}

	return true
}

// validateFormat checks if the commit message follows the expected format.
func (r *ImperativeVerbRule) validateFormat(subject string) bool {
	if r.isConventional {
		// Check for conventional commit format
		if !conventionalCommitRegex.MatchString(subject) {
			// If it doesn't match the pattern fully, figure out specific issue
			if regexp.MustCompile(`^([a-z]+)(?:\(([\w,/-]+)\))?(!)?:[ ]$`).MatchString(subject) {
				// Empty subject after colon
				r.addError(
					domain.ValidationErrorMissingSubject,
					"missing subject after type",
					map[string]string{
						"subject": subject,
						"word":    "(missing subject)",
					},
				)
			} else {
				// General format issue
				r.addError(
					domain.ValidationErrorInvalidFormat,
					"invalid conventional commit format",
					map[string]string{
						"subject": subject,
						"word":    "(invalid format)",
					},
				)
			}

			return false
		}
	} else {
		// For non-conventional commits, check if it starts with a valid character
		if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(strings.TrimSpace(subject)) {
			r.addError(
				domain.ValidationErrorInvalidFormat,
				"invalid commit format",
				map[string]string{
					"subject": subject,
					"word":    "(invalid format)",
				},
			)

			return false
		}
	}

	return true
}

// validateFirstWord extracts and validates the first word from the commit message.
func (r *ImperativeVerbRule) validateFirstWord(subject string) bool {
	// Extract first word
	word, err := r.extractFirstWord(subject)
	if err != nil {
		// Map generic errors to specific error codes
		if strings.Contains(err.Error(), "invalid conventional commit format") {
			r.addError(
				domain.ValidationErrorInvalidFormat,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(invalid format)",
				},
			)
		} else if strings.Contains(err.Error(), "missing subject after type") {
			r.addError(
				domain.ValidationErrorMissingSubject,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(missing subject)",
				},
			)
		} else if strings.Contains(err.Error(), "no valid first word found") {
			r.addError(
				domain.ValidationErrorNoFirstWord,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(no valid first word)",
				},
			)
		} else {
			// Generic error fallback
			r.addError(
				domain.ValidationErrorUnknown,
				err.Error(),
				map[string]string{
					"subject": subject,
					"word":    "(validation error)",
				},
			)
		}

		return false
	}

	// Store the first word for verbose output
	r.firstWord = word

	// Validate if the word is in imperative form
	r.validateIsImperative(word)

	return len(r.errors) == 0
}

// Result returns a concise result message.
func (r *ImperativeVerbRule) Result() string {
	if len(r.errors) > 0 {
		return "Non-imperative verb detected"
	}

	return "Commit begins with imperative verb"
}

// VerboseResult returns a detailed result message.
func (r *ImperativeVerbRule) VerboseResult() string {
	if len(r.errors) > 0 {
		// Return a more detailed error message based on error code
		//nolint:exhaustive
		switch domain.ValidationErrorCode(r.errors[0].Code) {
		case domain.ValidationErrorInvalidFormat:
			if strings.Contains(r.errors[0].Message, "conventional") {
				return "Invalid conventional commit format. Must follow 'type(scope): description' pattern."
			}

			return "Invalid commit format. Commit message should start with an imperative verb."

		case domain.ValidationErrorEmptyMessage:
			return "Commit message is empty. Cannot validate imperative verb."

		case domain.ValidationErrorMissingSubject:
			return "Missing subject after conventional commit type. Nothing to validate."

		case domain.ValidationErrorNoFirstWord:
			return "No valid first word found in commit message."

		case domain.ValidationErrorNonVerb:
			return "'" + r.firstWord + "' is a non-verb word (article, pronoun, etc.). Use an action verb instead."

		case domain.ValidationErrorPastTense:
			return "'" + r.firstWord + "' is in past tense. Use present imperative form instead (e.g., 'Add' not 'Added')."

		case domain.ValidationErrorGerund:
			return "'" + r.firstWord + "' is a gerund (-ing form). Use present imperative form instead (e.g., 'Add' not 'Adding')."

		case domain.ValidationErrorThirdPerson:
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
func (r *ImperativeVerbRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}

	// Check error code
	if len(r.errors) > 0 {
		//nolint:exhaustive
		switch domain.ValidationErrorCode(r.errors[0].Code) {
		case domain.ValidationErrorInvalidFormat:
			if strings.Contains(r.errors[0].Message, "conventional") {
				return `Format your commit message according to the Conventional Commits specification.
Example: feat(auth): Add login feature
The correct format is: type(scope): subject
- type: feat, fix, docs, etc.
- scope: optional context (in parentheses)
- subject: description of the change`
			}

			return "Make sure your commit message starts with an imperative verb (e.g., 'Add', 'Fix', 'Update')."

		case domain.ValidationErrorEmptyMessage:
			return "Provide a non-empty commit message with a verb in the imperative mood."

		case domain.ValidationErrorMissingSubject:
			return "Add a description after the type and colon in your conventional commit message."

		case domain.ValidationErrorNoFirstWord:
			return "Start your commit message with a word (letters or numbers). Remove any leading special characters."

		case domain.ValidationErrorNonVerb:
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

		case domain.ValidationErrorPastTense:
			return `Avoid using past tense verbs at the start of commit messages.
Instead of "Added feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`

		case domain.ValidationErrorGerund:
			return `Avoid using gerund (-ing) forms at the start of commit messages.
Instead of "Adding feature", use "Add feature".
Use the imperative mood that completes the sentence:
"If applied, this commit will [your commit message]"`

		case domain.ValidationErrorThirdPerson:
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
func (r *ImperativeVerbRule) Errors() []*domain.ValidationError {
	return r.errors
}

// addError adds a structured validation error using standard error codes.
func (r *ImperativeVerbRule) addError(code domain.ValidationErrorCode, message string, context map[string]string) {
	err := domain.NewStandardValidationError(r.Name(), code, message)

	// Add any context values
	for key, value := range context {
		err = err.WithContext(key, value)
	}

	// Store verb type for verbose output
	if context != nil {
		if wordType, ok := context["type"]; ok {
			r.verbType = wordType
		}

		if word, ok := context["word"]; ok {
			r.firstWord = word
		} else if code == domain.ValidationErrorInvalidFormat {
			r.firstWord = "(invalid format)"
		} else if code == domain.ValidationErrorMissingSubject {
			r.firstWord = "(missing subject)"
		} else if code == domain.ValidationErrorEmptyMessage {
			r.firstWord = "(empty message)"
		} else if code == domain.ValidationErrorNoFirstWord {
			r.firstWord = "(no valid first word)"
		}
	}

	r.errors = append(r.errors, err)
}

// extractFirstWord extracts the first word from the commit message.
func (r *ImperativeVerbRule) extractFirstWord(subject string) (string, error) {
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
func (r *ImperativeVerbRule) validateIsImperative(word string) {
	// Store the original word directly - this ensures we show the user exactly what they typed
	r.firstWord = word

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
		r.addError(
			domain.ValidationErrorNonVerb,
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
		r.validateWithSimpleRules(wordLower, word)

		return
	}

	// Check for specific non-imperative forms
	// Past tense verbs often end in "ed" and their stem is different
	if strings.HasSuffix(wordLower, "ed") && stem != wordLower && !r.isBaseFormWithEDEnding(wordLower) {
		r.addError(
			domain.ValidationErrorPastTense,
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
		r.addError(
			domain.ValidationErrorGerund,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be a gerund",
			map[string]string{
				"word": word,
				"type": "gerund",
			},
		)

		return
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !r.isBaseFormWithSEnding(wordLower) {
		r.addError(
			domain.ValidationErrorThirdPerson,
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
func (r *ImperativeVerbRule) validateWithSimpleRules(wordLower, originalWord string) {
	// Ensure the original word is stored
	r.firstWord = originalWord

	// Simple pattern checks for non-imperative forms
	if strings.HasSuffix(wordLower, "ed") && !r.isBaseFormWithEDEnding(wordLower) {
		r.addError(
			domain.ValidationErrorPastTense,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be past tense",
			map[string]string{
				"word": originalWord,
				"type": "past_tense",
			},
		)

		return
	}

	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		r.addError(
			domain.ValidationErrorGerund,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be a gerund",
			map[string]string{
				"word": originalWord,
				"type": "gerund",
			},
		)

		return
	}

	if strings.HasSuffix(wordLower, "s") && !r.isBaseFormWithSEnding(wordLower) && len(wordLower) > 2 {
		r.addError(
			domain.ValidationErrorThirdPerson,
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
func (r *ImperativeVerbRule) isBaseFormWithEDEnding(word string) bool {
	return r.baseFormsEndingWithED[word]
}

// isBaseFormWithSEnding checks if a word ending in "s" is actually a base form.
func (r *ImperativeVerbRule) isBaseFormWithSEnding(word string) bool {
	return r.baseFormsEndingWithS[word]
}
