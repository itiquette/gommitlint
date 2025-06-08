// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	cliAdapter "github.com/itiquette/gommitlint/internal/adapters/cli"
)

// FindHookPath determines the Git hook file path for the given repository and hook type.
// This is a pure function that uses the provided validator or defaults to the standard one.
func FindHookPath(repoPath, hookType string, validator cliAdapter.PathValidator) (string, error) {
	if validator == nil {
		validator = cliAdapter.DefaultPathValidator()
	}

	return validator.ValidateHookPath(repoPath, hookType)
}

// EnsureHooksDirectory ensures the Git hooks directory exists for the given repository.
// This is a pure function that uses the provided validator or defaults to the standard one.
func EnsureHooksDirectory(repoPath string, validator cliAdapter.PathValidator) error {
	if validator == nil {
		validator = cliAdapter.DefaultPathValidator()
	}

	return validator.EnsureHooksDirectory(repoPath)
}
