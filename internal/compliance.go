// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package internal

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/itiquette/gommitlint/internal/configuration"
	"github.com/itiquette/gommitlint/internal/git"
	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/rules"
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

func Compliance(options *model.Options, gommit *configuration.Gommit) (*model.Report, error) {
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
		gommit.Message = msgs[index]

		compliance(report, gitPtr, options, gommit)
	}

	return report, nil
}

// compliance checks the compliance with the policies of the given commit.
func compliance(report *model.Report, gitPtr *git.Git, options *model.Options, gommit *configuration.Gommit) {
	if gommit.Header != nil {
		if gommit.Header.Length != 0 {
			actualLength := len(gommit.HeaderFromMsg())
			report.AddCheck(rules.ValidateHeaderLength(gommit.Header.Length, actualLength))
		}

		if gommit.Header.Imperative {
			isConventional := false
			if gommit.Conventional != nil {
				isConventional = true
			}

			report.AddCheck(rules.ValidateImperative(isConventional, gommit.Message))
		}

		if gommit.Header.Case != "" {
			isConventional := false
			if gommit.Conventional != nil {
				isConventional = true
			}

			report.AddCheck(rules.ValidateHeaderCase(isConventional, gommit.Message, gommit.Header.Case))
		}

		if gommit.Header.InvalidSuffix != "" {
			report.AddCheck(rules.ValidateHeaderSuffix(gommit.HeaderFromMsg(), gommit.Header.InvalidSuffix))
		}

		if gommit.Header.Jira != nil {
			report.AddCheck(rules.ValidateJiraCheck(gommit.Message, gommit.Header.Jira.Keys))
		}
	}

	if gommit.DCO {
		report.AddCheck(rules.ValidateDCO(gommit.Message))
	}

	if gommit.GPG != nil {
		if gommit.GPG.Required {
			report.AddCheck(rules.ValidateGPGSign(gitPtr))

			if gommit.GPG.Identity != nil {
				report.AddCheck(rules.ValidateGPGIdentity(gitPtr, gommit.GPG.Identity.GitHubOrganization))
			}
		}
	}

	if gommit.Conventional != nil {
		report.AddCheck(rules.ValidateConventionalCommit(gommit.Message, gommit.Conventional.Types, gommit.Conventional.Scopes, gommit.Conventional.DescriptionLength))
	}

	if gommit.SpellCheck != nil {
		report.AddCheck(rules.ValidateSpelling(gommit.Message, gommit.SpellCheck.Locale))
	}

	if gommit.MaximumOfOneCommit {
		report.AddCheck(rules.ValidateNumberOfCommits(gitPtr, options.CommitRef))
	}

	if gommit.Body != nil {
		if gommit.Body.Required {
			report.AddCheck(rules.ValidateBody(gommit.Message))
		}
	}
}

func extractRevisionRange(options *model.Options) ([]string, error) {
	revs := strings.Split(options.RevisionRange, "..")
	if len(revs) > 2 || len(revs) == 0 || revs[0] == "" || revs[1] == "" {
		return nil, errors.New("invalid revision range")
	} else if len(revs) == 1 {
		// if no final rev is given, use HEAD as default
		revs = append(revs, "HEAD")
	}

	return revs, nil
}
