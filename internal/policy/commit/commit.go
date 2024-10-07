// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

// Package commit provides commit-related policies.
package commit

import (
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/git"
	"github.com/janderssonse/gommitlint/internal/policy"
)

// HeaderChecks is the configuration for checks on the header of a commit.
type HeaderChecks struct {
	// Length is the maximum length of the commit subject.
	Length int `mapstructure:"length"`
	// Imperative enforces the use of imperative verbs as the first word of a
	// commit message.
	Imperative bool `mapstructure:"imperative"`
	// HeaderCase is the case that the first word of the header must have ("upper" or "lower").
	Case string `mapstructure:"case"`
	// HeaderInvalidLastCharacters is a string containing all invalid last characters for the header.
	InvalidLastCharacters string `mapstructure:"invalidLastCharacters"`
	// Jira checks if the header containers a Jira project key.
	Jira *JiraChecks `mapstructure:"jira"`
}

// JiraChecks is the configuration for checks for Jira issues.
type JiraChecks struct {
	Keys []string `mapstructure:"keys"`
}

// BodyChecks is the configuration for checks on the body of a commit.
type BodyChecks struct {
	// Required enforces that the current commit has a body.
	Required bool `mapstructure:"required"`
}

// GPG is the configuration for checks GPG signature on the commit.
type GPG struct {
	// Required enforces that the current commit has a signature.
	Required bool `mapstructure:"required"`
	// Identity configures identity of the signature.
	Identity *struct {
		// GitHubOrganization enforces that commit should be signed with the key
		// of one of the organization public members.
		GitHubOrganization string `mapstructure:"gitHubOrganization"`
	} `mapstructure:"identity"`
}

// Commit implements the policy.Policy interface and enforces commit
// messages to gommitlint the Conventional Commit standard.
type Commit struct {
	// SpellCheck enforces correct spelling.
	SpellCheck *SpellCheck `mapstructure:"spellcheck"`
	// Conventional is the user specified settings for conventional commits.
	Conventional *Conventional `mapstructure:"conventional"`
	// Header is the user specified settings for the header of each commit.
	Header *HeaderChecks `mapstructure:"header"`
	// Header is the user specified settings for the body of each commit.
	Body *BodyChecks `mapstructure:"body"`
	// DCO enables the Developer Certificate of Origin check.
	DCO bool `mapstructure:"dco"`
	// GPG is the user specified settings for the GPG signature check.
	GPG *GPG `mapstructure:"gpg"`
	// GPGSignatureGitHubOrganization enforces that GPG signature should come from
	// one of the members of the GitHub org.
	GPGSignatureGitHubOrganization string `mapstructure:"gpgSignatureGitHubOrg"`
	// MaximumOfOneCommit enforces that the current commit is only one commit
	// ahead of a specified ref.
	MaximumOfOneCommit bool `mapstructure:"maximumOfOneCommit"`

	msg string
}

// FirstWordRegex is theregular expression used to find the first word in a
// commit.
var FirstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

// Compliance implements the policy.Policy.Compliance function.
func (commit *Commit) Compliance(options *policy.Options) (*policy.Report, error) {
	var err error

	report := &policy.Report{}

	// Setup the policy for all checks.
	var gitPtr *git.Git

	if gitPtr, err = git.NewGit(); err != nil {
		return report, errors.Errorf("failed to open git repo: %v", err)
	}

	var msgs []string

	switch option := options; {
	case option.CommitMsgFile != nil:
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
		commit.msg = msgs[index]

		commit.compliance(report, gitPtr, options)
	}

	return report, nil
}

// compliance checks the compliance with the policies of the given commit.
func (commit *Commit) compliance(report *policy.Report, gitPtr *git.Git, options *policy.Options) {
	if commit.Header != nil {
		if commit.Header.Length != 0 {
			report.AddCheck(commit.ValidateHeaderLength())
		}

		if commit.Header.Imperative {
			report.AddCheck(commit.ValidateImperative())
		}

		if commit.Header.Case != "" {
			report.AddCheck(commit.ValidateHeaderCase())
		}

		if commit.Header.InvalidLastCharacters != "" {
			report.AddCheck(commit.ValidateHeaderLastCharacter())
		}

		if commit.Header.Jira != nil {
			report.AddCheck(commit.ValidateJiraCheck())
		}
	}

	if commit.DCO {
		report.AddCheck(commit.ValidateDCO())
	}

	if commit.GPG != nil {
		if commit.GPG.Required {
			report.AddCheck(commit.ValidateGPGSign(gitPtr))

			if commit.GPG.Identity != nil {
				report.AddCheck(commit.ValidateGPGIdentity(gitPtr))
			}
		}
	}

	if commit.Conventional != nil {
		report.AddCheck(commit.ValidateConventionalCommit())
	}

	if commit.SpellCheck != nil {
		report.AddCheck(commit.ValidateSpelling())
	}

	if commit.MaximumOfOneCommit {
		report.AddCheck(commit.ValidateNumberOfCommits(gitPtr, options.CommitRef))
	}

	if commit.Body != nil {
		if commit.Body.Required {
			report.AddCheck(commit.ValidateBody())
		}
	}
}

func (commit Commit) firstWord() (string, error) {
	var (
		groups []string
		msg    string
	)

	if commit.Conventional != nil {
		groups = parseHeader(commit.msg)
		if len(groups) != 7 {
			return "", errors.Errorf("Invalid conventional commit format")
		}

		msg = groups[5]
	} else {
		msg = commit.msg
	}

	if msg == "" {
		return "", errors.Errorf("Invalid msg: %s", msg)
	}

	if groups = FirstWordRegex.FindStringSubmatch(msg); groups == nil {
		return "", errors.Errorf("Invalid msg: %s", msg)
	}

	return groups[0], nil
}

func (commit Commit) header() string {
	return strings.Split(strings.TrimPrefix(commit.msg, "\n"), "\n")[0]
}

func extractRevisionRange(options *policy.Options) ([]string, error) {
	revs := strings.Split(options.RevisionRange, "..")
	if len(revs) > 2 || len(revs) == 0 || revs[0] == "" || revs[1] == "" {
		return nil, errors.New("invalid revision range")
	} else if len(revs) == 1 {
		// if no final rev is given, use HEAD as default
		revs = append(revs, "HEAD")
	}

	return revs, nil
}
