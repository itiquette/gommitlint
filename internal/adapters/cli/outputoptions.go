// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/itiquette/gommitlint/internal/adapters/output"
	"github.com/itiquette/gommitlint/internal/domain"
)

// OutputOptions represents how validation results should be formatted and displayed.
// This is a focused value type with single responsibility for output concerns.
type OutputOptions struct {
	Format       string    // "text", "json", "github", "gitlab"
	Verbose      bool      // Show detailed validation results
	VerboseLevel int       // Verbose level (0=quiet, 1=verbose, 2=extra verbose)
	ShowHelp     bool      // Show help text and error codes
	RuleHelp     string    // Show detailed help for a specific rule
	Color        string    // When to colorize: "auto", "always", "never"
	Writer       io.Writer // Where to write output
}

// NewOutputOptions creates OutputOptions with sensible defaults.
func NewOutputOptions(writer io.Writer) OutputOptions {
	return OutputOptions{
		Format:       "text",
		Verbose:      false,
		VerboseLevel: 0,
		ShowHelp:     false,
		RuleHelp:     "",
		Color:        "auto",
		Writer:       writer,
	}
}

// WithFormat returns a new OutputOptions with the specified format.
func (o OutputOptions) WithFormat(format string) OutputOptions {
	o.Format = format

	return o
}

// WithVerbose returns a new OutputOptions with verbose enabled/disabled.
func (o OutputOptions) WithVerbose(verbose bool) OutputOptions {
	o.Verbose = verbose
	if verbose && o.VerboseLevel == 0 {
		o.VerboseLevel = 1
	} else if !verbose {
		o.VerboseLevel = 0
	}

	return o
}

// WithVerboseLevel returns a new OutputOptions with the specified verbose level.
func (o OutputOptions) WithVerboseLevel(level int) OutputOptions {
	o.VerboseLevel = level
	o.Verbose = level > 0

	return o
}

// WithShowHelp returns a new OutputOptions with help text enabled/disabled.
func (o OutputOptions) WithShowHelp(showHelp bool) OutputOptions {
	o.ShowHelp = showHelp

	return o
}

// WithRuleHelp returns a new OutputOptions with rule help.
func (o OutputOptions) WithRuleHelp(ruleHelp string) OutputOptions {
	o.RuleHelp = ruleHelp

	return o
}

// ValidateRuleHelp validates that the specified rule name exists.
func (o OutputOptions) ValidateRuleHelp() error {
	if o.RuleHelp == "" {
		return nil // Empty is valid
	}

	validRules := getValidRuleNames()
	ruleName := strings.TrimSpace(o.RuleHelp)

	// Check exact match first
	for _, valid := range validRules {
		if valid == ruleName {
			return nil
		}
	}

	// Check case-insensitive match
	ruleNameLower := strings.ToLower(ruleName)
	for _, valid := range validRules {
		if strings.ToLower(valid) == ruleNameLower {
			return nil
		}
	}

	// Show only factory keys in error message (user-friendly)
	userFriendlyRules := []string{
		"subject", "conventional", "commitbody", "jirareference",
		"signoff", "signature", "identity", "spell", "branchahead",
	}

	return fmt.Errorf("unknown rule '%s'. Valid rules: %s", o.RuleHelp, strings.Join(userFriendlyRules, ", "))
}

// GetNormalizedRuleHelp returns the rule name normalized for comparison with actual rule names.
func (o OutputOptions) GetNormalizedRuleHelp() string {
	if o.RuleHelp == "" {
		return ""
	}

	// Map factory keys to actual rule names
	factoryToActual := map[string]string{
		"subject":       "Subject",
		"conventional":  "ConventionalCommit",
		"commitbody":    "CommitBody",
		"jirareference": "JiraReference",
		"signoff":       "SignOff",
		"signature":     "Signature",
		"identity":      "SignedIdentity",
		"spell":         "Spell",
		"branchahead":   "BranchAhead",
	}

	ruleName := strings.TrimSpace(o.RuleHelp)
	ruleNameLower := strings.ToLower(ruleName)

	// If it's a factory key, return the actual rule name
	if actual, exists := factoryToActual[ruleNameLower]; exists {
		return actual
	}

	// If it's already an actual rule name, return as-is
	actualRules := []string{
		"Subject", "ConventionalCommit", "CommitBody", "JiraReference",
		"SignOff", "Signature", "SignedIdentity", "Spell", "BranchAhead",
	}

	for _, actual := range actualRules {
		if strings.EqualFold(actual, ruleName) {
			return actual
		}
	}

	// Return original if no match found
	return ruleName
}

// getValidRuleNames returns all valid rule names (both factory keys and actual rule names).
func getValidRuleNames() []string {
	return []string{
		// Factory keys (for convenience)
		"subject",
		"conventional",
		"commitbody",
		"jirareference",
		"signoff",
		"signature",
		"identity",
		"spell",
		"branchahead",
		// Actual rule Name() return values
		"Subject",
		"ConventionalCommit",
		"CommitBody",
		"JiraReference",
		"SignOff",
		"Signature",
		"SignedIdentity",
		"Spell",
		"BranchAhead",
	}
}

// WithColor returns a new OutputOptions with the specified color setting.
func (o OutputOptions) WithColor(color string) OutputOptions {
	o.Color = color

	return o
}

// ShouldShowHelp returns true if help should be shown for all rules.
func (o OutputOptions) ShouldShowHelp() bool {
	return o.ShowHelp
}

// ShowRuleHelp returns true if help should be shown for a specific rule.
func (o OutputOptions) ShowRuleHelp() bool {
	return o.RuleHelp != ""
}

// GetRuleHelp returns the specific rule to show help for.
func (o OutputOptions) GetRuleHelp() string {
	return o.RuleHelp
}

// ShouldUseColor determines whether to use color based on the color setting and output type.
// Respects NO_COLOR environment variable as per https://no-color.org specification.
func (o OutputOptions) ShouldUseColor() bool {
	return shouldUseColorWithEnv(o.Color, o.Writer, getEnv)
}

// shouldUseColorWithEnv determines color usage with explicit environment function (pure function).
func shouldUseColorWithEnv(colorSetting string, writer io.Writer, envFunc func(string) string) bool {
	// NO_COLOR environment variable takes precedence over all flags
	// https://no-color.org - when set (any value), disable color output
	if envFunc("NO_COLOR") != "" {
		return false
	}

	switch colorSetting {
	case "always":
		return true
	case "never":
		return false
	case "auto":
		fallthrough
	default:
		return isTerminal(writer)
	}
}

// getEnv is a wrapper around os.Getenv for dependency injection in tests.
func getEnv(key string) string {
	return os.Getenv(key)
}

// isTerminal checks if the writer is a terminal using proper TTY detection.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}

	return false
}

// FormatReport formats a domain report using the specified options (pure function).
func (o OutputOptions) FormatReport(report domain.Report) string {
	switch o.Format {
	case "json":
		return output.JSON(report)
	case "github":
		return output.GitHub(report)
	case "gitlab":
		return output.GitLab(report)
	case "text":
		fallthrough
	default:
		textOptions := output.TextOptions{
			Verbose:      o.Verbose,
			VerboseLevel: o.VerboseLevel,
			ShowHelp:     o.ShouldShowHelp(),
			ShowRuleHelp: o.ShowRuleHelp(),
			RuleHelpName: o.GetNormalizedRuleHelp(),
			UseColor:     o.ShouldUseColor(),
		}

		return output.Text(report, textOptions)
	}
}

// WriteReport formats and writes a report to the configured writer.
func (o OutputOptions) WriteReport(report domain.Report) error {
	content := o.FormatReport(report)
	_, err := o.Writer.Write([]byte(content))

	return err
}

// ToReportOptions converts OutputOptions to domain.ReportOptions.
func (o OutputOptions) ToReportOptions() domain.ReportOptions {
	return domain.ReportOptions{
		Format:   o.Format,
		Verbose:  o.Verbose,
		ShowHelp: o.ShouldShowHelp(),
		UseColor: o.ShouldUseColor(),
		Writer:   o.Writer,
	}
}
