// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/client9/misspell"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// Configuration context key for spell checking.
type spellCheckConfigKey string

const (
	// Key for storing spell check configuration in context.
	spellCheckKey spellCheckConfigKey = "spell_check_config"
)

// SpellCheckConfig represents configuration for spell checking.
type SpellCheckConfig struct {
	// Enabled indicates whether spell checking is enabled.
	Enabled bool
	// Language specifies the language/locale to use for spell checking.
	Language string
	// CustomDictionary contains words to ignore during spell checking.
	CustomDictionary []string
	// IgnoreCase indicates whether to ignore case when checking spelling.
	IgnoreCase bool
}

// DefaultSpellCheckConfig returns default configuration for spell checking.
func DefaultSpellCheckConfig() SpellCheckConfig {
	return SpellCheckConfig{
		Enabled:          true,
		Language:         "en_US",
		CustomDictionary: []string{},
		IgnoreCase:       true,
	}
}

// WithSpellCheckConfig adds spell check configuration to the context.
func WithSpellCheckConfig(ctx context.Context, config SpellCheckConfig) context.Context {
	return context.WithValue(ctx, spellCheckKey, config)
}

// SpellCheckConfigFromContext extracts spell check configuration from the context.
func SpellCheckConfigFromContext(ctx context.Context) SpellCheckConfig {
	if config, ok := ctx.Value(spellCheckKey).(SpellCheckConfig); ok {
		return config
	}

	return DefaultSpellCheckConfig()
}

// SpellRule validates spelling in commit messages.
type SpellRule struct {
	baseRule     BaseRule
	ignoreCase   bool
	ignoreWords  []string
	locale       string
	misspellings map[string]string
}

// SpellRuleOption configures a SpellRule.
type SpellRuleOption func(SpellRule) SpellRule

// WithIgnoreCase configures case sensitivity for spell checking.
func WithIgnoreCase(ignore bool) SpellRuleOption {
	return func(rule SpellRule) SpellRule {
		result := rule
		result.ignoreCase = ignore

		return result
	}
}

// WithIgnoreWords configures words to ignore during spell checking.
func WithIgnoreWords(words []string) SpellRuleOption {
	return func(rule SpellRule) SpellRule {
		result := rule
		// Create a deep copy of the words slice
		result.ignoreWords = make([]string, len(words))
		copy(result.ignoreWords, words)

		return result
	}
}

// WithLocale configures the locale for spell checking.
func WithLocale(locale string) SpellRuleOption {
	return func(rule SpellRule) SpellRule {
		result := rule
		result.locale = locale

		return result
	}
}

// Note: There are no test-specific options in the implementation

// WithMaxErrors sets the maximum number of spelling errors before the check fails.
func WithMaxErrors(_ int) SpellRuleOption {
	return func(rule SpellRule) SpellRule {
		// Currently, all errors are reported regardless of count
		result := rule

		return result
	}
}

// WithCustomWords adds custom words to the dictionary for spell checking.
func WithCustomWords(words []string) SpellRuleOption {
	return func(rule SpellRule) SpellRule {
		result := rule
		// Create a new slice with the combined words
		newIgnoreWords := make([]string, len(result.ignoreWords), len(result.ignoreWords)+len(words))
		copy(newIgnoreWords, result.ignoreWords)
		result.ignoreWords = append(newIgnoreWords, words...)

		return result
	}
}

// WithCustomWordsMap accepts a map of custom words for compatibility with older code.
// It extracts the keys and passes them to WithCustomWords.
func WithCustomWordsMap(wordsMap map[string]string) SpellRuleOption {
	return func(rule SpellRule) SpellRule {
		// Extract keys from the map
		words := make([]string, 0, len(wordsMap))
		for word := range wordsMap {
			words = append(words, word)
		}

		// Use the regular WithCustomWords function
		return WithCustomWords(words)(rule)
	}
}

// NewSpellRule creates a new rule for validating spelling in commit messages.
func NewSpellRule(options ...SpellRuleOption) SpellRule {
	// Create a rule with default values
	rule := SpellRule{
		baseRule:     NewBaseRule("Spell"),
		ignoreCase:   true,
		ignoreWords:  []string{},
		locale:       "en_US",
		misspellings: make(map[string]string),
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate checks the spelling in commit messages using configuration from context.
func (r SpellRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating spelling using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Skip validation if spell checking is disabled in config
	if !rule.isEnabled(ctx) {
		logger.Debug().Msg("Spell checking is disabled in configuration")

		return []appErrors.ValidationError{}
	}

	// Use the validation logic
	errors, _ := validateSpellWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r SpellRule) withContextConfig(ctx context.Context) SpellRule {
	logger := log.Logger(ctx)

	// Get configuration from context - try direct SpellCheckConfig first, then fallback to config
	var spellConfig SpellCheckConfig

	// Check for direct SpellCheckConfig in context
	directConfig := SpellCheckConfigFromContext(ctx)
	if directConfig.Language != "" {
		// Use the direct config if it exists
		spellConfig = directConfig

		logger.Trace().Msg("Using direct SpellCheckConfig from context")
	} else {
		// Fallback to config
		cfg := config.GetConfig(ctx)

		// Extract spell check config from config
		spellConfig = SpellCheckConfig{
			Enabled:          cfg.SpellCheck.Enabled,
			Language:         cfg.SpellCheck.Language,
			CustomDictionary: cfg.SpellCheck.CustomDictionary,
			IgnoreCase:       cfg.SpellCheck.IgnoreCase,
		}

		logger.Trace().Msg("Using SpellCheckConfig from config")
	}

	// Create a configured rule starting with current settings
	result := r

	// Apply configuration
	if spellConfig.Language != "" {
		result.locale = spellConfig.Language
	}

	if len(spellConfig.CustomDictionary) > 0 {
		// Create a deep copy of the custom dictionary
		result.ignoreWords = make([]string, len(spellConfig.CustomDictionary))
		copy(result.ignoreWords, spellConfig.CustomDictionary)
	}

	result.ignoreCase = spellConfig.IgnoreCase

	// Log configuration at debug level
	logger.Debug().
		Bool("enabled", spellConfig.Enabled).
		Str("locale", result.locale).
		Strs("custom_dictionary", result.ignoreWords).
		Bool("ignore_case", result.ignoreCase).
		Msg("Spell rule configuration from context")

	return result
}

// isEnabled returns whether spell checking is enabled based on context configuration.
func (r SpellRule) isEnabled(ctx context.Context) bool {
	// Check for direct SpellCheckConfig in context
	directConfig := SpellCheckConfigFromContext(ctx)
	if directConfig.Language != "" {
		return directConfig.Enabled
	}

	// Fallback to config
	cfg := config.GetConfig(ctx)

	return cfg.SpellCheck.Enabled
}

// validateSpellWithState validates spelling and returns both the errors and an updated rule.
func validateSpellWithState(rule SpellRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SpellRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()
	result.misspellings = make(map[string]string)

	// No conditional test mode in implementation

	// Check both subject and body text
	textToCheck := commit.Subject
	if commit.Body != "" {
		textToCheck += "\n\n" + commit.Body
	}

	// Find misspellings with real spell checker
	misspellings := checkSpelling(textToCheck, rule.locale, rule.ignoreCase, rule.ignoreWords)

	// Add any found misspellings to the result
	for word, suggestion := range misspellings {
		result.misspellings[word] = suggestion

		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrSpelling,
			fmt.Sprintf("potentially misspelled word: '%s'", word),
		).
			WithContext("word", word).
			WithContext("suggestion", suggestion)

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
}

// checkSpelling checks the text for misspelled words using the misspell library.
func checkSpelling(text, locale string, ignoreCase bool, ignoreWords []string) map[string]string {
	// Create map of words to ignore for quick lookup
	ignored := make(map[string]bool)

	for _, word := range ignoreWords {
		if ignoreCase {
			ignored[strings.ToLower(word)] = true
		} else {
			ignored[word] = true
		}
	}

	// Extract words from the text
	words := extractWords(text)

	// Create a map to store misspellings
	misspellings := make(map[string]string)

	// Create replacer from misspell library
	replacer := misspell.New()
	replacer.Compile() // Initialize the replacer

	// Handle British locale if specified
	if strings.ToLower(locale) == "en_gb" || strings.ToLower(locale) == "en_uk" {
		// Remove US spellings and add UK spellings
		replacer.RemoveRule(misspell.DictAmerican)
		replacer.AddRuleList(misspell.DictBritish)
		replacer.Compile() // Recompile with new rules
	}

	// Process each word
	for _, word := range words {
		// Skip short words (3 characters or less)
		if len(word) < 4 {
			continue
		}

		// Skip words in the ignore list
		wordToCheck := word
		if ignoreCase {
			wordToCheck = strings.ToLower(word)
		}

		if ignored[wordToCheck] {
			continue
		}

		// Check if word is misspelled
		corrected, diffs := replacer.Replace(word)
		if len(diffs) > 0 && corrected != word {
			misspellings[word] = corrected
		}
	}

	return misspellings
}

// extractWords extracts individual words from text.
func extractWords(text string) []string {
	// Split on spaces and remove punctuation
	fields := strings.Fields(text)

	var words []string

	for _, field := range fields {
		// Strip common punctuation
		word := strings.Trim(field, ".,;:!?\"'()[]{}*_-")
		if word != "" {
			words = append(words, word)
		}
	}

	return words
}

// Name returns the rule name.
func (r SpellRule) Name() string {
	return r.baseRule.Name()
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r SpellRule) SetErrors(errors []appErrors.ValidationError) SpellRule {
	result := r
	result.baseRule = result.baseRule.WithClearedErrors()
	result.misspellings = make(map[string]string)

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)

		// Extract word and suggestion from error context to populate the misspellings map
		if word, exists := err.Context["word"]; exists {
			suggestion := "correct spelling"
			if sug, exists := err.Context["suggestion"]; exists {
				suggestion = sug
			}

			result.misspellings[word] = suggestion
		}
	}

	return result
}

// SetErrorsWithDiffs is a backward compatibility method that accepts both errors and diffs.
// The diffs parameter is ignored in the new implementation.
func (r SpellRule) SetErrorsWithDiffs(errors []appErrors.ValidationError, _ interface{}) SpellRule {
	// Just delegate to the standard SetErrors method
	return r.SetErrors(errors)
}

// Errors returns all validation errors found by this rule.
func (r SpellRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SpellRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// Result returns a concise validation result.
func (r SpellRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return fmt.Sprintf("❌ %d possible misspelled word(s)", len(errors))
	}

	return "✓ No spelling errors"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SpellRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		var details strings.Builder

		details.WriteString(fmt.Sprintf("❌ Found %d potential spelling error(s):\n", len(errors)))

		// Get misspelled words and sort them
		var words []string
		for word := range r.misspellings {
			words = append(words, word)
		}

		sort.Strings(words)

		// Add each misspelling with suggestion
		for _, word := range words {
			suggestion := r.misspellings[word]
			details.WriteString(fmt.Sprintf("   - '%s' (did you mean '%s'?)\n", word, suggestion))
		}

		return details.String()
	}

	return "✓ No spelling errors found in commit message"
}

// Help returns guidance for fixing rule violations.
func (r SpellRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	help := "Potential spelling errors detected in your commit message.\n\n"
	help += "Suggestions:\n"

	// Get misspelled words and sort them
	words := make([]string, 0, len(r.misspellings))
	for word := range r.misspellings {
		words = append(words, word)
	}

	sort.Strings(words)

	// Add each misspelling with suggestion
	for _, word := range words {
		suggestion := r.misspellings[word]
		help += fmt.Sprintf("- Replace '%s' with '%s'\n", word, suggestion)
	}

	help += "\nIf these words are correct (e.g., technical terms), you can:\n"
	help += "1. Add them to your project-specific dictionary\n"
	help += "2. Configure the spell rule to ignore specific words\n"

	return help
}
