// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
// Package rules provides validation rules for Git commits.
package rules

import (
	"context"
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
func NewImperativeVerbRule(isImperativeRequired bool, options ...ImperativeVerbOption) ImperativeVerbRule {
	rule := ImperativeVerbRule{
		BaseRule:                    NewBaseRuleWithRequired("ImperativeVerb", isImperativeRequired),
		isConventional:              false, // Default to non-conventional
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

// NewImperativeVerbRuleWithConfig creates an ImperativeVerbRule using unified configuration.

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

	// If imperative verbs are not required, return no errors
	if !updatedRule.BaseRule.IsRequired {
		return []appErrors.ValidationError{}, updatedRule
	}

	subject := commit.Subject

	// Check for empty message
	if subject == "" {
		// Create error context with rich information
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		// Create a rich error
		helpMessage := `Your commit message is empty, which doesn't provide any information about the changes.

A good commit message should:
1. Begin with an imperative verb (e.g., "Add", "Fix", "Update")
2. Clearly describe what the commit does
3. Be concise but informative

Examples of good commit messages:
- Add user authentication feature
- Fix bug in payment processing
- Update documentation with examples
- Refactor database connection logic

Please write a descriptive commit message to help others understand your changes.`

		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrMissingSubject,
			"Commit message is empty",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("word", "(empty message)")

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
func (r ImperativeVerbRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	errors, _ := validateImperativeWithState(r, commit)

	return errors
}

// ValidateImperativeWithState is the exported version of validateImperativeWithState.
// This is needed for testing but follows the same pure function approach.
func ValidateImperativeWithState(rule ImperativeVerbRule, commit domain.CommitInfo) ([]appErrors.ValidationError, ImperativeVerbRule) {
	return validateImperativeWithState(rule, commit)
}

// Result returns a concise result message.
func (r ImperativeVerbRule) Result(_ []appErrors.ValidationError) string {
	if r.BaseRule.HasErrors() {
		return "Non-imperative verb detected"
	}

	return "Commit begins with imperative verb"
}

// VerboseResult returns a detailed result message.
func (r ImperativeVerbRule) VerboseResult(_ []appErrors.ValidationError) string {
	if r.BaseRule.HasErrors() {
		errors := r.BaseRule.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		validationErr := errors[0]
		// Return a more detailed error message based on error code
		// Use if statements instead of switch to avoid exhaustive linter complaints
		code := appErrors.ValidationErrorCode(validationErr.Code)

		if code == appErrors.ErrInvalidFormat {
			if strings.Contains(validationErr.Message, "conventional") {
				return "Invalid conventional commit format. Must follow 'type(scope): description' pattern."
			}

			// Return for invalid format case
			return "Invalid commit format. Commit message should start with an imperative verb."
		}

		if code == appErrors.ErrEmptyMessage {
			return "Commit message is empty. Cannot validate imperative verb."
		}

		if code == appErrors.ErrMissingSubject {
			return "Missing subject after conventional commit type. Nothing to validate."
		}

		if code == appErrors.ErrNoFirstWord {
			return "No valid first word found in commit message."
		}

		if code == appErrors.ErrNonVerb {
			return "'" + r.firstWord + "' is a non-verb word (article, pronoun, etc.). Use an action verb instead."
		}

		if code == appErrors.ErrPastTense {
			return "'" + r.firstWord + "' is in past tense. Use present imperative form instead (e.g., 'Add' not 'Added')."
		}

		if code == appErrors.ErrGerund {
			return "'" + r.firstWord + "' is a gerund (-ing form). Use present imperative form instead (e.g., 'Add' not 'Adding')."
		}

		if code == appErrors.ErrThirdPerson {
			return "'" + r.firstWord + "' is in third person form. Use present imperative form instead (e.g., 'Add' not 'Adds')."
		}

		// Default case
		// If we have a first word but no specific code match, include it in the error
		if r.firstWord != "" {
			return "Word '" + r.firstWord + "' detected: " + validationErr.Error()
		}

		return validationErr.Error()
	}

	return "Commit begins with proper imperative verb '" + r.firstWord + "'"
}

// Help returns guidance on how to fix rule violations.
func (r ImperativeVerbRule) Help(errors []appErrors.ValidationError) string {
	if !r.BaseRule.HasErrors() {
		return "No errors to fix. This rule checks that commit messages begin with an imperative verb (e.g., Add, Fix, Update) rather than descriptive forms (e.g., Adds, Fixed, Updated)."
	}

	if len(errors) > 0 {
		// Use if statements instead of switch to avoid exhaustive linter complaints
		code := appErrors.ValidationErrorCode(errors[0].Code)

		if code == appErrors.ErrInvalidFormat {
			if strings.Contains(errors[0].Message, "conventional") {
				return `Format your commit message according to the Conventional Commits specification.
Example: feat(auth): Add login feature
The correct format is: type(scope): subject
- type: feat, fix, docs, etc.
- scope: optional context (in parentheses)
- subject: description of the change`
			}

			return "Make sure your commit message starts with an imperative verb (e.g., 'Add', 'Fix', 'Update')."
		}

		if code == appErrors.ErrEmptyMessage {
			return "Provide a non-empty commit message with a verb in the imperative mood."
		}

		if code == appErrors.ErrMissingSubject {
			return "Add a description after the type and colon in your conventional commit message."
		}

		if code == appErrors.ErrNoFirstWord {
			return "Start your commit message with a word (letters or numbers). Remove any leading special characters."
		}

		if code == appErrors.ErrNonVerb {
			return `Non-Verb Error: The word "${r.firstWord}" is not a verb.

Git commit messages should start with an imperative verb. You used a non-verb word
like an article, pronoun, preposition, or noun instead of an action verb.

✅ CORRECT commit messages (start with verbs):
- Add user authentication feature
- Fix login timeout issue
- Update documentation with examples
- Remove deprecated API endpoints
- Refactor database connection logic

❌ INCORRECT commit messages (start with non-verbs):
- The new login feature              (starts with article)
- This commit fixes the bug          (starts with pronoun)
- User authentication added          (starts with noun)
- Documentation for API              (starts with noun)
- Our implementation of feature      (starts with pronoun)

WHY IMPERATIVE VERBS MATTER:
Git itself uses imperative form in auto-generated commit messages like "Merge branch..."
This creates a consistent, actionable commit history. Imperative verbs complete this sentence:
"If applied, this commit will [your commit message]"

COMMON IMPERATIVE VERBS TO USE:
Add       Fix       Update    Remove    Change
Refactor  Implement Optimize  Improve   Merge
Create    Delete    Revert    Extract   Move
Rename    Simplify  Adjust    Configure Test`
		}

		if code == appErrors.ErrPastTense {
			return `Past Tense Error: The word "${r.firstWord}" is in past tense.

Git commit messages should use the imperative mood (present tense), not past tense.
You used a past tense verb like "Added", "Updated", or "Fixed".

✅ CORRECT (imperative mood):       ❌ INCORRECT (past tense):
- Add feature                       - Added feature
- Fix bug                           - Fixed bug
- Update documentation              - Updated documentation
- Implement authentication          - Implemented authentication
- Remove deprecated code            - Removed deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your past tense verb to present imperative form:
- "Added" → "Add"
- "Fixed" → "Fix"
- "Updated" → "Update"
- "Implemented" → "Implement"
- "Removed" → "Remove"

Simply drop the "-ed" ending in most cases.`
		}

		if code == appErrors.ErrGerund {
			return `Gerund Error: The word "${r.firstWord}" is a gerund (verb ending in "-ing").

Git commit messages should use the imperative mood, not continuous/gerund forms.
You used a gerund verb form like "Adding", "Fixing", or "Updating".

✅ CORRECT (imperative mood):       ❌ INCORRECT (gerund form):
- Add feature                       - Adding feature
- Fix bug                           - Fixing bug
- Update documentation              - Updating documentation
- Implement authentication          - Implementing authentication
- Remove deprecated code            - Removing deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your gerund verb to present imperative form:
- "Adding" → "Add"
- "Fixing" → "Fix"
- "Updating" → "Update"
- "Implementing" → "Implement"
- "Removing" → "Remove"

Simply drop the "-ing" ending and use the base verb form.`
		}

		if code == appErrors.ErrThirdPerson {
			return `Third Person Error: The word "${r.firstWord}" is in third person form.

Git commit messages should use the imperative mood, not third-person present tense.
You used a third-person verb form like "Adds", "Fixes", or "Updates".

✅ CORRECT (imperative mood):       ❌ INCORRECT (third person):
- Add feature                       - Adds feature
- Fix bug                           - Fixes bug
- Update documentation              - Updates documentation
- Implement authentication          - Implements authentication
- Remove deprecated code            - Removes deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your third-person verb to present imperative form:
- "Adds" → "Add"
- "Fixes" → "Fix"
- "Updates" → "Update"
- "Implements" → "Implement"
- "Removes" → "Remove"

Simply drop the "-s" ending to get the imperative form.`
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
				// Create error context
				ctx := appErrors.NewContext()
				// We don't have the full commit here, but we can add the subject

				// Empty subject after colon
				helpMessage := `Your conventional commit is missing a description after the type and colon.

A conventional commit must include a description that explains what changes were made.

Correct format:
<type>[optional scope][optional !]: <description>

Examples:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions

Please add a clear description after the colon.`

				err := appErrors.CreateRichError(
					rule.Name(),
					appErrors.ErrMissingSubject,
					"Missing subject after type in conventional commit",
					helpMessage,
					ctx,
				)

				// Add context details
				err = err.WithContext("subject", subject)
				err = err.WithContext("word", "(missing subject)")

				rule.BaseRule = rule.BaseRule.WithError(err)
				rule.firstWord = "(missing subject)"

				return false
			}
			// Create error context
			ctx := appErrors.NewContext()
			// We don't have the full commit here, but we can add the subject

			// General format issue
			helpMessage := `Your commit message doesn't follow the conventional format.

The correct format is:
<type>[optional scope][optional !]: <description>

Examples:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions

Common types:
- feat: a new feature
- fix: a bug fix
- docs: documentation changes
- style: formatting changes
- refactor: code change that neither adds a feature nor fixes a bug
- test: adding or correcting tests

Please format your commit message according to the conventional commit specification.`

			err := appErrors.CreateRichError(
				rule.Name(),
				appErrors.ErrInvalidFormat,
				"Invalid conventional commit format",
				helpMessage,
				ctx,
			)

			// Add context details
			err = err.WithContext("subject", subject)
			err = err.WithContext("word", "(invalid format)")

			rule.BaseRule = rule.BaseRule.WithError(err)
			rule.firstWord = "(invalid format)"

			return false
		}
	}
	// For non-conventional commits, check if it starts with a valid character
	if !regexp.MustCompile(`^[a-zA-Z0-9]`).MatchString(strings.TrimSpace(subject)) {
		// Create error context
		ctx := appErrors.NewContext()
		// We don't have the full commit here, but we can add the subject

		helpMessage := `Your commit message doesn't start with a letter or number.

A commit message should start with an imperative verb like:
- Add
- Fix
- Update
- Remove
- Refactor

Examples of good commit messages:
- Add user authentication feature
- Fix bug in payment processing
- Update documentation with examples

Please revise your commit message to start with an imperative verb.`

		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidFormat,
			"Invalid commit message format",
			helpMessage,
			ctx,
		)

		// Add context details
		err = err.WithContext("subject", subject)
		err = err.WithContext("word", "(invalid format)")

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

		var helpMessage string

		// Create error context with commit information
		ctx := appErrors.NewContext()

		if strings.Contains(err.Error(), "invalid conventional commit format") {
			errCode = appErrors.ErrInvalidFormat
			updatedRule.firstWord = "(invalid format)"
			helpMessage = `Your commit message doesn't follow the conventional format.

The correct format is:
<type>[optional scope][optional !]: <description>

Examples:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions

Please format your commit message according to the conventional commit specification.`
		} else if strings.Contains(err.Error(), "missing subject after type") {
			errCode = appErrors.ErrMissingSubject
			updatedRule.firstWord = "(missing subject)"
			helpMessage = `Your conventional commit is missing a description after the type and colon.

A conventional commit must include a description that explains what changes were made.

Correct format:
<type>[optional scope][optional !]: <description>

Examples:
- feat: add user authentication
- fix(auth): resolve login timeout issue
- docs: update installation instructions

Please add a clear description after the colon.`
		} else if strings.Contains(err.Error(), "no valid first word found") {
			errCode = appErrors.ErrNoFirstWord
			updatedRule.firstWord = "(no valid first word)"
			helpMessage = `Your commit message doesn't start with a valid word.

A commit message should start with a letter or number, preferably an imperative verb.

Examples of good commit messages:
- Add user authentication feature
- Fix bug in payment processing
- Update documentation with examples

Please start your commit message with an imperative verb.`
		} else {
			errCode = appErrors.ErrUnknown
			updatedRule.firstWord = "(validation error)"
			helpMessage = `There was an error validating your commit message.

A commit message should:
1. Begin with an imperative verb (e.g., "Add", "Fix", "Update")
2. Follow conventional commit format if required
3. Be concise but descriptive

Please check your commit message format and try again.`
		}

		// Create a rich error with detailed help
		validationErr := appErrors.CreateRichError(
			updatedRule.Name(),
			errCode,
			err.Error(),
			helpMessage,
			ctx,
		)

		// Add additional context
		validationErr = validationErr.WithContext("subject", subject)
		validationErr = validationErr.WithContext("word", updatedRule.firstWord)

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
		// Create error context with rich information
		ctx := appErrors.NewContext()

		// Detailed help message
		helpMessage := `Non-Verb Error: The word "${word}" is not a verb.

Git commit messages should start with an imperative verb. You used a non-verb word
like an article, pronoun, preposition, or noun instead of an action verb.

✅ CORRECT commit messages (start with verbs):
- Add user authentication feature
- Fix login timeout issue
- Update documentation with examples
- Remove deprecated API endpoints
- Refactor database connection logic

❌ INCORRECT commit messages (start with non-verbs):
- The new login feature              (starts with article)
- This commit fixes the bug          (starts with pronoun)
- User authentication added          (starts with noun)
- Documentation for API              (starts with noun)
- Our implementation of feature      (starts with pronoun)

WHY IMPERATIVE VERBS MATTER:
Git itself uses imperative form in auto-generated commit messages.
This creates a consistent, actionable commit history.`

		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrNonVerb,
			"first word of commit must be an imperative verb: \""+word+"\" is not a verb",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("word", word)
		err = err.WithContext("type", "non_verb")

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
		// Create error context with rich information
		ctx := appErrors.NewContext()

		// Detailed help message for past tense
		helpMessage := `Past Tense Error: The word "${word}" is in past tense.

Git commit messages should use the imperative mood (present tense), not past tense.
You used a past tense verb like "Added", "Updated", or "Fixed".

✅ CORRECT (imperative mood):       ❌ INCORRECT (past tense):
- Add feature                       - Added feature
- Fix bug                           - Fixed bug
- Update documentation              - Updated documentation
- Implement authentication          - Implemented authentication
- Remove deprecated code            - Removed deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your past tense verb to present imperative form:
- "${word}" → "${stem}" or another present imperative form`

		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrPastTense,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be past tense",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("word", word)
		err = err.WithContext("type", "past_tense")
		err = err.WithContext("suggested_form", stem)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "past_tense"

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Gerunds end in "ing"
	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		// Create error context with rich information
		ctx := appErrors.NewContext()

		// Suggested base form (strip "ing" and optionally add "e" if needed)
		suggestedForm := strings.TrimSuffix(wordLower, "ing")
		// Handle special cases like "writing" -> "write" (not "writ")
		if len(suggestedForm) > 2 && !strings.HasSuffix(suggestedForm, "e") &&
			(suggestedForm[len(suggestedForm)-1] == 't' ||
				suggestedForm[len(suggestedForm)-1] == 'v' ||
				suggestedForm[len(suggestedForm)-1] == 'd') {
			suggestedForm += "e"
		}

		// Detailed help message for gerund
		helpMessage := `Gerund Error: The word "${word}" is a gerund (verb ending in "-ing").

Git commit messages should use the imperative mood, not continuous/gerund forms.
You used a gerund verb form like "Adding", "Fixing", or "Updating".

✅ CORRECT (imperative mood):       ❌ INCORRECT (gerund form):
- Add feature                       - Adding feature
- Fix bug                           - Fixing bug
- Update documentation              - Updating documentation
- Implement authentication          - Implementing authentication
- Remove deprecated code            - Removing deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your gerund verb to present imperative form:
- "${word}" → "${suggestedForm}" or another present imperative form

Simply drop the "-ing" ending and use the base verb form.`

		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrGerund,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be a gerund",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("word", word)
		err = err.WithContext("type", "gerund")
		err = err.WithContext("suggested_form", suggestedForm)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "gerund"

		return []appErrors.ValidationError{err}, updatedRule
	}

	// 3rd person singular typically ends in "s" and stem is different
	if strings.HasSuffix(wordLower, "s") && stem != wordLower && !isBaseFormWithSEnding(updatedRule, wordLower) {
		// Create error context with rich information
		ctx := appErrors.NewContext()

		// Detailed help message for third person
		helpMessage := `Third Person Error: The word "${word}" is in third person form.

Git commit messages should use the imperative mood, not third-person present tense.
You used a third-person verb form like "Adds", "Fixes", or "Updates".

✅ CORRECT (imperative mood):       ❌ INCORRECT (third person):
- Add feature                       - Adds feature
- Fix bug                           - Fixes bug
- Update documentation              - Updates documentation
- Implement authentication          - Implements authentication
- Remove deprecated code            - Removes deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your third-person verb to present imperative form:
- "${word}" → "${stem}" or another present imperative form

Simply drop the "-s" ending to get the imperative form.`

		err := appErrors.CreateRichError(
			updatedRule.Name(),
			appErrors.ErrThirdPerson,
			"first word of commit must be an imperative verb: \""+word+"\" appears to be 3rd person present",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("word", word)
		err = err.WithContext("type", "third_person")
		err = err.WithContext("suggested_form", stem)

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
		// Create suggested form by removing "ed"
		suggestedForm := strings.TrimSuffix(wordLower, "ed")
		// Handle special cases like "used" -> "use" (not "us")
		if len(suggestedForm) > 2 && !strings.HasSuffix(suggestedForm, "e") &&
			(suggestedForm[len(suggestedForm)-1] == 's' ||
				suggestedForm[len(suggestedForm)-1] == 'v' ||
				suggestedForm[len(suggestedForm)-1] == 'c') {
			suggestedForm += "e"
		}

		// Detailed help message for past tense
		helpMessage := `Past Tense Error: The word "${word}" is in past tense.

Git commit messages should use the imperative mood (present tense), not past tense.
You used a past tense verb like "Added", "Updated", or "Fixed".

✅ CORRECT (imperative mood):       ❌ INCORRECT (past tense):
- Add feature                       - Added feature
- Fix bug                           - Fixed bug
- Update documentation              - Updated documentation
- Implement authentication          - Implemented authentication
- Remove deprecated code            - Removed deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your past tense verb to present imperative form:
- "${word}" → "${suggestedForm}" or another present imperative form

Simply drop the "-ed" ending in most cases.`

		err := appErrors.ImperativeError(
			updatedRule.Name(),
			appErrors.ErrPastTense,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be past tense",
			helpMessage,
			originalWord,
			"past_tense",
			suggestedForm,
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "past_tense"

		return []appErrors.ValidationError{err}, updatedRule
	}

	if strings.HasSuffix(wordLower, "ing") && len(wordLower) > 4 {
		// Suggested base form (strip "ing" and optionally add "e" if needed)
		suggestedForm := strings.TrimSuffix(wordLower, "ing")
		// Handle special cases like "writing" -> "write" (not "writ")
		if len(suggestedForm) > 2 && !strings.HasSuffix(suggestedForm, "e") &&
			(suggestedForm[len(suggestedForm)-1] == 't' ||
				suggestedForm[len(suggestedForm)-1] == 'v' ||
				suggestedForm[len(suggestedForm)-1] == 'd') {
			suggestedForm += "e"
		}

		// Detailed help message for gerund
		helpMessage := `Gerund Error: The word "${word}" is a gerund (verb ending in "-ing").

Git commit messages should use the imperative mood, not continuous/gerund forms.
You used a gerund verb form like "Adding", "Fixing", or "Updating".

✅ CORRECT (imperative mood):       ❌ INCORRECT (gerund form):
- Add feature                       - Adding feature
- Fix bug                           - Fixing bug
- Update documentation              - Updating documentation
- Implement authentication          - Implementing authentication
- Remove deprecated code            - Removing deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your gerund verb to present imperative form:
- "${word}" → "${suggestedForm}" or another present imperative form

Simply drop the "-ing" ending and use the base verb form.`

		err := appErrors.ImperativeError(
			updatedRule.Name(),
			appErrors.ErrGerund,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be a gerund",
			helpMessage,
			originalWord,
			"gerund",
			suggestedForm,
		)

		updatedRule.BaseRule = updatedRule.BaseRule.WithError(err)
		updatedRule.verbType = "gerund"

		return []appErrors.ValidationError{err}, updatedRule
	}

	if strings.HasSuffix(wordLower, "s") && !isBaseFormWithSEnding(updatedRule, wordLower) && len(wordLower) > 2 {
		// Create suggested form by removing "s"
		suggestedForm := strings.TrimSuffix(wordLower, "s")

		// Detailed help message for third person
		helpMessage := `Third Person Error: The word "${word}" is in third person form.

Git commit messages should use the imperative mood, not third-person present tense.
You used a third-person verb form like "Adds", "Fixes", or "Updates".

✅ CORRECT (imperative mood):       ❌ INCORRECT (third person):
- Add feature                       - Adds feature
- Fix bug                           - Fixes bug
- Update documentation              - Updates documentation
- Implement authentication          - Implements authentication
- Remove deprecated code            - Removes deprecated code

WHY THIS MATTERS:
Using imperative mood creates consistency across the commit history and
follows Git's own conventions. It completes the sentence:
"If applied, this commit will [your commit message]"

HOW TO FIX:
Convert your third-person verb to present imperative form:
- "${word}" → "${suggestedForm}" or another present imperative form

Simply drop the "-s" ending to get the imperative form.`

		err := appErrors.ImperativeError(
			updatedRule.Name(),
			appErrors.ErrThirdPerson,
			"first word of commit must be an imperative verb: \""+originalWord+"\" appears to be 3rd person present",
			helpMessage,
			originalWord,
			"third_person",
			suggestedForm,
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
