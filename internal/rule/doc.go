// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package rule provides a comprehensive set of commit message and Git state validation
rules for gommitlint.

Each rule implements a specific validation strategy used to enforce project-specific
standards and best practices for Git commits. Rules can be used individually or
combined to create a custom validation pipeline.

Most rules implement the following common interface:
  - Name(): returns the rule identifier
  - Result(): returns a human-readable validation result
  - Errors(): returns any validation errors
  - Help(): returns guidance on how to fix violations

The package contains several categories of rules:

Commit Message Content Rules:

  - ConventionalCommit: Enforces the Conventional Commits specification format
    (https://www.conventionalcommits.org/) with optional type and scope restrictions.

  - ImperativeVerb: Validates that commit messages begin with a verb in the
    imperative mood, following Git conventions.

  - SignOff: Ensures commits include a valid Developer Certificate of Origin (DCO)
    sign-off line.

  - Spell: Checks for common misspellings in commit messages with locale-specific
    dictionaries.

  - SubjectCase: Enforces consistent capitalization of the first word in commit
    subjects.

  - SubjectLength: Limits the character length of commit subject lines to improve
    readability.

  - SubjectSuffix: Prevents commit subjects from ending with specified characters
    like periods or commas.

  - JiraReference: Validates that commits reference Jira issue keys in a consistent
    format, with optional project validation.

Security Rules:

  - Signature: Verifies commits have a cryptographic signature (GPG or SSH).

  - SignedIdentity: Validates commit signatures against trusted keys with
    cryptographic verification.

Git State Rules:

  - CommitsAhead: Limits how far a branch can diverge from a reference branch to
    reduce merge complexity.

Each rule provides detailed help and error messages designed to guide users toward
fixing issues in their commit messages or repository state. The error messages
include examples and step-by-step instructions for resolving the most common
validation problems.

Example Usage:

	repo := openRepository(".")
	message := "feat: add new feature"

	// Validate conventional commit format
	conventionalRule := rule.ValidateConventionalCommit(message, []string{"feat", "fix"}, nil, 0)
	if len(conventionalRule.Errors()) > 0 {
	    fmt.Println(conventionalRule.Help())
	}

	// Validate commits aren't too far ahead of main
	commitsAheadRule := rule.ValidateNumberOfCommits(repo, "main")
	if len(commitsAheadRule.Errors()) > 0 {
	    fmt.Println(commitsAheadRule.Help())
	}
*/
package rule
