// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/git"
	"github.com/janderssonse/gommitlint/internal/policy"
)

// NumberOfCommits enforces a maximum number of charcters on the commit
// header.
type NumberOfCommits struct {
	ref    string
	ahead  int
	errors []error
}

// Name returns the name of the check.
func (h NumberOfCommits) Name() string {
	return "Number of Commits"
}

// Message returns to check message.
func (h NumberOfCommits) Message() string {
	if len(h.errors) != 0 {
		return h.errors[0].Error()
	}

	return fmt.Sprintf("HEAD is %d commit(s) ahead of %s", h.ahead, h.ref)
}

// Errors returns any violations of the check.
func (h NumberOfCommits) Errors() []error {
	return h.errors
}

// ValidateNumberOfCommits checks the header length.
func (commit Commit) ValidateNumberOfCommits(gitPtr *git.Git, ref string) policy.Check { //nolint:ireturn
	check := &NumberOfCommits{
		ref: ref,
	}

	var err error

	check.ahead, _, err = gitPtr.AheadBehind(ref)
	if err != nil {
		check.errors = append(check.errors, err)

		return check
	}

	if check.ahead > 1 {
		check.errors = append(check.errors, errors.Errorf("HEAD is %d commit(s) ahead of %s", check.ahead, ref))

		return check
	}

	return check
}
