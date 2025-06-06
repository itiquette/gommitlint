// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
)

// GitHub formats a domain report for GitHub Actions (pure function).
func GitHub(report domain.Report) string {
	var builder strings.Builder

	// GitHub Actions group for summary
	builder.WriteString("::group::Summary\n")
	builder.WriteString(fmt.Sprintf("Validated %d commits\n", report.Summary.TotalCommits))
	builder.WriteString(fmt.Sprintf("Passed: %d, Failed: %d\n", report.Summary.PassedCommits, report.Summary.FailedCommits))
	builder.WriteString("::endgroup::\n")

	// Format each commit in its own group
	for i, commitReport := range report.Commits {
		if commitReport.Commit.Hash == "" {
			continue
		}

		builder.WriteString(fmt.Sprintf("::group::Commit #%d: %s\n", i+1, commitReport.Commit.Hash))
		builder.WriteString(fmt.Sprintf("Subject: %s\n", commitReport.Commit.Subject))

		writeGitHubRules(&builder, commitReport)
		builder.WriteString("::endgroup::\n")
	}

	// Format repository-level results
	if len(report.Repository.RuleResults) > 0 {
		builder.WriteString("::group::Repository Validation\n")

		for _, repoResult := range report.Repository.RuleResults {
			if repoResult.Status == domain.StatusFailed {
				for _, err := range repoResult.Errors {
					builder.WriteString(fmt.Sprintf("::error title=%s::%s\n",
						repoResult.Name, err.Message))
				}
			}
		}

		builder.WriteString("::endgroup::\n")
	}

	// Set GitHub Actions output
	if report.Summary.AllPassed {
		builder.WriteString("::set-output name=passed::true\n")
	} else {
		builder.WriteString("::set-output name=passed::false\n")
	}

	return builder.String()
}

func writeGitHubRules(builder *strings.Builder, commitReport domain.CommitReport) {
	failedCount := 0

	for _, ruleReport := range commitReport.RuleResults {
		if ruleReport.Status == domain.StatusFailed {
			failedCount++

			for _, err := range ruleReport.Errors {
				builder.WriteString(fmt.Sprintf("::error file=%s,line=1,title=%s::%s\n",
					commitReport.Commit.Hash, ruleReport.Name, err.Message))
			}
		}
	}

	if commitReport.Passed {
		builder.WriteString("✅ All rules passed\n")
	} else {
		builder.WriteString(fmt.Sprintf("❌ %d rules failed\n", failedCount))
	}
}
