// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package commit

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/janderssonse/gommitlint/internal/policy"
)

// DCORegex is the regular expression used for Developer Certificate of Origin.
var DCORegex = regexp.MustCompile(`^Signed-off-by: ([^<]+) <([^<>@]+@[^<>]+)>$`)

// DCOCheck ensures that the commit message contains a
// Developer Certificate of Origin.
type DCOCheck struct {
	errors []error
}

// Name returns the name of the check.
func (d DCOCheck) Name() string {
	return "DCO"
}

// Message returns to check message.
func (d DCOCheck) Message() string {
	if len(d.errors) != 0 {
		return d.errors[0].Error()
	}

	return "Developer Certificate of Origin was found"
}

// Errors returns any violations of the check.
func (d DCOCheck) Errors() []error {
	return d.errors
}

// ValidateDCO checks the commit message for a Developer Certificate of Origin.
func (commit Commit) ValidateDCO() policy.Check { //nolint:ireturn
	check := &DCOCheck{}

	for _, line := range strings.Split(commit.msg, "\n") {
		if DCORegex.MatchString(strings.TrimSpace(line)) {
			return check
		}
	}

	check.errors = append(check.errors, errors.Errorf("Commit does not have a DCO"))

	return check
}
