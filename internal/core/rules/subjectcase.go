// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// subjectCaseFirstWordRegex is the regular expression used to find the first word in a commit.
var subjectCaseFirstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

// Format: type(scope)!: description.
var subjectCaseRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// SubjectCaseRule enforces the case of the first word in the subject.
// This rule helps ensure commit messages follow a consistent style by validating
// the first letter's case based on the project's convention.
//
// For conventional commits (e.g., "feat(scope): add feature"), it checks the
// capitalization of the first word after the "type(scope): " prefix.
//
// For non-conventional commits, it simply checks the first word of the subject.
type SubjectCaseRule struct {
	BaseRule
	isConventional bool   // Whether to treat as a conventional commit format
	caseChoice     string // The desired case ("upper", "lower", or "ignore")
	allowNonAlpha  bool   // Whether to allow non-alphabetic first characters
	firstWord      string // Store for verbose output
	firstLetter    rune   // Store for verbose output
}

// SubjectCaseOption is a function that modifies a SubjectCaseRule.
type SubjectCaseOption func(SubjectCaseRule) SubjectCaseRule

// WithCaseChoice sets the desired case for the subject.
func WithCaseChoice(caseChoice string) SubjectCaseOption {
	return func(rule SubjectCaseRule) SubjectCaseRule {
		if caseChoice == "upper" || caseChoice == "lower" || caseChoice == "ignore" {
			rule.caseChoice = caseChoice
		}

		return rule
	}
}

// WithSubjectCaseCommitFormat configures whether to treat as a conventional commit.
func WithSubjectCaseCommitFormat(isConventional bool) SubjectCaseOption {
	return func(rule SubjectCaseRule) SubjectCaseRule {
		rule.isConventional = isConventional

		return rule
	}
}

// WithAllowNonAlpha sets whether to allow non-alphabetic first characters.
func WithAllowNonAlpha(allow bool) SubjectCaseOption {
	return func(rule SubjectCaseRule) SubjectCaseRule {
		rule.allowNonAlpha = allow

		return rule
	}
}

// NewSubjectCaseRule creates a new SubjectCaseRule with the specified options.
func NewSubjectCaseRule(options ...SubjectCaseOption) SubjectCaseRule {
	rule := SubjectCaseRule{
		BaseRule:       NewBaseRule("SubjectCase"),
		isConventional: false,
		caseChoice:     "lower", // Default to lowercase
		allowNonAlpha:  false,   // Default to requiring alphabetic first characters
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// NewSubjectCaseRuleWithConfig creates a SubjectCaseRule using configuration.
func NewSubjectCaseRuleWithConfig(config domain.SubjectConfigProvider, conventionalConfig domain.ConventionalConfigProvider) SubjectCaseRule {
	// Build options based on the configuration
	var options []SubjectCaseOption

	// Set the case choice if provided
	if caseChoice := config.SubjectCase(); caseChoice != "" {
		options = append(options, WithCaseChoice(caseChoice))
	}

	// Check if we need to use conventional commit format
	if conventionalConfig.ConventionalRequired() {
		options = append(options, WithSubjectCaseCommitFormat(true))
	}

	// If imperative mood is enforced, allow non-alphabetic characters
	if config.SubjectRequireImperative() {
		options = append(options, WithAllowNonAlpha(true))
	}

	return NewSubjectCaseRule(options...)
}

// Name returns the rule identifier.
func (r SubjectCaseRule) Name() string {
	return r.BaseRule.Name()
}

// SetErrors sets the validation errors for this rule.
// This method supports value semantics by returning a new instance.
func (r SubjectCaseRule) SetErrors(errors []appErrors.ValidationError) SubjectCaseRule {
	rule := r
	rule.BaseRule.errors = errors

	return rule
}

// ClearErrors removes all existing validation errors.
// This method supports value semantics by returning a new instance.
func (r SubjectCaseRule) ClearErrors() SubjectCaseRule {
	rule := r
	rule.BaseRule = rule.BaseRule.WithClearedErrors()

	return rule
}

// AddError adds a structured validation error.
// This method supports value semantics by returning a new instance.
func (r SubjectCaseRule) AddError(err appErrors.ValidationError) SubjectCaseRule {
	rule := r
	rule.BaseRule = rule.BaseRule.WithError(err)

	return rule
}

// AddErrorWithCode adds a validation error with the specified code and message.
// This method supports value semantics by returning a new instance.
func (r SubjectCaseRule) AddErrorWithCode(code appErrors.ValidationErrorCode, message string) SubjectCaseRule {
	rule := r
	rule.BaseRule = rule.BaseRule.WithErrorWithCode(code, message)

	return rule
}

// AddErrorWithContext adds a validation error with context information.
// This method supports value semantics by returning a new instance.
func (r SubjectCaseRule) AddErrorWithContext(code appErrors.ValidationErrorCode, message string, context map[string]string) SubjectCaseRule {
	rule := r
	rule.BaseRule = rule.BaseRule.WithErrorWithContext(code, message, context)

	return rule
}

// HasErrors returns true if the rule has validation errors.
func (r SubjectCaseRule) HasErrors() bool {
	return r.BaseRule.HasErrors()
}

// Errors returns all validation errors found by this rule.
func (r SubjectCaseRule) Errors() []appErrors.ValidationError {
	return r.BaseRule.Errors()
}

// setFirstWord sets the first word and letter for verbose output.
// This method supports value semantics by returning a new instance.
//
//nolint:unused
func (r SubjectCaseRule) setFirstWord(word string) SubjectCaseRule {
	r.firstWord = word
	first, _ := utf8.DecodeRuneInString(word)
	r.firstLetter = first

	return r
}

// ValidateWithState performs validation and returns errors along with an updated rule state.
// Returns both errors and a new rule instance with updated state.
// Exported for testing purposes.
func ValidateWithState(rule SubjectCaseRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectCaseRule) {
	subject := commit.Subject
	errors := make([]appErrors.ValidationError, 0)

	// Check for empty message first
	if subject == "" {
		// Create error context with rich information
		errorCtx := appErrors.NewContext()

		// Determine example subject based on rule configuration
		var exampleSubject string

		if rule.isConventional {
			var firstWord string
			if rule.caseChoice == "upper" {
				firstWord = "Add"
			} else {
				firstWord = "add"
			}

			exampleSubject = fmt.Sprintf("feat: %s a descriptive subject", firstWord)
		} else {
			var firstWord string
			if rule.caseChoice == "upper" {
				firstWord = "Add"
			} else {
				firstWord = "add"
			}

			exampleSubject = firstWord + " a descriptive subject"
		}

		helpMessage := fmt.Sprintf(`Empty Subject Error: Cannot validate case on an empty subject.

Your commit has an empty subject line, so case validation cannot be performed.

✅ CORRECT FORMAT:
- A commit message should start with a subject line:
  %s
  
  This is a descriptive body that explains the change in detail.
  It can span multiple lines.

❌ INCORRECT FORMAT:
- Your commit has an empty subject line

WHY THIS MATTERS:
- The subject line is the most visible part of a commit message
- It provides a concise summary of changes that appears in logs
- Without a subject, it's difficult to identify the purpose of the commit

NEXT STEPS:
1. Add a meaningful subject line to your commit
   - Use 'git commit --amend' to edit your most recent commit
   - Follow your project's capitalization conventions (%scase first letter)
   
2. If using conventional commits, remember the format:
   type(scope): subject`,
			exampleSubject,
			rule.caseChoice)

		// Create the error
		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrEmptyDescription,
			"subject is empty",
			helpMessage,
			errorCtx,
		)
		errors = append(errors, err)

		// Return with updated rule (using value semantics)
		return errors, rule.SetErrors(errors)
	}

	// Extract first word
	firstWord, err := extractSubjectCaseFirstWord(rule.isConventional, subject)
	if err != nil {
		// Determine the specific error type
		var errorCode appErrors.ValidationErrorCode

		if rule.isConventional {
			if strings.Contains(err.Error(), "missing subject after type") {
				errorCode = appErrors.ErrMissingSubject
			} else {
				errorCode = appErrors.ErrInvalidFormat
			}
		} else {
			errorCode = appErrors.ErrInvalidFormat
		}

		context := map[string]string{
			"subject":         subject,
			"is_conventional": strconv.FormatBool(rule.isConventional),
		}

		// Create error context with rich information
		errorCtx := appErrors.NewContext()

		var helpMessage string

		if rule.isConventional && strings.Contains(err.Error(), "conventional commit format") {
			helpMessage = fmt.Sprintf(`Invalid Format Error: The commit message does not follow conventional format.

Your commit message does not follow the conventional commit format required by this project.

✅ CORRECT FORMAT:
- Conventional commit format follows this pattern:
  type(optional scope): description
  
  Examples:
  feat: add new user registration
  fix(auth): resolve login timeout issue
  docs(readme): update installation instructions

❌ INCORRECT FORMAT:
- Your message: "%s"
- This doesn't match the expected pattern

WHY THIS MATTERS:
- Conventional commits provide a structured commit history
- They enable automated tools to generate changelogs
- They make it easier to understand the purpose of each commit

NEXT STEPS:
1. Rewrite your commit message following the conventional format:
   - Choose a type (feat, fix, docs, style, refactor, perf, test, etc.)
   - Add optional scope in parentheses if relevant
   - Add description after the colon and space
   
2. Use 'git commit --amend' to edit your most recent commit
   
3. Remember to apply the correct case (%scase) to the first letter of your description`,
				subject,
				rule.caseChoice)
		} else if strings.Contains(err.Error(), "missing subject") {
			helpMessage = fmt.Sprintf(`Missing Subject Error: No subject after conventional prefix.

Your commit has a conventional prefix but is missing the subject description.

✅ CORRECT FORMAT:
- Conventional commit format needs a description after the prefix:
  type(optional scope): description
  
  Examples:
  feat: add new user registration
  fix(auth): resolve login timeout issue

❌ INCORRECT FORMAT:
- Your message just has the prefix without a description: "%s"

WHY THIS MATTERS:
- The subject description explains what the commit does
- Without a subject, the commit's purpose is unclear
- Complete conventional commits follow a specific format

NEXT STEPS:
1. Add a descriptive subject after the conventional prefix:
   - Make it concise but informative
   - Start with a %scase letter as required by your project
   
2. Use 'git commit --amend' to edit your most recent commit`,
				subject,
				rule.caseChoice)
		} else {
			// Determine example format based on rule configuration
			var exampleFormat string

			if rule.isConventional {
				var firstWord string
				if rule.caseChoice == "upper" {
					firstWord = "Add"
				} else {
					firstWord = "add"
				}

				exampleFormat = fmt.Sprintf("feat: %s a descriptive subject", firstWord)
			} else {
				var firstWord string
				if rule.caseChoice == "upper" {
					firstWord = "Add"
				} else {
					firstWord = "add"
				}

				exampleFormat = firstWord + " a descriptive subject"
			}

			helpMessage = fmt.Sprintf(`Invalid Format Error: Cannot determine the first word to check case.

Your commit message has a format issue that prevents case validation.

✅ CORRECT FORMAT:
- Commit messages should start with a word that can be checked for case:
  %s
  
❌ INCORRECT FORMAT:
- Your message: "%s"
- No valid first word could be found for case checking

WHY THIS MATTERS:
- Clear, well-formatted commit messages are important for project history
- Following consistent formatting makes the repository more professional
- Case conventions help with readability and consistency

NEXT STEPS:
1. Ensure your commit message starts with a valid word
   - Start with a letter (not a symbol, number, or whitespace)
   - Follow your project's case conventions (%scase)
   
2. Use 'git commit --amend' to edit your most recent commit`,
				exampleFormat,
				subject,
				rule.caseChoice)
		}

		// Create the error
		validationErr := appErrors.CreateRichError(
			rule.Name(),
			errorCode,
			err.Error(),
			helpMessage,
			errorCtx,
		)

		// Store context for backward compatibility
		for k, v := range context {
			validationErr = validationErr.WithContext(k, v)
		}

		errors = append(errors, validationErr)

		// Update rule state using value semantics
		updatedRule := rule.SetErrors(errors)
		updatedRule.firstWord = ""
		updatedRule.firstLetter = 0

		return errors, updatedRule
	}

	// Store first word for verbose output
	updatedRule := rule
	updatedRule.firstWord = firstWord
	first, size := utf8.DecodeRuneInString(firstWord)
	updatedRule.firstLetter = first

	if first == utf8.RuneError && size == 0 {
		// Create the error
		validationErr := appErrors.New(
			rule.Name(),
			appErrors.ErrUnknown,
			"subject does not start with valid UTF-8 text",
			appErrors.WithContextMap(map[string]string{
				"subject": subject,
			}),
		)
		errors = append(errors, validationErr)

		return errors, updatedRule.SetErrors(errors)
	}

	// If AllowNonAlpha is enabled and the first character isn't a letter, skip case check
	if rule.allowNonAlpha && !unicode.IsLetter(first) {
		return errors, updatedRule
	}

	// Validate case
	var valid bool

	var errorCode = appErrors.ErrSubjectCase

	switch rule.caseChoice {
	case "upper":
		valid = unicode.IsUpper(first)
	case "lower":
		valid = unicode.IsLower(first)
	case "ignore":
		valid = true
	default:
		valid = unicode.IsLower(first) // Default to lowercase if unspecified
	}

	if !valid {
		// Create error context with rich information
		errorCtx := appErrors.NewContext()
		actualCase := map[bool]string{true: "upper", false: "lower"}[unicode.IsUpper(first)]

		// Create correct example
		var correctExample, incorrectExample string

		if rule.isConventional {
			if rule.caseChoice == "upper" {
				correctExample = "feat: Add new feature"
				incorrectExample = "feat: add new feature"
			} else {
				correctExample = "feat: add new feature"
				incorrectExample = "feat: Add new feature"
			}
		} else {
			if rule.caseChoice == "upper" {
				correctExample = "Add new feature"
				incorrectExample = "add new feature"
			} else {
				correctExample = "add new feature"
				incorrectExample = "Add new feature"
			}
		}

		// Determine correct and incorrect first letter description
		var correctCaseDescription, incorrectCaseDescription string
		if rule.caseChoice == "upper" {
			correctCaseDescription = "First letter should be capitalized:"
			incorrectCaseDescription = "First letter should NOT be lowercase:"
		} else {
			correctCaseDescription = "First letter should be lowercase:"
			incorrectCaseDescription = "First letter should NOT be uppercase:"
		}

		// Determine action verb
		var actionVerb string
		if rule.caseChoice == "upper" {
			actionVerb = "capitalize"
		} else {
			actionVerb = "use lowercase for"
		}

		// Create correction example
		var suggestionExample string

		if rule.caseChoice == "upper" {
			upperFirst := strings.ToUpper(string(first))
			suggestionExample = fmt.Sprintf("Instead of: \"%s\", use: \"%s%s\"", firstWord, upperFirst, firstWord[1:])
		} else {
			lowerFirst := strings.ToLower(string(first))
			suggestionExample = fmt.Sprintf("Instead of: \"%s\", use: \"%s%s\"", firstWord, lowerFirst, firstWord[1:])
		}

		// Create example
		var caseExample string
		if rule.caseChoice == "upper" {
			caseExample = "Example: \"Add\" not \"add\""
		} else {
			caseExample = "Example: \"add\" not \"Add\""
		}

		helpMessage := fmt.Sprintf(`Subject Case Error: First word should start with %scase.

Your commit subject's first word "%s" does not use the required capitalization.

✅ CORRECT FORMAT:
- %s
  %s

❌ INCORRECT FORMAT:
- %s
  %s

WHY THIS MATTERS:
- Consistent capitalization improves readability of commit history
- It helps maintain a professional and organized project history
- It follows project conventions for commit message style

NEXT STEPS:
1. Edit your commit message to %s the first letter of the subject
   %s
   
2. If using conventional commits, remember the format is:
   type(scope): subject  <-- apply case rules to first word of subject

3. For single-word commits, pay extra attention to the first letter
   %s`,
			rule.caseChoice,
			firstWord,
			correctCaseDescription,
			correctExample,
			incorrectCaseDescription,
			incorrectExample,
			actionVerb,
			suggestionExample,
			caseExample)

		// Create the error
		validationErr := appErrors.CreateRichError(
			rule.Name(),
			errorCode,
			"commit subject case is not "+rule.caseChoice,
			helpMessage,
			errorCtx,
		)

		// Store context for backward compatibility
		validationErr = validationErr.WithContext("expected_case", rule.caseChoice)
		validationErr = validationErr.WithContext("actual_case", actualCase)
		validationErr = validationErr.WithContext("first_word", firstWord)
		validationErr = validationErr.WithContext("subject", subject)

		errors = append(errors, validationErr)
	}

	return errors, updatedRule.SetErrors(errors)
}

// Validate validates the commit subject case.
func (r SubjectCaseRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Use the pure function approach
	errors, _ := ValidateWithState(r, commit)

	return errors
}

// resultImpl returns a concise result message.
func resultImpl(rule SubjectCaseRule) string {
	if rule.HasErrors() {
		// Check for case-specific error
		errors := rule.Errors()
		if len(errors) > 0 {
			switch appErrors.ValidationErrorCode(errors[0].Code) { //nolint:exhaustive
			case appErrors.ErrSubjectCase:
				return "Subject should start with " + rule.caseChoice
			case appErrors.ErrEmptyDescription, appErrors.ErrEmptyMessage:
				return "Subject is empty"
			case appErrors.ErrInvalidFormat:
				return "Invalid format"
			default:
				return "Subject case validation failed"
			}
		}

		return "Invalid subject case"
	}

	return "Subject case is correct"
}

// Result returns a concise result message.
func (r SubjectCaseRule) Result() string {
	return resultImpl(r)
}

// verboseResultImpl returns a more detailed explanation for verbose mode.
func verboseResultImpl(rule SubjectCaseRule) string {
	if rule.HasErrors() {
		// Get errors
		errors := rule.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		// We're deliberately not handling all possible validation error codes here,
		// just the ones that can be generated by this specific rule.

		switch appErrors.ValidationErrorCode(validationErr.Code) { //nolint:exhaustive
		case appErrors.ErrEmptyDescription, appErrors.ErrEmptyMessage:
			return "Commit subject is empty. Cannot validate case."

		case appErrors.ErrUnknown:
			errMsg := validationErr.Message
			if strings.Contains(errMsg, "UTF-8") {
				return "Subject doesn't start with valid UTF-8 text. Check for encoding issues."
			}

		case appErrors.ErrInvalidFormat:
			errMsg := validationErr.Message
			if strings.Contains(errMsg, "conventional commit format") {
				return "Invalid conventional commit format. Expected 'type(scope): subject'."
			} else if strings.Contains(errMsg, "no valid first word") {
				return "Invalid commit format. Subject should start with a valid word."
			}

		case appErrors.ErrMissingSubject:
			return "Missing subject after conventional commit prefix."

		case appErrors.ErrSubjectCase:
			if rule.caseChoice == "upper" {
				return "First letter should be uppercase. Found '" + string(rule.firstLetter) + "' in '" + rule.firstWord + "'."
			} else if rule.caseChoice == "lower" {
				return "First letter should be lowercase. Found '" + string(rule.firstLetter) + "' in '" + rule.firstWord + "'."
			}
		}

		// Default case
		return validationErr.Error()
	}

	// Construct a detailed success message
	var formatType string
	if rule.isConventional {
		formatType = "conventional commit"
	} else {
		formatType = "standard commit"
	}

	return "Subject has correct " + rule.caseChoice + "case for " + formatType + ": '" + rule.firstWord + "'"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SubjectCaseRule) VerboseResult() string {
	return verboseResultImpl(r)
}

// helpImpl returns a description of how to fix the rule violation.
func helpImpl(rule SubjectCaseRule) string {
	if !rule.HasErrors() {
		return "No errors to fix. This rule checks that commit message subjects follow the required case style (" + rule.caseChoice + " case) for consistency."
	}

	// Get errors
	errors := rule.Errors()
	if len(errors) == 0 {
		return "No specific guidance available"
	}

	// errors[0] is already a ValidationError, so no need for type assertion
	validationErr := errors[0]
	// We're deliberately not handling all possible validation error codes here,
	// just the ones that can be generated by this specific rule.

	switch appErrors.ValidationErrorCode(validationErr.Code) { //nolint:exhaustive
	case appErrors.ErrEmptyDescription, appErrors.ErrEmptyMessage:
		return "Provide a non-empty commit message subject with appropriate capitalization."

	case appErrors.ErrUnknown:
		if strings.Contains(validationErr.Message, "UTF-8") {
			return "Ensure your commit message begins with valid UTF-8 text. Remove any invalid characters from the start."
		}

	case appErrors.ErrInvalidFormat:
		if strings.Contains(validationErr.Message, "conventional commit format") {
			return "Format your commit message according to the Conventional Commits specification: type(scope): subject\n" +
				"Example: feat(auth): Add login feature"
		}

		return "Ensure your commit message starts with a valid word following proper capitalization rules."

	case appErrors.ErrMissingSubject:
		return "Add a subject after the type(scope): prefix in your conventional commit message.\n" +
			"Example: fix(api): Resolve timeout issue"

	case appErrors.ErrSubjectCase:
		if rule.caseChoice == "upper" {
			return "Capitalize the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): Add feature (not 'add feature')\n" +
				"Example for standard commit: Add feature (not 'add feature')"
		} else if rule.caseChoice == "lower" {
			return "Use lowercase for the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): add feature (not 'Add feature')\n" +
				"Example for standard commit: add feature (not 'Add feature')"
		}
	}

	// Default help
	return "Check the capitalization of the first letter in your commit message subject according to your project's guidelines."
}

// Help returns a description of how to fix the rule violation.
func (r SubjectCaseRule) Help() string {
	return helpImpl(r)
}

// extractSubjectCaseFirstWord extracts the first word from the commit message.
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
func extractSubjectCaseFirstWord(isConventional bool, subject string) (string, error) {
	if isConventional {
		// For conventional commits, try to extract the part after type(scope):
		matches := subjectCaseRegex.FindStringSubmatch(subject)

		// Validate conventional commit format
		if len(matches) != 5 {
			return "", errors.New("invalid conventional commit format")
		}

		msg := matches[4]
		if msg == "" {
			return "", errors.New("missing subject after type")
		}

		matches = subjectCaseFirstWordRegex.FindStringSubmatch(msg)
		if len(matches) == 0 {
			return "", errors.New("no valid first word found")
		}

		return matches[1], nil
	}

	// For non-conventional commits
	matches := subjectCaseFirstWordRegex.FindStringSubmatch(subject)
	if len(matches) == 0 {
		return "", errors.New("no valid first word found")
	}

	return matches[1], nil
}
