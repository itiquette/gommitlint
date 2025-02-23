// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

// package internal

// import (
// 	"os"
// 	"strings"

// 	"github.com/itiquette/gommitlint/internal/configuration"
// 	"github.com/itiquette/gommitlint/internal/git"
// 	"github.com/itiquette/gommitlint/internal/model"
// 	"github.com/itiquette/gommitlint/internal/rule"
// 	"github.com/pkg/errors"
// )

// func Validate(options *model.Options, gommit *configuration.Gommit) (*model.Report, error) {
// 	var err error

// 	report := &model.Report{}

// 	var gitPtr *git.Git

// 	if gitPtr, err = git.NewGit(); err != nil {
// 		return report, errors.Errorf("failed to open git repo: %v", err)
// 	}

// 	var msgs []git.CommitInfo

// 	switch option := options; {
// 	case option.CommitMsgFile != nil:
// 		var contents []byte

// 		if contents, err = os.ReadFile(*options.CommitMsgFile); err != nil {
// 			return report, errors.Errorf("failed to read commit message file: %v", err)
// 		}

// 		msgs = append(msgs, git.CommitInfo{Message: string(contents)})
// 	case option.RevisionRange != "":
// 		revs, err := extractRevisionRange(options)
// 		if err != nil {
// 			return report, errors.Errorf("failed to get commit message: %v", err)
// 		}

// 		msgs, err = gitPtr.Messages(revs[0], revs[1])
// 		if err != nil {
// 			return report, errors.Errorf("failed to get commit message: %v", err)
// 		}
// 	default:
// 		msg, err := gitPtr.Message()
// 		if err != nil {
// 			return report, errors.Errorf("failed to get commit message: %v", err)
// 		}

// 		msgs = append(msgs, msg)
// 	}

// 	for index := range msgs {
// 		gommit.Message = msgs[index].Message

// 		compliance(report, gitPtr, options, gommit)
// 	}

// 	return report, nil
// }

// // compliance checks the compliance with the policies of the given commit.
// func compliance(report *model.Report, gitPtr *git.Git, options *model.Options, gommit *configuration.Gommit) {
// 	if gommit.Header != nil {
// 		if gommit.Header.Length != 0 {
// 			report.AddRule(rule.ValidateHeaderLength(gommit.Message, gommit.Header.Length))
// 		}

// 		if gommit.Header.Imperative {
// 			isConventional := false
// 			if gommit.Conventional != nil {
// 				isConventional = true
// 			}

// 			report.AddRule(rule.ValidateImperative(isConventional, gommit.Message))
// 		}

// 		if gommit.Header.Case != "" {
// 			isConventional := false
// 			if gommit.Conventional != nil {
// 				isConventional = true
// 			}

// 			report.AddRule(rule.ValidateHeaderCase(isConventional, gommit.Message, gommit.Header.Case))
// 		}

// 		if gommit.Header.InvalidSuffix != "" {
// 			report.AddRule(rule.ValidateHeaderSuffix(gommit.HeaderFromMsg(), gommit.Header.InvalidSuffix))
// 		}

// 		if gommit.Header.Jira != nil {
// 			isConventional := false
// 			if gommit.Conventional != nil {
// 				isConventional = true
// 			}

// 			report.AddRule(rule.ValidateJira(gommit.Message, gommit.Header.Jira.Keys, isConventional))
// 		}
// 	}

// 	if gommit.DCO {
// 		report.AddRule(rule.ValidateSignOff(gommit.Message))
// 	}

// 	if gommit.GPG != nil {
// 		if gommit.GPG.Required {
// 			report.AddRule(rule.ValidateSignature(gitPtr))

// 			if gommit.GPG.Identity != nil {
// 				report.AddRule(rule.ValidateGPGIdentity(gommit.Signature, gommit.RawCommit, gommit.GPG.Identity.PublicKeyURI))
// 			}
// 		}
// 	}

// 	if gommit.Conventional != nil {
// 		report.AddRule(rule.ValidateConventionalCommit(gommit.Message, gommit.Conventional.Types, gommit.Conventional.Scopes, gommit.Conventional.DescriptionLength))
// 	}

// 	if gommit.SpellCheck != nil {
// 		report.AddRule(rule.ValidateSpelling(gommit.Message, gommit.SpellCheck.Locale))
// 	}

// 	if gommit.MaximumOfOneCommit {
// 		report.AddRule(rule.ValidateNumberOfCommits(gitPtr, options.CommitRef))
// 	}

// 	if gommit.Body != nil {
// 		if gommit.Body.Required {
// 			report.AddRule(rule.ValidateCommitBody(gommit.Message))
// 		}
// 	}
// }

// func extractRevisionRange(options *model.Options) ([]string, error) {
// 	revs := strings.Split(options.RevisionRange, "..")
// 	if len(revs) > 2 || len(revs) == 0 || revs[0] == "" || revs[1] == "" {
// 		return nil, errors.New("invalid revision range")
// 	} else if len(revs) == 1 {
// 		// if no final rev is given, use HEAD as default
// 		revs = append(revs, "HEAD")
// 	}

//		return revs, nil
//	}
package internal
