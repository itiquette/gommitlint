// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SubjectRule validates commit subject length, case, suffix, and imperative mood.
type SubjectRule struct {
	maxLength           int
	caseChoice          string
	invalidSuffixes     string
	checkCommit         bool
	allowNonAlpha       bool
	requireImperative   bool
	imperativeValidator *ImperativeValidator // Modular imperative validation
}

// NewSubjectRule creates a new SubjectRule from config.
func NewSubjectRule(cfg config.Config) SubjectRule {
	maxLength := cfg.Message.Subject.MaxLength
	if maxLength <= 0 {
		maxLength = 100
	}

	caseChoice := cfg.Message.Subject.Case
	if caseChoice == "" {
		caseChoice = "lower" // Default
	}
	// Only support lower, upper, and ignore
	if caseChoice != "lower" && caseChoice != "upper" && caseChoice != "ignore" {
		caseChoice = "lower"
	}

	invalidSuffixes := ".,;:!?"

	if len(cfg.Message.Subject.ForbidEndings) > 0 {
		var builder strings.Builder
		for _, suffix := range cfg.Message.Subject.ForbidEndings {
			builder.WriteString(suffix)
		}

		if builder.Len() > 0 {
			invalidSuffixes = builder.String()
		}
	}

	isConventionalEnabled := domain.IsRuleActive("conventional", cfg.Rules.Enabled, cfg.Rules.Disabled)

	var imperativeValidator *ImperativeValidator
	if cfg.Message.Subject.RequireImperative {
		imperativeValidator = NewImperativeValidator()
	}

	return SubjectRule{
		maxLength:           maxLength,
		caseChoice:          caseChoice,
		invalidSuffixes:     invalidSuffixes,
		checkCommit:         isConventionalEnabled,
		allowNonAlpha:       false,
		requireImperative:   cfg.Message.Subject.RequireImperative,
		imperativeValidator: imperativeValidator,
	}
}

// Name returns the rule name.
func (r SubjectRule) Name() string {
	return "Subject"
}

// Validate performs pure commit validation.
func (r SubjectRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	var errors []domain.ValidationError

	// Check for empty subject first
	if commit.Subject == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrEmptySubject, "Subject is empty").
				WithContextMap(map[string]string{
					"subject":  commit.Subject,
					"expected": r.caseChoice,
				}).
				WithHelp("Provide a non-empty commit message subject"),
		}
	}

	// Length validation with UTF-8 character counting
	subjectLength := utf8.RuneCountInString(commit.Subject)
	if subjectLength > r.maxLength {
		// Calculate how much over the limit
		excess := subjectLength - r.maxLength

		errors = append(errors,
			domain.New(r.Name(), domain.ErrSubjectTooLong,
				fmt.Sprintf("Subject too long: %d characters (maximum allowed: %d)", subjectLength, r.maxLength)).
				WithContextMap(map[string]string{
					"actual":   strconv.Itoa(subjectLength),
					"expected": fmt.Sprintf("max %d", r.maxLength),
					"subject":  commit.Subject,
				}).
				WithHelp(fmt.Sprintf("Shorten your commit message subject line by %d characters. "+
					"A good subject should be brief but descriptive, ideally under 50 characters.", excess)))
	}

	// Case validation
	if caseErrors := r.validateCase(commit.Subject); len(caseErrors) > 0 {
		errors = append(errors, caseErrors...)
	}

	// Suffix validation
	if suffixErrors := r.validateSuffix(commit.Subject); len(suffixErrors) > 0 {
		errors = append(errors, suffixErrors...)
	}

	// Imperative validation
	if imperativeErrors := r.validateImperative(commit.Subject); len(imperativeErrors) > 0 {
		errors = append(errors, imperativeErrors...)
	}

	return errors
}

// validateCase validates the case style of commit subjects - only checks first letter.
func (r SubjectRule) validateCase(subject string) []domain.ValidationError {
	// Extract first letter with enhanced conventional commit parsing
	firstLetter, textChecked, err := r.extractFirstLetterWithConventionalHandling(subject)
	if err != nil {
		// Determine specific error type based on the error
		errorCode := domain.ErrInvalidFormat

		// Handle conventional commit errors
		if r.checkCommit {
			if strings.Contains(err.Error(), "missing subject after type") {
				errorCode = domain.ErrMissingConventionalSubject
			} else {
				errorCode = domain.ErrInvalidConventionalFormat
			}
		}

		helpText := "Format your commit message according to the proper convention"

		return []domain.ValidationError{
			domain.New(r.Name(), errorCode, err.Error()).
				WithContextMap(map[string]string{
					"subject": subject,
				}).
				WithHelp(helpText),
		}
	}

	// Check case of first letter
	isValid := false

	switch r.caseChoice {
	case "upper":
		isValid = unicode.IsUpper(firstLetter)
	case "lower":
		isValid = unicode.IsLower(firstLetter)
	case "ignore":
		isValid = true // Always valid when ignoring case
	}

	if !isValid {
		// Get expected letter
		var expectedLetter rune
		if r.caseChoice == "upper" {
			expectedLetter = unicode.ToUpper(firstLetter)
		} else {
			expectedLetter = unicode.ToLower(firstLetter)
		}

		// Extract first word for display
		words := strings.Fields(textChecked)
		firstWord := string(firstLetter)

		if len(words) > 0 {
			firstWord = words[0]
		}

		expectedWord := firstWord

		if r.caseChoice == "upper" {
			runes := []rune(firstWord)
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
				expectedWord = string(runes)
			}
		} else {
			runes := []rune(firstWord)
			if len(runes) > 0 {
				runes[0] = unicode.ToLower(runes[0])
				expectedWord = string(runes)
			}
		}

		var errorCode domain.ValidationErrorCode

		switch r.caseChoice {
		case "upper":
			errorCode = domain.ErrWrongCaseUpper
		case "lower":
			errorCode = domain.ErrWrongCaseLower
		}

		helpMessage := r.generateCaseHelp(r.caseChoice, firstWord, expectedWord)

		// Create concise, actionable basic message
		basicMessage := fmt.Sprintf("First letter '%c' should be '%c'", firstLetter, expectedLetter)

		return []domain.ValidationError{
			domain.New(r.Name(), errorCode, basicMessage).
				WithContextMap(map[string]string{
					"first_word":    firstWord,
					"expected_word": expectedWord,
					"subject":       subject,
					"expected":      r.caseChoice,
				}).
				WithHelp(helpMessage),
		}
	}

	return nil
}

// validateSuffix validates that the commit subject doesn't end with invalid characters using UTF-8.
func (r SubjectRule) validateSuffix(subject string) []domain.ValidationError {
	// Use UTF-8 aware last character extraction
	lastChar, size := utf8.DecodeLastRuneInString(subject)

	// Check for invalid UTF-8
	if lastChar == utf8.RuneError && size == 0 {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrInvalidUTF8, "Subject does not end with valid UTF-8 text").
				WithContextMap(map[string]string{
					"subject": subject,
				}).
				WithHelp("Ensure your commit message contains only valid UTF-8 characters"),
		}
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(r.invalidSuffixes, lastChar) {
		// Generate clean subject suggestion
		cleanSubject := subject
		if size > 0 {
			cleanSubject = subject[:len(subject)-size]
		}

		// Create list of invalid suffixes for help message
		var suffixList []string
		for _, suffix := range r.invalidSuffixes {
			suffixList = append(suffixList, string(suffix))
		}

		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrSubjectSuffix,
				fmt.Sprintf("Subject has invalid suffix %q (invalid suffixes: %q)", string(lastChar), r.invalidSuffixes)).
				WithContextMap(map[string]string{
					"subject":  subject,
					"actual":   string(lastChar),
					"expected": cleanSubject,
				}).
				WithHelp(fmt.Sprintf("Remove the punctuation or special character from the end of your subject line. "+
					"The subject should end with a letter or number, not punctuation like: %s",
					strings.Join(suffixList, ", "))),
		}
	}

	return nil
}

// Helper functions

// extractFirstLetterWithConventionalHandling extracts the first letter from subject with proper conventional commit handling.
func (r SubjectRule) extractFirstLetterWithConventionalHandling(subject string) (rune, string, error) {
	if subject == "" {
		return 0, "", errors.New("subject is empty")
	}

	var textToCheck string

	if r.checkCommit {
		// Use shared conventional commit parser for consistency
		parsed := domain.ParseConventionalCommit(subject)
		if parsed.IsValid {
			description := strings.TrimSpace(parsed.Description)
			if description == "" {
				return 0, "", errors.New("missing subject after type")
			}

			textToCheck = description
		} else if domain.IsConventionalCommitLike(subject) {
			return 0, "", errors.New("invalid conventional commit format")
		} else {
			textToCheck = subject
		}
	} else {
		// When conventional commit rule is disabled, just use the subject as-is
		textToCheck = subject
	}

	textToCheck = strings.TrimSpace(textToCheck)
	if textToCheck == "" {
		return 0, "", errors.New("invalid format")
	}

	// Find first letter
	for _, char := range textToCheck {
		if unicode.IsLetter(char) {
			return char, textToCheck, nil
		}
	}

	return 0, "", errors.New("no letter found")
}

// validateImperative validates that the subject uses imperative mood using the modular validator.
func (r SubjectRule) validateImperative(subject string) []domain.ValidationError {
	if !r.requireImperative || r.imperativeValidator == nil {
		return nil
	}

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return nil
	}

	// Use the sophisticated modular imperative validator
	return r.imperativeValidator.ValidateImperative(subject, subject, r.checkCommit, r.Name())
}

// generateCaseHelp creates helpful error message with examples based on case type.
func (r SubjectRule) generateCaseHelp(caseType, _, _ string) string {
	var examples []string

	switch caseType {
	case "upper":
		if r.checkCommit {
			examples = []string{
				"✓ feat(auth): Add user authentication",
				"✓ fix(api): Fix memory leak in handler",
				"✓ docs: Update README with examples",
			}
		} else {
			examples = []string{
				"✓ Add user authentication feature",
				"✓ Fix memory leak in handler",
				"✓ Update documentation with examples",
			}
		}
	case "lower":
		if r.checkCommit {
			examples = []string{
				"✓ feat(auth): add user authentication",
				"✓ fix(api): fix memory leak in handler",
				"✓ docs: update README with examples",
			}
		} else {
			examples = []string{
				"✓ add user authentication feature",
				"✓ fix memory leak in handler",
				"✓ update documentation with examples",
			}
		}
	}

	helpText := fmt.Sprintf(`Examples of correct %s case:

%s
`,
		caseType,
		strings.Join(examples, "\n"))

	// Only show conventional commit help if the rule is enabled
	if r.checkCommit {
		helpText += `

For conventional commits, case applies after type(scope)`
	}

	return helpText
}
