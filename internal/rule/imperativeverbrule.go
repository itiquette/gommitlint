// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jdkato/prose/v3"
	"github.com/pkg/errors"
)

// imperativeTags represents verb tags that are not imperative.
var imperativeTags = map[string]bool{
	"VBD": true, // Past tense
	"VBG": true, // Gerund
	"VBZ": true, // 3rd person singular present
}

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

	// Create prose document to analyze verb
	doc, err := createProseDocument(word)
	if err != nil {
		rule.errors = append(rule.errors, fmt.Errorf("failed to create document: %w", err))

		return rule
	}

	// Validate verb type
	if err := validateVerbType(doc, word); err != nil {
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

// createProseDocument creates a prose document for verb analysis.
func createProseDocument(word string) (*prose.Document, error) {
	return prose.NewDocument("I " + strings.ToLower(word))
}

func validateVerbType(doc *prose.Document, word string) error {
	tokens := doc.Tokens()
	if len(tokens) != 2 {
		return fmt.Errorf("expected 2 tokens, got %d", len(tokens))
	}

	tok := tokens[1]
	if imperativeTags[tok.Tag] {
		return fmt.Errorf("first word of commit must be an imperative verb: %q is invalid", word)
	}

	return nil
}

// firstWordRegex is the regular expression used to find the first word in a commit.
var firstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

func parseSubject(msg string) []string {
	subject := strings.Split(msg, "\n")[0]
	groups := SubjectRegex.FindStringSubmatch(subject)

	return groups
}
