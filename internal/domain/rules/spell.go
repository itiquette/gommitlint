// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"strings"

	"github.com/client9/misspell"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SpellRule validates spelling in commit messages.
type SpellRule struct {
	ignoreWords []string
}

// NewSpellRule creates a new SpellRule from config.
func NewSpellRule(cfg config.Config) SpellRule {
	return SpellRule{
		ignoreWords: cfg.Spell.IgnoreWords,
	}
}

// Name returns the rule name.
func (r SpellRule) Name() string {
	return "Spell"
}

// Validate checks spelling in the commit message.
func (r SpellRule) Validate(commit domain.Commit, _ config.Config) []domain.ValidationError {
	// Create a map of ignored words for efficient lookup
	ignoreWordsMap := make(map[string]bool)
	for _, word := range r.ignoreWords {
		ignoreWordsMap[strings.ToLower(word)] = true
	}

	// Create spell checker
	replacer := misspell.New()

	// Process commit subject and body
	textToCheck := r.preprocessText(commit.Subject + " " + commit.Body)

	// Skip spell check if text is empty after preprocessing
	if strings.TrimSpace(textToCheck) == "" {
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

	// Create failures
	failures := make([]domain.ValidationError, 0, len(validMisspellings))

	for _, misspelling := range validMisspellings {
		failure := domain.New(r.Name(), domain.ErrMisspelledWord,
			"Misspelled word: "+misspelling.word).
			WithHelp("Check spelling or add to ignore list")
		failures = append(failures, failure)
	}

	return failures
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
