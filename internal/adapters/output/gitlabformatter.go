// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// GitLab formats a domain report for GitLab CI (pure function).
func GitLab(report domain.Report) string {
	var builder strings.Builder

	// GitLab CI section for summary
	builder.WriteString("section_start:$(date +%s):summary[collapsed=true]\n")
	builder.WriteString(fmt.Sprintf("Validated %d commits\n", report.Summary.TotalCommits))
	builder.WriteString(fmt.Sprintf("Passed: %d, Failed: %d\n", report.Summary.PassedCommits, report.Summary.FailedCommits))
	builder.WriteString("section_end:$(date +%s):summary\n")

	// Format each commit in its own section
	for idx, commitReport := range report.Commits {
		if commitReport.Commit.Hash == "" {
			continue
		}

		sectionName := fmt.Sprintf("commit_%d", idx+1)
		builder.WriteString("section_start:$(date +%s):" + sectionName + "[collapsed=true]\n")
		builder.WriteString(fmt.Sprintf("Commit #%d: %s\n", idx+1, commitReport.Commit.Hash))
		builder.WriteString(fmt.Sprintf("Subject: %s\n", commitReport.Commit.Subject))

		writeGitLabRules(&builder, commitReport)
		builder.WriteString("section_end:$(date +%s):" + sectionName + "\n")
	}

	// Format repository-level results
	if len(report.Repository.RuleResults) > 0 {
		builder.WriteString("section_start:$(date +%s):repository[collapsed=true]\n")
		builder.WriteString("Repository Validation\n")

		for _, repoResult := range report.Repository.RuleResults {
			if repoResult.Status == domain.StatusFailed {
				for _, err := range repoResult.Errors {
					builder.WriteString(fmt.Sprintf("ERROR: %s - %s\n",
						repoResult.Name, err.Message))
				}
			} else {
				builder.WriteString(fmt.Sprintf("✅ %s: passed\n", repoResult.Name))
			}
		}

		builder.WriteString("section_end:$(date +%s):repository\n")
	}

	return builder.String()
}

func writeGitLabRules(builder *strings.Builder, commitReport domain.CommitReport) {
	failedCount := 0

	for _, ruleReport := range commitReport.RuleResults {
		if ruleReport.Status == domain.StatusFailed {
			failedCount++

			for _, err := range ruleReport.Errors {
				hash := commitReport.Commit.Hash
				if len(hash) > 7 {
					hash = hash[:7]
				}

				builder.WriteString(fmt.Sprintf("ERROR: %s - %s: %s\n", hash, ruleReport.Name, err.Message))
			}
		}
	}

	if commitReport.Passed {
		builder.WriteString("✅ All rules passed\n")
	} else {
		builder.WriteString(fmt.Sprintf("❌ %d rules failed\n", failedCount))
	}
}
