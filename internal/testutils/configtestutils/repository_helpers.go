// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithRepository creates a new Config with the repository config.
func (t *TestUtils) WithRepository(cfg types.Config, repo types.RepositoryConfig) types.Config {
	result := cfg
	result.Repository = repo

	return result
}

// WithReferenceBranch returns a new RepositoryConfig with the updated reference branch.
func (t *TestUtils) WithReferenceBranch(c types.RepositoryConfig, branch string) types.RepositoryConfig {
	result := c
	result.ReferenceBranch = branch

	return result
}

// WithMaxCommitsAhead returns a new RepositoryConfig with the updated max commits ahead.
func (t *TestUtils) WithMaxCommitsAhead(c types.RepositoryConfig, maxCommits int) types.RepositoryConfig {
	result := c
	result.MaxCommitsAhead = maxCommits

	return result
}

// WithMaxHistoryDays returns a new RepositoryConfig with the updated max history days.
func (t *TestUtils) WithMaxHistoryDays(c types.RepositoryConfig, days int) types.RepositoryConfig {
	result := c
	result.MaxHistoryDays = days

	return result
}
