// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"fmt"
	"strings"
)

func (v *Validator) parseRevisionRange() ([]string, error) {
	revs := strings.Split(v.options.RevisionRange, "..")
	if !v.isValidRevisionRange(revs) {
		return nil, fmt.Errorf("invalid revision range: %s", v.options.RevisionRange)
	}

	if len(revs) == 1 {
		revs = append(revs, "HEAD")
	}

	return revs, nil
}

func (v *Validator) isValidRevisionRange(revs []string) bool {
	return len(revs) > 0 && len(revs) <= 2 && revs[0] != "" && (len(revs) == 1 || revs[1] != "")
}
