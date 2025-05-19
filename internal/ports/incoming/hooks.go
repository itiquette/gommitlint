// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package incoming

import "context"

// HookService defines the interface for git hook management.
type HookService interface {
	// InstallHook installs the commit-msg hook
	InstallHook(ctx context.Context, repoPath string) error

	// RemoveHook removes the commit-msg hook
	RemoveHook(ctx context.Context, repoPath string) error
}
