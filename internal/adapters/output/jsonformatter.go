// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"encoding/json"
	"maps"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
)

// JSON formats a domain report as JSON (pure function).
func JSON(report domain.Report) string {
	output := map[string]interface{}{
		"timestamp":     report.Metadata.Timestamp.Format(time.RFC3339),
		"allPassed":     report.Summary.AllPassed,
		"totalCommits":  report.Summary.TotalCommits,
		"passedCommits": report.Summary.PassedCommits,
		"ruleSummary":   report.Summary.FailedRules,
		"commitResults": convertCommitsToJSON(report.Commits),
	}

	if len(report.Repository.RuleResults) > 0 {
		output["repositoryResults"] = convertRepositoryResultsToJSON(report.Repository.RuleResults)
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal JSON"}`
	}

	return string(jsonBytes)
}

func convertCommitsToJSON(commits []domain.CommitReport) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(commits))

	for _, commitReport := range commits {
		if commitReport.Commit.Hash == "" {
			continue
		}

		commit := map[string]interface{}{
			"hash":         commitReport.Commit.Hash,
			"subject":      commitReport.Commit.Subject,
			"passed":       commitReport.Passed,
			"ruleResults":  convertRulesToJSON(commitReport.RuleResults),
			"errorCount":   countErrors(commitReport.RuleResults),
			"warningCount": 0,
		}

		if commitReport.Commit.CommitDate != "" {
			commit["commitDate"] = commitReport.Commit.CommitDate
		} else {
			commit["commitDate"] = time.Now().Format(time.RFC3339)
		}

		if commitReport.Commit.Author != "" {
			authorInfo := commitReport.Commit.Author
			if commitReport.Commit.AuthorEmail != "" {
				authorInfo += " <" + commitReport.Commit.AuthorEmail + ">"
			}

			commit["author"] = authorInfo
		} else {
			commit["author"] = "Unknown"
		}

		results = append(results, commit)
	}

	return results
}

func convertRulesToJSON(rules []domain.RuleReport) []map[string]interface{} {
	results := make([]map[string]interface{}, len(rules))

	for i, ruleReport := range rules {
		results[i] = map[string]interface{}{
			"id":      ruleReport.Name,
			"name":    ruleReport.Name,
			"status":  string(ruleReport.Status),
			"message": ruleReport.Message,
			"errors":  convertErrorsToJSON(ruleReport.Errors),
		}
	}

	return results
}

func convertRepositoryResultsToJSON(repoResults []domain.RuleReport) []map[string]interface{} {
	results := make([]map[string]interface{}, len(repoResults))

	for i, ruleReport := range repoResults {
		results[i] = map[string]interface{}{
			"id":     ruleReport.Name,
			"name":   ruleReport.Name,
			"status": string(ruleReport.Status),
			"errors": convertErrorsToJSON(ruleReport.Errors),
		}
	}

	return results
}

func convertErrorsToJSON(validationErrors []domain.ValidationError) []map[string]interface{} {
	if len(validationErrors) == 0 {
		return nil
	}

	results := make([]map[string]interface{}, len(validationErrors))
	for idx, err := range validationErrors {
		results[idx] = map[string]interface{}{
			"rule":    err.Rule,
			"code":    err.Code,
			"message": err.Message,
			"context": maps.Clone(err.Context),
		}
		if err.Help != "" {
			results[idx]["help"] = err.Help
		}
	}

	return results
}

func countErrors(rules []domain.RuleReport) int {
	total := 0

	for _, rule := range rules {
		if rule.Status == domain.StatusFailed {
			total += len(rule.Errors)
		}
	}

	return total
}
