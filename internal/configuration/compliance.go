// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/janderssonse/gommitlint/internal/git"
	"github.com/janderssonse/gommitlint/internal/interfaces"
	"github.com/janderssonse/gommitlint/internal/model"
	"github.com/janderssonse/gommitlint/internal/rules"
	"github.com/pkg/errors"
)

// Validate enforces all policies defined in the gommitlint.yaml file.
func Validate(checks []interfaces.Check) error {
	const padding = 8
	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	fmt.Fprintln(tabWriter, "CHECK\tSTATUS\tMESSAGE\t")

	pass := true

	for _, check := range checks {
		if len(check.Errors()) != 0 {
			for _, err := range check.Errors() {
				fmt.Fprintf(tabWriter, "%s\t%s\t%v\t\n", check.Status(), "FAILED", err)
			}

			pass = false
		} else {
			fmt.Fprintf(tabWriter, "%s\t%s\t%s\t\n", check.Status(), "PASS", check.Message())
		}
	}

	tabWriter.Flush()

	if !pass {
		return errors.New("1 or more rules failed")
	}

	return nil
}

func Compliance(options *Options, commit *Gommit) (*model.Report, error) {
	var err error

	report := &model.Report{}

	var gitPtr *git.Git

	if gitPtr, err = git.NewGit(); err != nil {
		return report, errors.Errorf("failed to open git repo: %v", err)
	}

	var msgs []string

	switch option := options; {
	case option.CommitMsgFile != nil:
		fmt.Println("C")

		var contents []byte

		if contents, err = os.ReadFile(*options.CommitMsgFile); err != nil {
			return report, errors.Errorf("failed to read commit message file: %v", err)
		}

		msgs = append(msgs, string(contents))
	case option.RevisionRange != "":
		revs, err := extractRevisionRange(options)
		if err != nil {
			return report, errors.Errorf("failed to get commit message: %v", err)
		}

		msgs, err = gitPtr.Messages(revs[0], revs[1])
		if err != nil {
			return report, errors.Errorf("failed to get commit message: %v", err)
		}
	default:
		msg, err := gitPtr.Message()
		if err != nil {
			return report, errors.Errorf("failed to get commit message: %v", err)
		}

		msgs = append(msgs, msg)
	}

	for index := range msgs {
		commit.Message = msgs[index]

		compliance(report, gitPtr, options, commit)
	}

	return report, nil
}

// compliance checks the compliance with the policies of the given commit.
func compliance(report *model.Report, gitPtr *git.Git, options *Options, commit *Gommit) {
	if commit.Header != nil {
		if commit.Header.Length != 0 {
			actualLength := len(commit.HeaderFromMsg())
			report.AddCheck(rules.ValidateHeaderLength(commit.Header.Length, actualLength))
		}

		if commit.Header.Imperative {
			isConventional := false
			if commit.Conventional != nil {
				isConventional = true
			}

			report.AddCheck(rules.ValidateImperative(isConventional, commit.Message))
		}

		if commit.Header.Case != "" {
			isConventional := false
			if commit.Conventional != nil {
				isConventional = true
			}

			report.AddCheck(rules.ValidateHeaderCase(isConventional, commit.Message, commit.Header.Case))
		}

		if commit.Header.InvalidSuffix != "" {
			report.AddCheck(rules.ValidateHeaderSuffix(commit.HeaderFromMsg(), commit.Header.InvalidSuffix))
		}

		if commit.Header.Jira != nil {
			report.AddCheck(rules.ValidateJiraCheck(commit.Message, commit.Header.Jira.Keys))
		}
	}

	if commit.DCO {
		report.AddCheck(rules.ValidateDCO(commit.Message))
	}

	if commit.GPG != nil {
		if commit.GPG.Required {
			report.AddCheck(rules.ValidateGPGSign(gitPtr))

			if commit.GPG.Identity != nil {
				report.AddCheck(rules.ValidateGPGIdentity(gitPtr, commit.GPG.Identity.GitHubOrganization))
			}
		}
	}

	if commit.Conventional != nil {
		report.AddCheck(rules.ValidateConventionalCommit(commit.Message, commit.Conventional.Types, commit.Conventional.Scopes, commit.Conventional.DescriptionLength))
	}

	if commit.SpellCheck != nil {
		report.AddCheck(rules.ValidateSpelling(commit.Message, commit.SpellCheck.Locale))
	}

	if commit.MaximumOfOneCommit {
		report.AddCheck(rules.ValidateNumberOfCommits(gitPtr, options.CommitRef))
	}

	if commit.Body != nil {
		if commit.Body.Required {
			report.AddCheck(rules.ValidateBody(commit.Message))
		}
	}
}

func extractRevisionRange(options *Options) ([]string, error) {
	revs := strings.Split(options.RevisionRange, "..")
	if len(revs) > 2 || len(revs) == 0 || revs[0] == "" || revs[1] == "" {
		return nil, errors.New("invalid revision range")
	} else if len(revs) == 1 {
		// if no final rev is given, use HEAD as default
		revs = append(revs, "HEAD")
	}

	return revs, nil
}
