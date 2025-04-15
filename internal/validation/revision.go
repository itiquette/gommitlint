// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal"
)

func (v *Validator) parseRevisionRange() ([]string, error) {
	revs := strings.Split(v.options.RevisionRange, "..")
	if !v.isValidRevisionRange(revs) {
		return nil, internal.NewInputError(
			fmt.Errorf("invalid revision range: %s", v.options.RevisionRange),
			map[string]string{"revision_range": v.options.RevisionRange})
	}

	if len(revs) == 1 {
		revs = append(revs, "HEAD")
	}

	return revs, nil
}

func (v *Validator) isValidRevisionRange(revs []string) bool {
	return len(revs) > 0 && len(revs) <= 2 && revs[0] != "" && (len(revs) == 1 || revs[1] != "")
}
