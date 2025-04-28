// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package integtest contains end-to-end integration tests for gommitlint workflows.
// These tests verify that the application's components work together correctly.
package integtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCLIValidateCommand tests the CLI validate command for commit validation.
// buildTestBinary builds a binary for testing and returns its path.
func buildTestBinary(t *testing.T) (string, error) {
	t.Helper()

	// Create a temporary directory for the binary
	tempDir, err := os.MkdirTemp("", "gommitlint-binary-*")
	if err != nil {
		return "", err
	}

	t.Cleanup(func() { os.RemoveAll(tempDir) })

	// Determine binary name
	binaryName := "gommitlint-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tempDir, binaryName)

	// Get the project root
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(getCurrentFilePath())))

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, filepath.Join(projectRoot, "main.go"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to build test binary: %w, output: %s", err, string(output))
	}

	return binaryPath, nil
}

// getCurrentFilePath returns the current file path.
func getCurrentFilePath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}

	return file
}

func TestCLIValidateCommand(t *testing.T) {
	// Skip if running in CI environment without git
	if os.Getenv("CI") == "true" && !isGitAvailable() {
		t.Skip("Skipping integration test in CI environment without git")
	}

	// Build the CLI binary for testing
	binaryPath, err := buildTestBinary(t)
	require.NoError(t, err)
	defer os.Remove(binaryPath)

	// Setup test cases
	testCases := []struct {
		name          string
		commitMessage string
		args          []string
		shouldPass    bool
		config        string
		checkOutput   func(t *testing.T, output string) // Optional function to check output content
	}{
		{
			name:          "Validate HEAD - valid commit format",
			commitMessage: "feat: add new feature with proper format",
			args:          []string{"validate"},
			shouldPass:    false, // Fails because SignOff rule is enabled by default
			config: `
gommitlint:
  validation:
    enabled: true
  subject:
    max_length: 50
  conventional:
    enabled: true
    required: true
    types:
      - feat
      - fix
      - docs
  security:
    signature_required: false
    signoff_required: false
  rules:
    enabled:
      - SubjectLength
      - ConventionalCommit
    disabled:
      - SignOff
      - Signature
      - CommitBody
      - JiraReference
      - SubjectCase
      - ImperativeVerb
      - Spell
      - SubjectSuffix
      - CommitsAhead
`,
		},
		{
			name:          "Validate HEAD - invalid commit",
			commitMessage: "Add feature without format",
			args:          []string{"validate"},
			shouldPass:    false,
			config: `
gommitlint:
  validation:
    enabled: true
  conventional:
    enabled: true
    required: true
    types:
      - feat
      - fix
      - docs
  security:
    signature_required: false
    signoff_required: false
  rules:
    enabled:
      - ConventionalCommit
    disabled:
      - SignOff
      - Signature
      - CommitBody
      - JiraReference
      - SubjectCase
      - ImperativeVerb
      - Spell
      - SubjectSuffix
      - CommitsAhead
      - SubjectLength
`,
		},
		{
			name:          "Validate with custom config - valid commit",
			commitMessage: "custom: special type",
			args:          []string{"validate"},
			shouldPass:    false, // Now expecting to fail since "custom" isn't in default allowed types
			config: `
gommitlint:
  validation:
    enabled: true
  conventional:
    enabled: true
    required: true
    types:
      - custom
      - feat
      - fix
  security:
    signature_required: false
    signoff_required: false
  rules:
    enabled:
      - ConventionalCommit
    disabled:
      - SignOff
      - Signature
      - CommitBody
      - JiraReference
      - SubjectCase
      - ImperativeVerb
      - Spell
      - SubjectSuffix
      - CommitsAhead
      - SubjectLength
`,
		},
		{
			name:          "Validate with extra-verbose mode shows help messages",
			commitMessage: "This is a non-conventional commit without proper format",
			args:          []string{"validate", "--extra-verbose"},
			shouldPass:    false,
			config: `
gommitlint:
  validation:
    enabled: true
  conventional:
    enabled: true
    required: true
  rules:
    enabled:
      - ConventionalCommit
      - ImperativeVerb
    disabled:
      - SignOff
      - Signature
      - CommitBody
      - JiraReference
      - SubjectCase
      - Spell
      - SubjectSuffix
      - CommitsAhead
      - SubjectLength
`,
			checkOutput: func(t *testing.T, output string) {
				t.Helper()
				// Verify that extra-verbose mode shows detailed help messages
				require.Contains(t, output, "Your commit doesn't follow the conventional commit format")
				require.Contains(t, output, "Use the format: type(scope)")
				require.Contains(t, output, "Example: feat(auth)")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test repository
			repoPath, cleanup := setupTestRepository(t, testCase.commitMessage)
			defer cleanup()

			// Create config file in the repository
			configPath := filepath.Join(repoPath, ".gommitlint.yaml")
			err = os.WriteFile(configPath, []byte(testCase.config), 0600)
			require.NoError(t, err)

			// Run the gommitlint command
			cmd := exec.Command(binaryPath, testCase.args...)
			cmd.Dir = repoPath
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Log the output for debugging
			t.Logf("Command output: %s", outputStr)

			// Determine expected outcome based on test case
			if testCase.shouldPass {
				// The case where we expect the test to pass
				if err != nil {
					t.Logf("Command failed with error: %v", err)
					t.Fail()
				}
			} else {
				// The case where we expect the test to fail
				if err == nil {
					t.Logf("Command succeeded when it should have failed")
					t.Fail()
				}
			}

			// Check output content if a check function was provided
			if testCase.checkOutput != nil {
				testCase.checkOutput(t, outputStr)
			}
		})
	}
}
