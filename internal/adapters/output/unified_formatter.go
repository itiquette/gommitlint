// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package format

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/itiquette/gommitlint/internal/domain"
)

// OutputType represents different output format types.
type OutputType string

const (
	// OutputText is plain text format.
	OutputText OutputType = "text"
	// OutputJSON is JSON format.
	OutputJSON OutputType = "json"
	// OutputGitHub is GitHub Actions format.
	OutputGitHub OutputType = "github"
	// OutputGitLab is GitLab CI format.
	OutputGitLab OutputType = "gitlab"
)

// UnifiedFormatter handles all output formats in a single type.
type UnifiedFormatter struct {
	outputType OutputType
	verbose    bool
	showHelp   bool
	lightMode  bool
}

// NewUnifiedFormatter creates a new unified formatter.
func NewUnifiedFormatter(outputType OutputType) UnifiedFormatter {
	return UnifiedFormatter{
		outputType: outputType,
		verbose:    false,
		showHelp:   false,
		lightMode:  false,
	}
}

// WithVerbose sets verbose output.
func (f UnifiedFormatter) WithVerbose(verbose bool) UnifiedFormatter {
	f.verbose = verbose

	return f
}

// WithShowHelp enables help output.
func (f UnifiedFormatter) WithShowHelp(showHelp bool) UnifiedFormatter {
	f.showHelp = showHelp

	return f
}

// WithLightMode enables light mode colors.
func (f UnifiedFormatter) WithLightMode(lightMode bool) UnifiedFormatter {
	f.lightMode = lightMode

	return f
}

// Format formats validation results according to the configured output type.
func (f UnifiedFormatter) Format(_ context.Context, results interface{}) string {
	switch validationResults := results.(type) {
	case domain.ValidationResults:
		return f.formatResults(validationResults)
	case []domain.ValidationResult:
		return f.formatNewResults(validationResults)
	case domain.ValidationResult:
		return f.formatNewResults([]domain.ValidationResult{validationResults})
	default:
		return f.formatError("invalid results type")
	}
}

// formatResults handles ValidationResults type.
func (f UnifiedFormatter) formatResults(results domain.ValidationResults) string {
	switch f.outputType {
	case OutputJSON:
		return f.formatJSON(results)
	case OutputGitHub:
		return f.formatGitHub(results)
	case OutputGitLab:
		return f.formatGitLab(results)
	case OutputText:
		return f.formatText(results)
	default:
		return f.formatText(results)
	}
}

// formatNewResults handles the new simplified ValidationResult slice.
func (f UnifiedFormatter) formatNewResults(results []domain.ValidationResult) string {
	switch f.outputType {
	case OutputJSON:
		return f.formatNewJSON(results)
	case OutputGitHub:
		return f.formatNewGitHub(results)
	case OutputGitLab:
		return f.formatNewGitLab(results)
	case OutputText:
		return f.formatNewText(results)
	default:
		return f.formatNewText(results)
	}
}

// ContentType returns the MIME type for the configured output format.
func (f UnifiedFormatter) ContentType() string {
	switch f.outputType {
	case OutputJSON:
		return "application/json"
	case OutputText, OutputGitHub, OutputGitLab:
		return "text/plain"
	default:
		return "text/plain"
	}
}

// formatError formats error messages consistently across all output types.
func (f UnifiedFormatter) formatError(message string) string {
	switch f.outputType {
	case OutputJSON:
		return fmt.Sprintf(`{"error": "%s"}`, message)
	case OutputText, OutputGitHub, OutputGitLab:
		return "Error: " + message
	default:
		return "Error: " + message
	}
}

// formatText formats results as plain text.
func (f UnifiedFormatter) formatText(results domain.ValidationResults) string {
	var builder strings.Builder

	colors := f.getColorScheme()

	// Summary for multiple commits
	if len(results.Results) > 1 {
		if results.AllPassed() {
			builder.WriteString(colors.Success(fmt.Sprintf("SUCCESS: All %d commits passed validation\n\n", results.TotalCommits)))
		} else {
			builder.WriteString(colors.Warning(fmt.Sprintf("SUMMARY: %d of %d commits passed validation\n", results.PassedCommits, results.TotalCommits)))
			f.writeFailedRulesSummary(&builder, results, colors)
			builder.WriteString("\n")
		}
	}

	// Format each commit
	for i, commitResult := range results.Results {
		f.writeCommitHeader(&builder, commitResult, i, len(results.Results), colors)
		f.writeCommitRules(&builder, commitResult, colors)
	}

	// Format repository-level results at the end
	if len(results.RepositoryResults) > 0 {
		f.writeRepositoryResults(&builder, results.RepositoryResults, colors)
	}

	return builder.String()
}

// formatJSON formats results as JSON.
func (f UnifiedFormatter) formatJSON(results domain.ValidationResults) string {
	output := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"allPassed":     results.AllPassed(),
		"totalCommits":  results.TotalCommits,
		"passedCommits": results.PassedCommits,
		"ruleSummary":   results.GetFailedRules(),
		"commitResults": f.convertCommitsToJSON(results.Results),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return f.formatError("Failed to marshal JSON")
	}

	return string(jsonBytes)
}

// formatGitHub formats results for GitHub Actions.
func (f UnifiedFormatter) formatGitHub(results domain.ValidationResults) string {
	var builder strings.Builder

	builder.WriteString("::group::Summary\n")
	builder.WriteString(fmt.Sprintf("Validated %d commits\n", results.TotalCommits))
	builder.WriteString(fmt.Sprintf("Passed: %d, Failed: %d\n", results.PassedCommits, results.TotalCommits-results.PassedCommits))
	builder.WriteString("::endgroup::\n")

	for i, commitResult := range results.Results {
		builder.WriteString(fmt.Sprintf("::group::Commit #%d: %s\n", i+1, commitResult.Commit.Hash))
		builder.WriteString(fmt.Sprintf("Subject: %s\n", commitResult.Commit.Subject))

		f.writeCIRules(&builder, commitResult, "github")
		builder.WriteString("::endgroup::\n")
	}

	// Set GitHub Actions output
	if results.AllPassed() {
		builder.WriteString("::set-output name=passed::true\n")
	} else {
		builder.WriteString("::set-output name=passed::false\n")
	}

	return builder.String()
}

// formatGitLab formats results for GitLab CI.
func (f UnifiedFormatter) formatGitLab(results domain.ValidationResults) string {
	var builder strings.Builder

	builder.WriteString("section_start:$(date +%s):summary[collapsed=true]\n")
	builder.WriteString(fmt.Sprintf("Validated %d commits\n", results.TotalCommits))
	builder.WriteString(fmt.Sprintf("Passed: %d, Failed: %d\n", results.PassedCommits, results.TotalCommits-results.PassedCommits))
	builder.WriteString("section_end:$(date +%s):summary\n")

	for i, commitResult := range results.Results {
		sectionName := fmt.Sprintf("commit_%d", i+1)
		builder.WriteString("section_start:$(date +%s):" + sectionName + "[collapsed=true]\n")
		builder.WriteString(fmt.Sprintf("Commit #%d: %s\n", i+1, commitResult.Commit.Hash))
		builder.WriteString(fmt.Sprintf("Subject: %s\n", commitResult.Commit.Subject))

		f.writeCIRules(&builder, commitResult, "gitlab")
		builder.WriteString("section_end:$(date +%s):" + sectionName + "\n")
	}

	return builder.String()
}

// Helper methods for text formatting.
func (f UnifiedFormatter) getColorScheme() colorScheme {
	if f.lightMode {
		return colorScheme{
			Success: color.New(color.FgGreen).SprintFunc(),
			Warning: color.New(color.FgRed).SprintFunc(),
			Error:   color.New(color.FgRed, color.Bold).SprintFunc(),
			Header:  color.New(color.FgBlue, color.Bold).SprintFunc(),
			Bold:    color.New(color.Bold).SprintFunc(),
			Muted:   color.New(color.FgBlack).SprintFunc(),
		}
	}

	return colorScheme{
		Success: color.New(color.FgGreen).SprintFunc(),
		Warning: color.New(color.FgYellow).SprintFunc(),
		Error:   color.New(color.FgRed, color.Bold).SprintFunc(),
		Header:  color.New(color.FgCyan, color.Bold).SprintFunc(),
		Bold:    color.New(color.Bold).SprintFunc(),
		Muted:   color.New(color.FgHiBlack).SprintFunc(),
	}
}

func (f UnifiedFormatter) writeFailedRulesSummary(builder *strings.Builder, results domain.ValidationResults, colors colorScheme) {
	failedRules := results.GetFailedRules()
	if len(failedRules) == 0 {
		return
	}

	builder.WriteString(colors.Bold("Rule failures summary:\n"))

	// Sort rule names
	ruleNames := make([]string, 0, len(failedRules))
	for ruleName := range failedRules {
		ruleNames = append(ruleNames, ruleName)
	}

	sort.Strings(ruleNames)

	for _, ruleName := range ruleNames {
		count := failedRules[ruleName]
		builder.WriteString(fmt.Sprintf("  - %s: %d failure(s)\n", colors.Bold(ruleName), count))
	}
}

func (f UnifiedFormatter) writeCommitHeader(builder *strings.Builder, commitResult domain.CommitResult, index, totalCommits int, colors colorScheme) {
	if commitResult.Commit.Hash == "" {
		return
	}

	if totalCommits > 1 {
		builder.WriteString(colors.Header(fmt.Sprintf("COMMIT #%d:\n", index+1)))
	}

	divider := strings.Repeat("=", 80)
	builder.WriteString(colors.Header(divider) + "\n")

	shortSHA := commitResult.Commit.Hash
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}

	builder.WriteString(fmt.Sprintf("%s %s\n", colors.Header("COMMIT-SHA:"), colors.Bold(shortSHA)))
	builder.WriteString(fmt.Sprintf("%s %s\n", colors.Header("SUBJECT:"), commitResult.Commit.Subject))

	if commitResult.Commit.Message != "" {
		parts := strings.SplitN(commitResult.Commit.Message, "\n", 2)
		if len(parts) > 1 && parts[1] != "" {
			builder.WriteString(fmt.Sprintf("%s\n%s\n", colors.Header("MESSAGE:"), parts[1]))
		}
	}

	builder.WriteString(colors.Header(divider) + "\n\n")
}

func (f UnifiedFormatter) writeCommitRules(builder *strings.Builder, commitResult domain.CommitResult, colors colorScheme) {
	// Sort rules alphabetically
	sortedRules := make([]domain.RuleResult, len(commitResult.RuleResults))
	copy(sortedRules, commitResult.RuleResults)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].RuleName < sortedRules[j].RuleName
	})

	passedCount := 0

	for _, ruleResult := range sortedRules {
		if ruleResult.Status == domain.StatusPassed {
			passedCount++
			symbol := colors.Success("✓")
			message := Result(ruleResult.RuleName, ruleResult.Errors)
			builder.WriteString(fmt.Sprintf("%s %s: %s\n", symbol, colors.Bold(ruleResult.RuleName), message))
		} else if ruleResult.Status == domain.StatusFailed {
			symbol := colors.Error("✗")
			builder.WriteString(fmt.Sprintf("%s %s: \n", symbol, colors.Bold(ruleResult.RuleName)))

			for _, err := range ruleResult.Errors {
				message := err.Message
				if f.verbose && err.Code != "" {
					message = fmt.Sprintf("[%s] %s", err.Code, message)
				}

				builder.WriteString(fmt.Sprintf("    %s\n", message))
			}

			builder.WriteString("\n")
		}
	}

	// Summary line
	totalRules := len(sortedRules)
	if passedCount == totalRules {
		builder.WriteString(colors.Success(fmt.Sprintf("\nPASS: All %d rules passed\n\n", totalRules)))
	} else {
		builder.WriteString(colors.Warning(fmt.Sprintf("\nFAIL: %d of %d rules passed\n\n", passedCount, totalRules)))
	}
}

func (f UnifiedFormatter) writeCIRules(builder *strings.Builder, commitResult domain.CommitResult, ciType string) {
	failedCount := 0

	for _, ruleResult := range commitResult.RuleResults {
		if ruleResult.Status == domain.StatusFailed {
			failedCount++

			for _, err := range ruleResult.Errors {
				if ciType == "github" {
					builder.WriteString(fmt.Sprintf("::error file=%s,line=1,title=%s::%s\n",
						commitResult.Commit.Hash, ruleResult.RuleName, err.Message))
				} else {
					builder.WriteString(fmt.Sprintf("ERROR: %s:1:%s - %s\n",
						commitResult.Commit.Hash, ruleResult.RuleName, err.Message))
				}
			}
		}
	}

	if commitResult.Passed {
		builder.WriteString("✅ All rules passed\n")
	} else {
		builder.WriteString(fmt.Sprintf("❌ %d rules failed\n", failedCount))
	}
}

func (f UnifiedFormatter) writeRepositoryResults(builder *strings.Builder, repoResults []domain.RuleResult, colors colorScheme) {
	builder.WriteString("\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString("REPOSITORY VALIDATION\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString("\n")

	for _, ruleResult := range repoResults {
		var icon string

		status := "No errors"

		if ruleResult.Status == domain.StatusFailed {
			icon = colors.Error("✗")

			if len(ruleResult.Errors) > 0 {
				status = ruleResult.Errors[0].Message
			}
		} else {
			icon = colors.Success("✓")
		}

		builder.WriteString(fmt.Sprintf("%s %s: ", icon, ruleResult.RuleName))

		if ruleResult.Status == domain.StatusFailed {
			builder.WriteString("\n")

			for _, err := range ruleResult.Errors {
				builder.WriteString(fmt.Sprintf("    %s\n", err.Message))
			}
		} else {
			builder.WriteString(status + "\n")
		}

		if f.showHelp && len(ruleResult.Errors) > 0 && ruleResult.Errors[0].Help != "" {
			builder.WriteString(fmt.Sprintf("    Help: %s\n", ruleResult.Errors[0].Help))
		}

		builder.WriteString("\n")
	}
}

func (f UnifiedFormatter) convertCommitsToJSON(commits []domain.CommitResult) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(commits))

	for _, commitResult := range commits {
		if commitResult.Commit.Hash == "" {
			continue
		}

		commit := map[string]interface{}{
			"hash":         commitResult.Commit.Hash,
			"subject":      commitResult.Commit.Subject,
			"passed":       commitResult.Passed,
			"ruleResults":  f.convertRulesToJSON(commitResult.RuleResults),
			"errorCount":   f.countErrors(commitResult.RuleResults),
			"warningCount": 0,
		}

		if commitResult.Commit.CommitDate != "" {
			commit["commitDate"] = commitResult.Commit.CommitDate
		} else {
			commit["commitDate"] = time.Now().Format(time.RFC3339)
		}

		if commitResult.Commit.Author != "" {
			authorInfo := commitResult.Commit.Author
			if commitResult.Commit.AuthorEmail != "" {
				authorInfo += " <" + commitResult.Commit.AuthorEmail + ">"
			}

			commit["author"] = authorInfo
		} else {
			commit["author"] = "Unknown"
		}

		results = append(results, commit)
	}

	return results
}

func (f UnifiedFormatter) convertRulesToJSON(rules []domain.RuleResult) []map[string]interface{} {
	results := make([]map[string]interface{}, len(rules))

	for i, ruleResult := range rules {
		results[i] = map[string]interface{}{
			"id":             ruleResult.RuleName,
			"name":           ruleResult.RuleName,
			"status":         string(ruleResult.Status),
			"message":        Result(ruleResult.RuleName, ruleResult.Errors),
			"verboseMessage": VerboseResult(ruleResult.RuleName, ruleResult.Errors),
			"help":           Help(ruleResult.RuleName, ruleResult.Errors),
			"errors":         f.convertErrorsToJSON(ruleResult.Errors),
		}
	}

	return results
}

func (f UnifiedFormatter) convertErrorsToJSON(validationErrors []domain.ValidationError) []map[string]interface{} {
	if len(validationErrors) == 0 {
		return nil
	}

	results := make([]map[string]interface{}, len(validationErrors))
	for i, err := range validationErrors {
		results[i] = map[string]interface{}{
			"rule":    err.Rule,
			"code":    err.Code,
			"message": err.Message,
			"context": maps.Clone(err.Context),
		}
	}

	return results
}

func (f UnifiedFormatter) countErrors(rules []domain.RuleResult) int {
	total := 0

	for _, rule := range rules {
		if rule.Status == domain.StatusFailed {
			total += len(rule.Errors)
		}
	}

	return total
}

// colorScheme provides color functions for different output elements.
type colorScheme struct {
	Success func(a ...interface{}) string
	Warning func(a ...interface{}) string
	Error   func(a ...interface{}) string
	Header  func(a ...interface{}) string
	Bold    func(a ...interface{}) string
	Muted   func(a ...interface{}) string
}

// Helper functions for formatting (moved from format_helpers.go)

// Result formats a rule result into a simple message.
func Result(_ string, errors []domain.ValidationError) string {
	if len(errors) == 0 {
		return "No errors"
	}

	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Message)
	}

	return strings.Join(messages, "; ")
}

// VerboseResult formats a rule result with more detail.
func VerboseResult(_ string, errors []domain.ValidationError) string {
	if len(errors) == 0 {
		return "No errors"
	}

	parts := make([]string, 0, len(errors))

	for _, err := range errors {
		msg := err.Message
		if err.Code != "" {
			msg = fmt.Sprintf("[%s] %s", err.Code, msg)
		}

		if len(err.Context) > 0 {
			var contextParts []string
			for k, v := range err.Context {
				contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
			}

			msg = fmt.Sprintf("%s (%s)", msg, strings.Join(contextParts, ", "))
		}

		parts = append(parts, msg)
	}

	return strings.Join(parts, "; ")
}

// Help extracts help messages from validation errors.
func Help(_ string, errors []domain.ValidationError) string {
	var helps []string

	seen := make(map[string]bool)

	for _, err := range errors {
		if err.Help != "" && !seen[err.Help] {
			helps = append(helps, err.Help)
			seen[err.Help] = true
		}
	}

	return strings.Join(helps, "; ")
}

// formatNewText formats new ValidationResult slice as plain text.
func (f UnifiedFormatter) formatNewText(results []domain.ValidationResult) string {
	var builder strings.Builder

	// Summary for multiple commits
	if len(results) > 1 {
		passed := 0

		for _, result := range results {
			if result.Passed() {
				passed++
			}
		}

		if passed == len(results) {
			fmt.Fprintf(&builder, "SUCCESS: All %d commits passed validation\n\n", len(results))
		} else {
			fmt.Fprintf(&builder, "SUMMARY: %d of %d commits passed validation\n\n", passed, len(results))
		}
	}

	// Format each commit
	for i, result := range results {
		if len(results) > 1 {
			fmt.Fprintf(&builder, "COMMIT #%d:\n", i+1)
		}

		fmt.Fprintf(&builder, "================================================================================\n")

		hash := result.Commit.Hash
		if hash == "" {
			hash = "0000000"
		}

		if len(hash) > 7 {
			hash = hash[:7]
		}

		fmt.Fprintf(&builder, "COMMIT-SHA: %s\n", hash)
		fmt.Fprintf(&builder, "SUBJECT: %s\n", result.Commit.Subject)
		fmt.Fprintf(&builder, "================================================================================\n\n")

		if result.Passed() {
			fmt.Fprintf(&builder, "✓ All rules passed\n\n")
		} else {
			for _, failure := range result.Failures {
				fmt.Fprintf(&builder, "✗ %s: %s\n", failure.Rule, failure.Message)

				if failure.Help != "" {
					fmt.Fprintf(&builder, "    → %s\n", failure.Help)
				}

				fmt.Fprintln(&builder)
			}
		}
	}

	return builder.String()
}

// formatNewJSON formats new ValidationResult slice as JSON.
func (f UnifiedFormatter) formatNewJSON(results []domain.ValidationResult) string {
	return ConvertValidationResultsToJSON(results)
}

// formatNewGitHub formats new ValidationResult slice for GitHub Actions.
func (f UnifiedFormatter) formatNewGitHub(results []domain.ValidationResult) string {
	var builder strings.Builder

	for _, result := range results {
		if !result.Passed() {
			for _, failure := range result.Failures {
				fmt.Fprintf(&builder, "::error file=%s,title=%s::%s\n",
					result.Commit.Hash, failure.Rule, failure.Message)
			}
		}
	}

	return builder.String()
}

// formatNewGitLab formats new ValidationResult slice for GitLab CI.
func (f UnifiedFormatter) formatNewGitLab(results []domain.ValidationResult) string {
	var builder strings.Builder

	for _, result := range results {
		if !result.Passed() {
			for _, failure := range result.Failures {
				hash := result.Commit.Hash
				if len(hash) > 7 {
					hash = hash[:7]
				}

				fmt.Fprintf(&builder, "ERROR: %s - %s: %s\n", hash, failure.Rule, failure.Message)
			}
		}
	}

	return builder.String()
}

// Ensure UnifiedFormatter implements Formatter.
var _ Formatter = (*UnifiedFormatter)(nil)
