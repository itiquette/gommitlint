// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package configuration

import (
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type AppConf struct {
	GommitConf *Gommit `koanf:"gommitlint"`
}

// New loads the gommitlint.yaml file and unmarshals it into a Gommitlint struct.
func New() (*AppConf, error) {
	gommitLintConf, err := DefaultConfigLoader{}.LoadConfiguration()
	if err != nil {
		return nil, err
	}

	return gommitLintConf, nil
}

type Gommit struct {
	// Header is the user specified settings for the body of each commit.
	Body *BodyChecks `koanf:"body"`
	// Conventional is the user specified settings for conventional commits.
	Conventional *Conventional `koanf:"conventional"`
	// DCO enables the Developer Certificate of Origin check.
	DCO bool `koanf:"dco"`
	// GPG is the user specified settings for the GPG signature check.
	GPG *GPG `koanf:"gpg"`
	// Header is the user specified settings for the header of each commit.
	Header *HeaderChecks `koanf:"header"`
	// MaximumOfOneCommit enforces that the current commit is only one commit
	// ahead of a specified ref
	MaximumOfOneCommit bool `koanf:"maximumOfOneCommit"`
	// SpellCheck enforces correct spelling.
	SpellCheck *SpellCheck `koanf:"spellcheck"`

	Message string

	Signature string

	RawCommit *object.Commit
}

// HeaderChecks is the configuration for checks on the header of a commit.
type HeaderChecks struct {
	// Case is the case that the first word of the header must have ("upper" or "lower").
	Case string `koanf:"case"`
	// Imperative enforces the use of imperative verbs as the first word of a commit message.
	Imperative bool `koanf:"imperative"`
	// InvalidSuffix is a string containing all invalid last characters for the header.
	InvalidSuffix string `koanf:"invalidsuffix"`
	// Jira checks if the header containers a Jira project key.
	Jira *JiraChecks `koanf:"jira"`
	// Length is the maximum length of the commit subject.
	Length int `koanf:"length"`
}

type Conventional struct {
	DescriptionLength int      `koanf:"descriptionLength"`
	Scopes            []string `koanf:"scopes"`
	Types             []string `koanf:"types"`
}

// SpellCheck represents the locale to use for spell checking.
type SpellCheck struct {
	Locale string `koanf:"locale"`
}

// JiraChecks is the configuration for checks for Jira issues.
type JiraChecks struct {
	Keys []string `koanf:"keys"`
}

// BodyChecks is the configuration for checks on the body of a commit.
type BodyChecks struct {
	// Required enforces that the current commit has a body.
	Required bool `koanf:"required"`
}

// GPG is the configuration for checks GPG signature on the commit.
type GPG struct {
	// Identity configures identity of the signature.
	Identity *struct {
		// PublicKeyURI enforces that commit should be signed with the key
		// of one of the organization public members.
		PublicKeyURI string `koanf:"filepath"`
	} `koanf:"identity"`
	// Required enforces that the current commit has a signature.
	Required bool `koanf:"required"`
}

func (commit Gommit) HeaderFromMsg() string {
	return strings.Split(strings.TrimPrefix(commit.Message, "\n"), "\n")[0]
}
