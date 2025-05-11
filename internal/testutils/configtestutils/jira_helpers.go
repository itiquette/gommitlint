// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configtestutils

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

// WithPattern returns a new JiraConfig with the updated pattern.
func (t *TestUtils) WithPattern(c types.JiraConfig, pattern string) types.JiraConfig {
	result := c
	result.Pattern = pattern

	return result
}

// WithProjects returns a new JiraConfig with the updated projects.
func (t *TestUtils) WithProjects(c types.JiraConfig, projects []string) types.JiraConfig {
	result := c
	result.Projects = projects

	return result
}

// WithBodyRef returns a new JiraConfig with the updated body reference setting.
func (t *TestUtils) WithBodyRef(c types.JiraConfig, bodyRef bool) types.JiraConfig {
	result := c
	result.BodyRef = bodyRef

	return result
}
