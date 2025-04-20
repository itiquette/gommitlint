// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// setupGitRepo creates a real Git repository in the specified directory.
func setupGitRepo(dir string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = dir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Set git user configuration for the repository
	configCmd := exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = dir

	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user.name: %w", err)
	}

	configCmd = exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = dir

	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("failed to set git user.email: %w", err)
	}

	// Create a sample file
	sampleFile := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(sampleFile, []byte("Sample content"), 0600); err != nil {
		return fmt.Errorf("failed to create sample file: %w", err)
	}

	// Stage the file
	addCmd := exec.Command("git", "add", "sample.txt")
	addCmd.Dir = dir

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}

	// Commit the file
	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = dir

	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit file: %w", err)
	}

	return nil
}
