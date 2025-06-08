// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// createTestGitRepo creates a temporary directory with a minimal .git directory
// to satisfy Git repository validation requirements.
func createTestGitRepo(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create minimal git config to make it a valid git repo
	configFile := filepath.Join(gitDir, "config")
	err = os.WriteFile(configFile, []byte("[core]\n"), 0600)
	require.NoError(t, err)

	return tempDir
}

// TestConfigPrecedenceWithRepoPath tests config file discovery order with --repo-path flag.
func TestConfigPrecedenceWithRepoPath(t *testing.T) {
	// Create test directories with Git repositories
	currentDir := createTestGitRepo(t)
	repoDir := createTestGitRepo(t)

	// Change to current directory for test
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err := os.Chdir(currentDir)
	require.NoError(t, err)

	// Create config files in current directory
	currentConfigContent := `gommitlint:
  output: json
  message:
    subject:
      maxlength: 50`
	err = os.WriteFile(filepath.Join(currentDir, ".gommitlint.yaml"), []byte(currentConfigContent), 0600)
	require.NoError(t, err)

	// Create config files in repo directory
	repoConfigContent := `gommitlint:
  output: text
  message:
    subject:
      maxlength: 60`
	err = os.WriteFile(filepath.Join(repoDir, ".gommitlint.yaml"), []byte(repoConfigContent), 0600)
	require.NoError(t, err)

	tests := []struct {
		name           string
		repoPath       string
		expectedOutput string
		expectedSource string
		description    string
	}{
		{
			name:           "no repo-path uses current directory config",
			repoPath:       "",
			expectedOutput: "json",
			expectedSource: ".gommitlint.yaml",
			description:    "should use config from current directory when no --repo-path",
		},
		{
			name:           "repo-path uses repository directory config",
			repoPath:       repoDir,
			expectedOutput: "text",
			expectedSource: filepath.Join(repoDir, ".gommitlint.yaml") + " (--repo-path)",
			description:    "should use config from repository directory when --repo-path specified",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context
			app := &cli.Command{}
			app.Root().Flags = []cli.Flag{
				&cli.StringFlag{Name: "repo-path"},
			}

			if testCase.repoPath != "" {
				err := app.Root().Set("repo-path", testCase.repoPath)
				require.NoError(t, err)
			}

			// Load config from command
			result, err := LoadConfigFromCommand(app)
			require.NoError(t, err, testCase.description)

			// Verify config content
			require.Equal(t, testCase.expectedOutput, result.Config.Output, testCase.description)
			require.Equal(t, testCase.expectedSource, result.Source, testCase.description)
		})
	}
}

// TestConfigPrecedenceOrder tests the complete precedence order for config loading.
func TestConfigPrecedenceOrder(t *testing.T) {
	// Create test directories with Git repositories
	currentDir := createTestGitRepo(t)
	repoDir := createTestGitRepo(t)

	// Change to current directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err := os.Chdir(currentDir)
	require.NoError(t, err)

	// Create explicit config file (use relative path for security compliance)
	explicitConfigFile := "explicit-config.yaml"
	explicitConfigContent := `gommitlint:
  output: github`
	err = os.WriteFile(explicitConfigFile, []byte(explicitConfigContent), 0600)
	require.NoError(t, err)

	// Create config in current directory
	currentConfigContent := `gommitlint:
  output: json`
	err = os.WriteFile(filepath.Join(currentDir, ".gommitlint.yaml"), []byte(currentConfigContent), 0600)
	require.NoError(t, err)

	// Create config in repo directory
	repoConfigContent := `gommitlint:
  output: text`
	err = os.WriteFile(filepath.Join(repoDir, ".gommitlint.yaml"), []byte(repoConfigContent), 0600)
	require.NoError(t, err)

	tests := []struct {
		name           string
		expectedOutput string
		expectedSource string
		description    string
	}{
		{
			name:           "explicit config takes highest precedence",
			expectedOutput: "github",
			expectedSource: explicitConfigFile + " (--gommitconfig)",
			description:    "--gommitconfig should override --repo-path",
		},
		{
			name:           "ignore-config overrides everything",
			expectedOutput: "text",
			expectedSource: "defaults (--ignore-config)",
			description:    "--ignore-config should use only defaults",
		},
		{
			name:           "repo-path config when specified",
			expectedOutput: "text",
			expectedSource: filepath.Join(repoDir, ".gommitlint.yaml") + " (--repo-path)",
			description:    "--repo-path should find config in repository directory",
		},
		{
			name:           "current directory config as fallback",
			expectedOutput: "json",
			expectedSource: ".gommitlint.yaml",
			description:    "should use current directory config when no flags",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context - all flags on same command for simplicity in testing
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "repo-path"},
					&cli.StringFlag{Name: "gommitconfig"},
					&cli.BoolFlag{Name: "ignore-config"},
				},
			}

			// Setup flags manually based on test case
			switch testCase.name {
			case "explicit config takes highest precedence":
				_ = app.Set("repo-path", repoDir)
				_ = app.Set("gommitconfig", explicitConfigFile)
			case "ignore-config overrides everything":
				_ = app.Set("repo-path", repoDir)
				_ = app.Set("ignore-config", "true")
			case "repo-path config when specified":
				err := app.Set("repo-path", repoDir)
				require.NoError(t, err, "failed to set repo-path flag")
			case "current directory config as fallback":
				// No flags set - intentionally empty case
				_ = "no-op"
			}

			// Debug: Print the repo-path value
			repoPathValue := app.Root().String("repo-path")
			t.Logf("repo-path value: '%s'", repoPathValue)

			// Load config
			result, err := LoadConfigFromCommand(app)
			require.NoError(t, err, testCase.description)

			// Verify results
			require.Equal(t, testCase.expectedOutput, result.Config.Output, testCase.description)
			require.Equal(t, testCase.expectedSource, result.Source, testCase.description)
		})
	}
}

// TestConfigFormatPrecedenceWithRepoPath tests config file format precedence in repository paths.
func TestConfigFormatPrecedenceWithRepoPath(t *testing.T) {
	repoDir := createTestGitRepo(t)

	// Create config files in different formats
	yamlContent := `gommitlint:
  output: json`
	err := os.WriteFile(filepath.Join(repoDir, ".gommitlint.yaml"), []byte(yamlContent), 0600)
	require.NoError(t, err)

	ymlContent := `gommitlint:
  output: github`
	err = os.WriteFile(filepath.Join(repoDir, ".gommitlint.yml"), []byte(ymlContent), 0600)
	require.NoError(t, err)

	tomlContent := `[gommitlint]
output = "gitlab"`
	err = os.WriteFile(filepath.Join(repoDir, ".gommitlint.toml"), []byte(tomlContent), 0600)
	require.NoError(t, err)

	tests := []struct {
		name           string
		filesToRemove  []string
		expectedOutput string
		expectedFile   string
		description    string
	}{
		{
			name:           "yaml has highest precedence",
			filesToRemove:  []string{},
			expectedOutput: "json",
			expectedFile:   ".gommitlint.yaml",
			description:    ".yaml should be preferred over .yml and .toml",
		},
		{
			name:           "yml has precedence over toml",
			filesToRemove:  []string{".gommitlint.yaml"},
			expectedOutput: "github",
			expectedFile:   ".gommitlint.yml",
			description:    ".yml should be preferred over .toml when .yaml missing",
		},
		{
			name:           "toml as fallback",
			filesToRemove:  []string{".gommitlint.yaml", ".gommitlint.yml"},
			expectedOutput: "gitlab",
			expectedFile:   ".gommitlint.toml",
			description:    ".toml should be used when yaml formats missing",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Remove specified files for this test
			for _, file := range testCase.filesToRemove {
				_ = os.Remove(filepath.Join(repoDir, file))
			}

			// Restore files after test
			defer func() {
				for _, file := range testCase.filesToRemove {
					var content string

					switch file {
					case ".gommitlint.yaml":
						content = yamlContent
					case ".gommitlint.yml":
						content = ymlContent
					case ".gommitlint.toml":
						content = tomlContent
					}

					_ = os.WriteFile(filepath.Join(repoDir, file), []byte(content), 0600)
				}
			}()

			// Create CLI app context - all flags on same command for simplicity in testing
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "repo-path"},
				},
			}

			err := app.Set("repo-path", repoDir)
			require.NoError(t, err)

			// Load config
			result, err := LoadConfigFromCommand(app)
			require.NoError(t, err, testCase.description)

			// Verify results
			require.Equal(t, testCase.expectedOutput, result.Config.Output, testCase.description)
			require.Contains(t, result.Source, testCase.expectedFile, testCase.description)
			require.Contains(t, result.Source, "(--repo-path)", testCase.description)
		})
	}
}

// TestConfigDiscoveryErrorHandling tests error handling in config discovery with --repo-path.
func TestConfigDiscoveryErrorHandling(t *testing.T) {
	// Create working directory with Git repository but without any config files
	workingDir := createTestGitRepo(t)

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()

	err := os.Chdir(workingDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupFlags  func(*cli.Command) error
		expectError bool
		description string
	}{
		{
			name: "non-existent explicit config file",
			setupFlags: func(app *cli.Command) error {
				return app.Set("gommitconfig", "/nonexistent/config.yaml")
			},
			expectError: true,
			description: "should error when explicit config file doesn't exist",
		},
		{
			name: "non-existent repo path",
			setupFlags: func(app *cli.Command) error {
				return app.Root().Set("repo-path", "/nonexistent/repo")
			},
			expectError: true,
			description: "should error when repo path doesn't exist due to security validation",
		},
		{
			name: "conflicting flags",
			setupFlags: func(app *cli.Command) error {
				if err := app.Set("gommitconfig", "config.yaml"); err != nil {
					return err
				}

				return app.Set("ignore-config", "true")
			},
			expectError: true,
			description: "should error when both --gommitconfig and --ignore-config are specified",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context - all flags on same command for simplicity in testing
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "repo-path"},
					&cli.StringFlag{Name: "gommitconfig"},
					&cli.BoolFlag{Name: "ignore-config"},
				},
			}

			// Setup flags
			err := testCase.setupFlags(app)
			require.NoError(t, err)

			// Load config
			_, err = LoadConfigFromCommand(app)

			if testCase.expectError {
				require.Error(t, err, testCase.description)
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}
