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
	BaseRule                    BaseRule
	isConventional              bool
	customNonImperativeStarters map[string]bool
	baseFormsEndingWithED       map[string]bool
	baseFormsEndingWithS        map[string]bool
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
		BaseRule:                    NewBaseRule("ImperativeVerb"),
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
	}
	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// NewImperativeVerbRuleWithConfig creates an ImperativeVerbRule using configuration.
func NewImperativeVerbRuleWithConfig(config domain.SubjectConfigProvider, conventionalConfig domain.ConventionalConfigProvider) ImperativeVerbRule {
	// Build options based on the configuration
	var options []ImperativeVerbOption

	// Check if imperative is required
	isConventional := conventionalConfig.ConventionalRequired()

	// Check if we need to use conventional commit format
	if isConventional {
		options = append(options, WithImperativeConventionalCommit(true))
	}

	return NewImperativeVerbRule(config.SubjectRequireImperative(), options...)
}

// Name returns the rule name.
func (r ImperativeVerbRule) Name() string {
	return r.BaseRule.Name()
}

// validateImperativeWithState validates a commit and returns both errors and an updated rule state.
func validateImperativeWithState(rule ImperativeVerbRule, commit domain.CommitInfo) ([]appErrors.ValidationError, ImperativeVerbRule) {
	// Start with a clean slate and mark as run
	updatedRule := rule
	updatedRule.BaseRule = updatedRule.BaseRule.WithClearedErrors().WithRun()
	updatedRule.firstWord = ""
	updatedRule.verbType = ""

	subject := commit.Subject

	// Check for empty message
	if subject == "" {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrMissingSubject,
			"Commit message is empty",
			appErrors.WithContextMap(map[string]string{
				"word": "(empty message)",
			}),
		)
		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.firstWord = "(empty message)"

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Check format
	if !validateFormat(&updatedRule, subject) {
		return updatedRule.BaseRule.Errors(), updatedRule
	}

	// Extract and validate first word
	return validateFirstWord(updatedRule, subject)
}

// Validate performs validation against a commit and returns any errors.
// This uses value semantics and does not modify the rule's state.
func (r ImperativeVerbRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	errors, _ := validateImperativeWithState(r, commit)

	return errors
}

// ValidateImperativeWithState is the exported version of validateImperativeWithState.
// This is needed for testing but follows the same pure function approach.
func ValidateImperativeWithState(rule ImperativeVerbRule, commit domain.CommitInfo) ([]appErrors.ValidationError, ImperativeVerbRule) {
	return validateImperativeWithState(rule, commit)
}

// Result returns a concise result message.
func (r ImperativeVerbRule) Result() string {
	if r.BaseRule.HasErrors() {
		return "Non-imperative verb detected"
	}

	return "Commit begins with imperative verb"
}

// VerboseResult returns a detailed result message.
func (r ImperativeVerbRule) VerboseResult() string {
	if r.BaseRule.HasErrors() {
		errors := r.BaseRule.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		validationErr := errors[0]
		// Return a more detailed error message based on error code
		switch appErrors.ValidationErrorCode(validationErr.Code) { //nolint:exhaustive
		case appErrors.ErrInvalidFormat:
			if strings.Contains(validationErr.Message, "conventional") {
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
				return "Word '" + r.firstWord + "' detected: " + validationErr.Error()
			}

			return validationErr.Error()
		}
	}

	return "Commit begins with proper imperative verb '" + r.firstWord + "'"
}

// Help returns guidance on how to fix rule violations.
func (r ImperativeVerbRule) Help() string {
	if !r.BaseRule.HasErrors() {
		return "No errors to fix. This rule checks that commit messages begin with an imperative verb (e.g., Add, Fix, Update) rather than descriptive forms (e.g., Adds, Fixed, Updated)."
	}
	// Check error code
	errors := r.BaseRule.Errors()
	if len(errors) > 0 {
		switch appErrors.ValidationErrorCode(errors[0].Code) { //nolint:exhaustive
		case appErrors.ErrInvalidFormat:
			if strings.Contains(errors[0].Message, "conventional") {
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
	errMsg := errors[0].Message
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
	return r.BaseRule.Errors()
}

// HasErrors checks if there are any validation errors.
func (r ImperativeVerbRule) HasErrors() bool {
	return r.BaseRule.HasErrors()
}

// validateFormat checks if the commit message follows the expected format.
func validateFormat(rule *ImperativeVerbRule, subject string) bool {
	if rule.isConventional {
		// Check for conventional commit format
		if !conventionalCommitRegex.MatchString(subject) {
			// If it doesn't match the pattern fully, figure out specific issue
			if regexp.MustCompile(`^([a-z]+)(?:\(([\w,/-]+)\))?(!)?:[ ]$`).MatchString(subject) {
				// Empty subject after colon
				err := appErrors.New(
					rule.Name(),
					appErrors.ErrMissingSubject,
					"Missing subject after type in conventional commit",
					appErrors.WithContextMap(map[string]string{
						"subject": subject,
						"word":    "(missing subject)",
					}),
				)

				rule.BaseRule = rule.BaseRule.WithError(err)
				rule.firstWord = "(missing subject)"

				return false
			}
			// General format issue
			err := appErrors.New(
				rule.Name(),
				appErrors.ErrInvalidFormat,
				"Invalid conventional commit format",
				appErrors.WithContextMap(map[string]string{
					"subject": subject,
					"word":    "(invalid format)",
				}),
			)

			rule.BaseRule = rule.BaseRule.WithError(err)
			rule.firstWord = "(invalid format)"

			return false
		}
	}
	// For non-conventional commits, check if it starts with a valid character
	if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(strings.TrimSpace(subject)) {
		err := appErrors.New(
			rule.Name(),
			appErrors.ErrInvalidFormat,
			"Invalid commit message format",
			appErrors.WithContextMap(map[string]string{
				"subject": subject,
				"word":    "(invalid format)",
			}),
		)

		rule.BaseRule = rule.BaseRule.WithError(err)
		rule.firstWord = "(invalid format)"

		return false
	}

	return true
}

// validateFirstWord extracts and validates the first word from the commit message.
func validateFirstWord(rule ImperativeVerbRule, subject string) ([]appErrors.ValidationError, ImperativeVerbRule) {
	updatedRule := rule

	// Extract first word
	word, err := extractFirstWord(rule, subject)
	if err != nil {
		// Map generic errors to specific error codes
		var errCode appErrors.ValidationErrorCode

		context := map[string]string{"subject": subject}

		if strings.Contains(err.Error(), "invalid conventional commit format") {
			errCode = appErrors.ErrInvalidFormat
			context["word"] = "(invalid format)"
			updatedRule.firstWord = "(invalid format)"
		} else if strings.Contains(err.Error(), "missing subject after type") {
			errCode = appErrors.ErrMissingSubject
			context["word"] = "(missing subject)"
			updatedRule.firstWord = "(missing subject)"
		} else if strings.Contains(err.Error(), "no valid first word found") {
			errCode = appErrors.ErrNoFirstWord
			context["word"] = "(no valid first word)"
			updatedRule.firstWord = "(no valid first word)"
		} else {
			errCode = appErrors.ErrUnknown
			context["word"] = "(validation error)"
			updatedRule.firstWord = "(validation error)"
		}

		validationErr := appErrors.New(
			updatedRule.Name(),
			errCode,
			err.Error(),
			appErrors.WithContextMap(context),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(validationErr)

		return []appErrors.ValidationError{validationErr}, updatedRule
	}

	// Store the first word for verbose output
	updatedRule.firstWord = word

	// Validate if the word is in imperative form
	return validateIsImperative(updatedRule, word)
}

// validateIsImperative checks if a word is in imperative form using snowball stemming.
func validateIsImperative(rule ImperativeVerbRule, word string) ([]appErrors.ValidationError, ImperativeVerbRule) {
	updatedRule := rule
	wordLower := strings.ToLower(word)

	// Default non-imperative starters
	nonImperativeStarters := map[string]bool{
		"i": true, "we": true, "they": true, "he": true, "she": true, "it": true,
		"the": true, "a": true, "an": true, "this": true, "that": true,
		"these": true, "those": true, "my": true, "your": true, "our": true,
	}

	// Add custom non-imperative starters if provided
	for word, val := range rule.customNonImperativeStarters {
		nonImperativeStarters[strings.ToLower(word)] = val
	}

	if nonImperativeStarters[wordLower] {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrNonVerb,
			"first word of commit must be an imperative verb: \""+word+"\" is not a verb",
			appErrors.WithContextMap(map[string]string{
				"word": word,
				"type": "non_verb",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "non_verb"

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Use snowball stemmer to get the base form
	stem, err := snowball.Stem(wordLower, "english", true)
	if err != nil {
		// If stemming fails, fall back to simpler checks
		return validateWithSimpleRules(updatedRule, wordLower, word)
	}

	// Check for specific non-imperative forms
	// Past tense verbs often end in "ed" and their stem is different
	if strings.HasSuffix(wordLower, "ed") && stem != wordLower && !isBaseFormWithEDEnding(updatedRule, wordLower) {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrPastTense,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be past tense",
			appErrors.WithContextMap(map[string]string{
				"word": word,
				"type": "past_tense",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "past_tense"

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Gerunds end in "ing"
	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrGerund,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be a gerund",
			appErrors.WithContextMap(map[string]string{
				"word": word,
				"type": "gerund",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "gerund"

		return []appErrors.ValidationError{err}, updatedRule
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !isBaseFormWithSEnding(updatedRule, wordLower) {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrThirdPerson,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be 3rd person present",
			appErrors.WithContextMap(map[string]string{
				"word": word,
				"type": "third_person",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "third_person"

		return []appErrors.ValidationError{err}, updatedRule
	}

	return []appErrors.ValidationError{}, updatedRule
}

// validateWithSimpleRules provides a fallback if stemming fails.
func validateWithSimpleRules(rule ImperativeVerbRule, wordLower, originalWord string) ([]appErrors.ValidationError, ImperativeVerbRule) {
	updatedRule := rule

	// Simple pattern checks for non-imperative forms
	if strings.HasSuffix(wordLower, "ed") && !isBaseFormWithEDEnding(updatedRule, wordLower) {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrPastTense,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be past tense",
			appErrors.WithContextMap(map[string]string{
				"word": originalWord,
				"type": "past_tense",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "past_tense"

		return []appErrors.ValidationError{err}, updatedRule
	}

	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrGerund,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be a gerund",
			appErrors.WithContextMap(map[string]string{
				"word": originalWord,
				"type": "gerund",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "gerund"

		return []appErrors.ValidationError{err}, updatedRule
	}

	if strings.HasSuffix(wordLower, "s") && !isBaseFormWithSEnding(updatedRule, wordLower) && len(wordLower) > 2 {
		err := appErrors.New(
			updatedRule.Name(),
			appErrors.ErrThirdPerson,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be 3rd person present",
			appErrors.WithContextMap(map[string]string{
				"word": originalWord,
				"type": "third_person",
			}),
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "third_person"

		return []appErrors.ValidationError{err}, updatedRule
	}

	return []appErrors.ValidationError{}, updatedRule
}

// isBaseFormWithEDEnding checks if a word ending in "ed" is actually a base form.
func isBaseFormWithEDEnding(rule ImperativeVerbRule, word string) bool {
	return rule.baseFormsEndingWithED[word]
}

// isBaseFormWithSEnding checks if a word ending in "s" is actually a base form.
func isBaseFormWithSEnding(rule ImperativeVerbRule, word string) bool {
	return rule.baseFormsEndingWithS[word]
}

// extractFirstWord extracts the first word from the commit message.
func extractFirstWord(rule ImperativeVerbRule, subject string) (string, error) {
	if rule.isConventional {
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
