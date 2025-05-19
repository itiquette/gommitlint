// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/client9/misspell"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SpellRule validates spelling in commit messages.
type SpellRule struct {
	name      string
	verbosity string
}

// SpellOption represents optional parameters for the spell rule.
type SpellOption func(SpellRule) SpellRule

// WithCustomDictionary option for adding custom words to the spell checker.
func WithCustomDictionary(_ []string) SpellOption {
	return func(r SpellRule) SpellRule {
		return r
	}
}

// NewSpellRule creates a new SpellRule.
func NewSpellRule() SpellRule {
	return SpellRule{
		name: "Spell",
	}
}

// Name returns the rule name.
func (r SpellRule) Name() string {
	return r.name
}

// Validate checks spelling in the commit message.
func (r SpellRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Log validation at trace level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Checking spelling",
		"rule", r.Name(),
		"commit_hash", commit.Hash)

	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		// If no config is available, return nil (no validation)
		return nil
	}

	// Check if spell checking is enabled
	if !cfg.GetBool("spell.enabled") {
		logger.Debug("Spell checking disabled")

		return nil
	}

	// Define ignored words from configuration
	ignoreWords := cfg.GetStringSlice("spell.ignore_words")
	ignoreWordsMap := make(map[string]bool)

	for _, word := range ignoreWords {
		ignoreWordsMap[strings.ToLower(word)] = true
	}

	// Log configuration at debug level
	logger.Debug("Spell check configuration from context",
		"enabled", true,
		"ignored_words", ignoreWords)

	// Create spell checker
	replacer := misspell.New()

	// Process commit subject and body
	textToCheck := r.preprocessText(commit.Subject + " " + commit.Body)

	// Skip spell check if text is empty after preprocessing
	if strings.TrimSpace(textToCheck) == "" {
		logger.Debug("No text to check after preprocessing")

		return nil
	}

	// Check spelling
	_, diffs := replacer.Replace(textToCheck)

	// Convert diffs to misspellings
	validMisspellings := []struct {
		word     string
		position int
	}{}

	for _, diff := range diffs {
		word := diff.Original
		// Skip ignored words
		if ignoreWordsMap[strings.ToLower(word)] {
			continue
		}
		// Also check if it's a technical term that should be ignored
		if r.isTechnicalTerm(word) {
			continue
		}

		validMisspellings = append(validMisspellings, struct {
			word     string
			position int
		}{
			word:     word,
			position: diff.Column,
		})
	}

	// Log results at debug level
	logger.Debug("Spell check complete",
		"total_misspellings", len(diffs),
		"valid_misspellings", len(validMisspellings),
		"ignored_count", len(diffs)-len(validMisspellings))

	// Create errors with enhanced context
	errors := make([]appErrors.ValidationError, 0, len(validMisspellings))

	for _, misspelling := range validMisspellings {
		// Create context data for the error
		contextData := map[string]string{
			"word":     misspelling.word,
			"position": strconv.Itoa(misspelling.position),
		}

		if r.verbosity == "verbose" {
			// In verbose mode, include context around the misspelling
			contextSnippet := r.getContextSnippet(textToCheck, misspelling.position, misspelling.word)
			contextData["context"] = contextSnippet
		}

		var err appErrors.ValidationError
		err = appErrors.New(
			r.Name(),
			appErrors.ErrSpellCheckFailed,
			fmt.Sprintf("Spelling error: '%s'", misspelling.word),
		)

		// Add all context data
		for k, v := range contextData {
			err = err.WithContext(k, v)
		}

		errors = append(errors, err)
	}

	return errors
}

// preprocessText prepares text for spell checking by cleaning up special characters.
func (r SpellRule) preprocessText(text string) string {
	// Remove markdown-style headers
	lines := strings.Split(text, "\n")
	processedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip lines that are just headers
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		processedLines = append(processedLines, line)
	}

	text = strings.Join(processedLines, " ")

	// Replace various separators with spaces
	replacer := strings.NewReplacer(
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
		":", " ",
		"/", " ",
		"-", " ",
		"_", " ",
	)
	text = replacer.Replace(text)

	// Remove multiple consecutive spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return strings.TrimSpace(text)
}

// isTechnicalTerm checks if a word is a common technical term that should be ignored.
func (r SpellRule) isTechnicalTerm(word string) bool {
	// Common technical terms and abbreviations that spell checkers often flag
	technicalTerms := map[string]bool{
		"api":       true,
		"apis":      true,
		"cli":       true,
		"json":      true,
		"yaml":      true,
		"yml":       true,
		"xml":       true,
		"url":       true,
		"urls":      true,
		"uri":       true,
		"uris":      true,
		"http":      true,
		"https":     true,
		"ftp":       true,
		"tcp":       true,
		"udp":       true,
		"ip":        true,
		"ui":        true,
		"ux":        true,
		"css":       true,
		"js":        true,
		"ts":        true,
		"jsx":       true,
		"tsx":       true,
		"sql":       true,
		"db":        true,
		"uuid":      true,
		"guid":      true,
		"sha":       true,
		"md5":       true,
		"oauth":     true,
		"jwt":       true,
		"env":       true,
		"config":    true,
		"configs":   true,
		"changelog": true,
		"bugfix":    true,
		"bugfixes":  true,
		"refactor":  true,
		"async":     true,
		"auth":      true,
		"utils":     true,
		"util":      true,
		"bool":      true,
		"int":       true,
		"uint":      true,
		"str":       true,
		"regex":     true,
		"regexp":    true,
		"impl":      true,
	}

	return technicalTerms[strings.ToLower(word)]
}

// getContextSnippet extracts a snippet of text around a misspelling.
func (r SpellRule) getContextSnippet(text string, position int, word string) string {
	const contextSize = 30 // Characters before and after

	// Ensure position is within bounds
	if position < 0 || position >= len(text) {
		return word
	}

	contextStart := position - contextSize
	if contextStart < 0 {
		contextStart = 0
	}

	contextEnd := position + len(word) + contextSize
	if contextEnd > len(text) {
		contextEnd = len(text)
	}

	snippet := text[contextStart:contextEnd]

	// Add ellipsis if truncated
	if contextStart > 0 {
		snippet = "..." + snippet
	}

	if contextEnd < len(text) {
		snippet = snippet + "..."
	}

	return snippet
}

// WithVerbosity returns a new SpellRule with the specified verbosity level.
func (r SpellRule) WithVerbosity(verbosity string) SpellRule {
	newRule := r
	newRule.verbosity = verbosity

	return newRule
}

// Result returns a formatted result string.
func (r SpellRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return r.Name() + ": Passed"
	}

	// Extract unique misspelled words
	misspelledWords := make(map[string]bool)

	for _, err := range errors {
		if word, ok := err.Context["word"]; ok {
			misspelledWords[word] = true
		}
	}

	wordList := make([]string, 0, len(misspelledWords))
	for word := range misspelledWords {
		wordList = append(wordList, word)
	}

	sort.Strings(wordList)

	return fmt.Sprintf("%s: Found misspellings: %s", r.Name(), strings.Join(wordList, ", "))
}

// VerboseResult returns a detailed result string.
func (r SpellRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return r.Name() + ": Passed - No spelling errors found"
	}

	var result strings.Builder

	result.WriteString(fmt.Sprintf("%s: Found %d spelling error(s):\n", r.Name(), len(errors)))

	for i, err := range errors {
		ctx := err.Context
		word := ctx["word"]
		position := ctx["position"]

		result.WriteString(fmt.Sprintf("%d. '%s' at position %s", i+1, word, position))
		result.WriteString("\n")

		if contextSnippet, ok := ctx["context"]; ok {
			result.WriteString(fmt.Sprintf("   Context: %s\n", contextSnippet))
		}
	}

	return result.String()
}

// Help returns guidance for fixing errors from this rule.
func (r SpellRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "All words are spelled correctly."
	}

	var help strings.Builder

	help.WriteString("Fix the following spelling errors:\n")

	for _, err := range errors {
		ctx := err.Context
		word := ctx["word"]

		help.WriteString(fmt.Sprintf("- '%s'", word))
		help.WriteString("\n")
	}

	help.WriteString("\nYou can also add words to your personal dictionary or use the ignore_words configuration.")

	return help.String()
}

// Errors is not used (we return errors directly from Validate).
func (r SpellRule) Errors() []appErrors.ValidationError {
	return nil
}
