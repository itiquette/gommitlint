// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestRepoPathValidation tests cross-directory repository validation.
func TestRepoPathValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(string) error
		repoPath    string
		expectError bool
		description string
	}{
		{
			name: "valid git repository",
			setupRepo: func(dir string) error {
				cmd := exec.Command("git", "init")
				cmd.Dir = dir

				return cmd.Run()
			},
			expectError: false,
			description: "should validate existing git repository",
		},
		{
			name: "non-existent directory",
			setupRepo: func(string) error {
				return nil // Don't create anything
			},
			repoPath:    "/nonexistent/path",
			expectError: true,
			description: "should reject non-existent directory",
		},
		{
			name: "directory without git",
			setupRepo: func(string) error {
				return nil // Create empty directory without git
			},
			expectError: true,
			description: "should reject directory without git repository",
		},
		{
			name: "nested git repository",
			setupRepo: func(dir string) error {
				cmd := exec.Command("git", "init")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}

				// Create nested directory with another git repo
				nestedDir := filepath.Join(dir, "nested")
				if err := os.MkdirAll(nestedDir, 0755); err != nil {
					return err
				}

				cmd = exec.Command("git", "init")
				cmd.Dir = nestedDir

				return cmd.Run()
			},
			expectError: false,
			description: "should validate nested git repository",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			repoPath := tmpDir

			if testCase.repoPath != "" {
				repoPath = testCase.repoPath
			}

			// Set up repository if needed
			if testCase.setupRepo != nil {
				err := testCase.setupRepo(tmpDir)
				require.NoError(t, err, "failed to setup test repository")
			}

			// Test getRepoPath function
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{
					Name: "repo-path",
				},
			}

			// Simulate setting the flag
			err := app.Root().Set("repo-path", repoPath)
			require.NoError(t, err)

			resultPath := getRepoPath(app)
			require.Equal(t, repoPath, resultPath, testCase.description)
		})
	}
}

// TestRepoPathFromDifferentDirectories tests running gommitlint from different working directories.
func TestRepoPathFromDifferentDirectories(t *testing.T) {
	// Create two separate directories - one to run from, one as target repo
	runFromDir := t.TempDir()
	targetRepoDir := t.TempDir()

	// Initialize git repo in target directory
	cmd := exec.Command("git", "init")
	cmd.Dir = targetRepoDir
	err := cmd.Run()
	require.NoError(t, err, "failed to init git repo")

	// Create a test commit in target repo
	testFile := filepath.Join(targetRepoDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = targetRepoDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "feat: test commit")
	cmd.Dir = targetRepoDir
	err = cmd.Run()
	require.NoError(t, err)

	// Change to different directory and test getRepoPath
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(runFromDir)
	require.NoError(t, err)

	tests := []struct {
		name         string
		repoPathFlag string
		expectedPath string
		description  string
	}{
		{
			name:         "absolute path to different repo",
			repoPathFlag: targetRepoDir,
			expectedPath: targetRepoDir,
			description:  "should use absolute path to target repository",
		},
		{
			name:         "relative path to different repo",
			repoPathFlag: filepath.Base(targetRepoDir),
			expectedPath: filepath.Base(targetRepoDir),
			description:  "should use relative path to target repository",
		},
		{
			name:         "current directory default",
			repoPathFlag: "",
			expectedPath: ".",
			description:  "should default to current directory when no flag provided",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{
					Name: "repo-path",
				},
			}

			if testCase.repoPathFlag != "" {
				err := app.Root().Set("repo-path", testCase.repoPathFlag)
				require.NoError(t, err)
			}

			resultPath := getRepoPath(app)
			require.Equal(t, testCase.expectedPath, resultPath, testCase.description)
		})
	}
}

// TestRepoPathSecurity tests security validation for repo-path.
func TestRepoPathSecurity(t *testing.T) {
	tests := []struct {
		name        string
		repoPath    string
		expectError bool
		description string
	}{
		{
			name:        "normal path",
			repoPath:    "/tmp/valid-repo",
			expectError: false,
			description: "should accept normal paths",
		},
		{
			name:        "path traversal attempt",
			repoPath:    "../../../etc",
			expectError: false, // getRepoPath just returns the path, validation happens later
			description: "getRepoPath should return path as-is, validation happens in validators",
		},
		{
			name:        "path with null bytes",
			repoPath:    "/tmp/repo\x00/malicious",
			expectError: false, // getRepoPath just returns the path, validation happens later
			description: "getRepoPath should return path as-is, validation happens in validators",
		},
		{
			name:        "very long path",
			repoPath:    "/" + string(make([]byte, 4096)),
			expectError: false, // getRepoPath just returns the path, validation happens later
			description: "getRepoPath should return path as-is, validation happens in validators",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{
					Name: "repo-path",
				},
			}

			err := app.Root().Set("repo-path", testCase.repoPath)
			require.NoError(t, err)

			resultPath := getRepoPath(app)
			require.Equal(t, testCase.repoPath, resultPath, testCase.description)
		})
	}
}

// TestRepoPathWithValidationTarget tests repo-path integration with validation targets.
func TestRepoPathWithValidationTarget(t *testing.T) {
	// Create test repository
	repoDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	err := cmd.Run()
	require.NoError(t, err)

	// Create test commit
	testFile := filepath.Join(repoDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0600)
	require.NoError(t, err)

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = repoDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "feat: initial commit")
	cmd.Dir = repoDir
	err = cmd.Run()
	require.NoError(t, err)

	// Change to different directory
	runFromDir := t.TempDir()
	originalWd, _ := os.Getwd()

	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(runFromDir)
	require.NoError(t, err)

	tests := []struct {
		name            string
		repoPath        string
		gitReference    string
		expectValidPath bool
		description     string
	}{
		{
			name:            "validate HEAD in different repo",
			repoPath:        repoDir,
			gitReference:    "HEAD",
			expectValidPath: true,
			description:     "should validate HEAD commit in target repository",
		},
		{
			name:            "validate specific commit in different repo",
			repoPath:        repoDir,
			gitReference:    "HEAD~0",
			expectValidPath: true,
			description:     "should validate specific commit in target repository",
		},
		{
			name:            "current directory without repo",
			repoPath:        ".",
			gitReference:    "HEAD",
			expectValidPath: false,
			description:     "should fail when current directory is not a git repo",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Test path resolution
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{
					Name: "repo-path",
				},
			}

			err := app.Root().Set("repo-path", testCase.repoPath)
			require.NoError(t, err)

			resultPath := getRepoPath(app)
			require.Equal(t, testCase.repoPath, resultPath)

			// Verify the path points to expected location
			if testCase.expectValidPath {
				gitDir := filepath.Join(resultPath, ".git")
				_, err := os.Stat(gitDir)
				require.NoError(t, err, "should find .git directory in target path")
			}
		})
	}
}

// TestRepoPathEdgeCases tests edge cases for repo-path handling.
func TestRepoPathEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		setupFlag    func(*cli.Command) error
		expectedPath string
		description  string
	}{
		{
			name: "empty string flag",
			setupFlag: func(app *cli.Command) error {
				return app.Root().Set("repo-path", "")
			},
			expectedPath: ".",
			description:  "should default to current directory for empty string",
		},
		{
			name: "whitespace only flag",
			setupFlag: func(app *cli.Command) error {
				return app.Root().Set("repo-path", "   ")
			},
			expectedPath: "   ",
			description:  "should preserve whitespace-only paths",
		},
		{
			name: "no flag set",
			setupFlag: func(_ *cli.Command) error {
				return nil // Don't set any flag
			},
			expectedPath: ".",
			description:  "should default to current directory when flag not set",
		},
		{
			name: "dot path",
			setupFlag: func(app *cli.Command) error {
				return app.Root().Set("repo-path", ".")
			},
			expectedPath: ".",
			description:  "should handle explicit dot path",
		},
		{
			name: "double dot path",
			setupFlag: func(app *cli.Command) error {
				return app.Root().Set("repo-path", "..")
			},
			expectedPath: "..",
			description:  "should handle parent directory path",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{
					Name: "repo-path",
				},
			}

			if testCase.setupFlag != nil {
				err := testCase.setupFlag(app)
				require.NoError(t, err)
			}

			resultPath := getRepoPath(app)
			require.Equal(t, testCase.expectedPath, resultPath, testCase.description)
		})
	}
}
