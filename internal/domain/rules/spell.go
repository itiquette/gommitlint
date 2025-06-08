// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SpellChecker defines the interface for spell checking operations.
type SpellChecker interface {
	CheckText(text string) []domain.Misspelling
}

// SpellRule validates spelling in commit messages.
type SpellRule struct {
	checker     SpellChecker
	ignoreWords []string
}

// NewSpellRule creates a new SpellRule with the provided checker.
func NewSpellRule(checker SpellChecker, cfg config.Config) SpellRule {
	return SpellRule{
		checker:     checker,
		ignoreWords: cfg.Spell.IgnoreWords,
	}
}

// Name returns the rule name.
func (r SpellRule) Name() string {
	return "Spell"
}

// Validate checks spelling in the commit message using functional composition.
func (r SpellRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Functional composition approach
	textToCheck := preprocessText(commit.Subject + " " + commit.Body)

	// Skip spell check if text is empty after preprocessing
	if strings.TrimSpace(textToCheck) == "" {
		return nil
	}

	// Extract misspellings using interface
	misspellings := r.checker.CheckText(textToCheck)
	filteredMisspellings := filterIgnoredWords(misspellings, r.ignoreWords)

	// Convert to validation errors
	return buildSpellErrors(filteredMisspellings, r.Name())
}

// Infrastructure code moved to adapters layer

// preprocessText prepares text for spell checking by cleaning up special characters.
// This is a pure function that doesn't depend on any receiver state.
func preprocessText(text string) string {
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

// extractMisspellings moved to adapter - domain now uses interface

// filterIgnoredWords removes misspellings that should be ignored based on ignore list.
// This is a pure function that filters misspellings without side effects.
func filterIgnoredWords(misspellings []domain.Misspelling, ignoreWords []string) []domain.Misspelling {
	if len(ignoreWords) == 0 {
		return misspellings
	}

	// Create ignore map for efficient lookup
	ignoreMap := make(map[string]bool, len(ignoreWords))
	for _, word := range ignoreWords {
		ignoreMap[strings.ToLower(word)] = true
	}

	filtered := make([]domain.Misspelling, 0, len(misspellings))

	for _, misspelling := range misspellings {
		if !ignoreMap[strings.ToLower(misspelling.Word)] {
			filtered = append(filtered, misspelling)
		}
	}

	return filtered
}

// buildSpellErrors converts misspellings to domain validation errors with rich context.
// This is a pure function that builds error objects without side effects.
func buildSpellErrors(misspellings []domain.Misspelling, ruleName string) []domain.ValidationError {
	if len(misspellings) == 0 {
		return nil
	}

	errors := make([]domain.ValidationError, 0, len(misspellings))

	// Build comprehensive help text with all corrections
	helpText := buildComprehensiveHelp(misspellings)

	for _, misspelling := range misspellings {
		contextMap := map[string]string{
			"actual":   misspelling.Word,
			"expected": misspelling.Suggestion,
		}

		err := domain.New(ruleName, domain.ErrMisspelledWord,
			fmt.Sprintf("Misspelled word: '%s'", misspelling.Word)).
			WithContextMap(contextMap).
			WithHelp(helpText)

		errors = append(errors, err)
	}

	return errors
}

// buildComprehensiveHelp creates detailed help text listing all misspellings and corrections.
// Restored from original implementation to provide rich user guidance.
func buildComprehensiveHelp(misspellings []domain.Misspelling) string {
	if len(misspellings) == 0 {
		return ""
	}

	var helpBuilder strings.Builder

	helpBuilder.WriteString("Found spelling errors:\n")

	// Limit display to first 5 misspellings (like original)
	displayCount := len(misspellings)
	if displayCount > 5 {
		displayCount = 5
	}

	for i := 0; i < displayCount; i++ {
		m := misspellings[i]
		helpBuilder.WriteString(fmt.Sprintf("  • '%s' → '%s'\n", m.Word, m.Suggestion))
	}

	// Add "and X more..." if there are many misspellings
	if len(misspellings) > 5 {
		helpBuilder.WriteString(fmt.Sprintf("  ... and %d more misspellings\n", len(misspellings)-5))
	}

	helpBuilder.WriteString("\nCorrect the spelling errors or add valid technical terms to the ignore list in your configuration.")

	return strings.TrimSpace(helpBuilder.String())
}
