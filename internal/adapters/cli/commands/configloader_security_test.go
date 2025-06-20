// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// TestGommitConfigSecurity tests security aspects of --gommitconfig flag.
func TestGommitConfigSecurity(t *testing.T) {
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalDir) }()

	tests := []struct {
		name         string
		configPath   string
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name:         "valid relative path",
			configPath:   "config.yaml",
			expectError:  true, // File doesn't exist, but path is valid
			errorPattern: "not found",
			description:  "should handle valid relative paths",
		},
		{
			name:         "path traversal attack",
			configPath:   "../../../etc/passwd",
			expectError:  true,
			errorPattern: "path traversal",
			description:  "should reject path traversal attempts",
		},
		{
			name:         "absolute path to sensitive file",
			configPath:   "/etc/passwd",
			expectError:  true,
			errorPattern: "absolute path not allowed",
			description:  "should reject absolute paths to system files",
		},
		{
			name:         "null byte injection",
			configPath:   "config.yaml\x00.txt",
			expectError:  true,
			errorPattern: "invalid character",
			description:  "should reject paths with null bytes",
		},
		{
			name:         "very long path",
			configPath:   strings.Repeat("a", 1001),
			expectError:  true,
			errorPattern: "path too long",
			description:  "should reject excessively long paths",
		},
		{
			name:         "path with newlines",
			configPath:   "config\n.yaml",
			expectError:  true,
			errorPattern: "invalid character",
			description:  "should reject paths with control characters",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			err := os.Chdir(tmpDir)
			require.NoError(t, err)

			// Create CLI app context
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "gommitconfig"},
					&cli.BoolFlag{Name: "ignore-config"},
				},
			}

			err = app.Set("gommitconfig", testCase.configPath)
			require.NoError(t, err)

			// Test config loading
			_, err = LoadConfigFromCommand(app)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

// TestGommitConfigFilePermissions tests file permission validation.
func TestGommitConfigFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name         string
		fileMode     os.FileMode
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name:        "readable config file",
			fileMode:    0600,
			expectError: false,
			description: "should accept readable config files",
		},
		{
			name:         "world-writable config file",
			fileMode:     0666,
			expectError:  true,
			errorPattern: "insecure permissions",
			description:  "should reject world-writable config files",
		},
		{
			name:         "group-writable config file",
			fileMode:     0620,
			expectError:  true,
			errorPattern: "insecure permissions",
			description:  "should reject group-writable config files",
		},
		{
			name:         "executable config file",
			fileMode:     0700,
			expectError:  true,
			errorPattern: "config file should not be executable",
			description:  "should reject executable config files",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config file with specific permissions
			configFile := "test-config-" + testCase.name + ".yaml"
			configContent := `gommitlint:
  message:
    subject:
      max_length: 50`

			err := os.WriteFile(configFile, []byte(configContent), testCase.fileMode)
			require.NoError(t, err)

			// Explicitly set the file mode to override umask effects
			err = os.Chmod(configFile, testCase.fileMode)
			require.NoError(t, err)

			// Create CLI app context
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "gommitconfig"},
				},
			}

			err = app.Set("gommitconfig", configFile)
			require.NoError(t, err)

			// Test config loading
			_, err = LoadConfigFromCommand(app)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

// TestGommitConfigSymlinkSecurity tests symlink attack prevention.
func TestGommitConfigSymlinkSecurity(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping symlink tests when running as root")
	}

	tmpDir := t.TempDir()

	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name         string
		setupSymlink func() string
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name: "symlink to sensitive file",
			setupSymlink: func() string {
				// Create symlink to /etc/passwd
				symlinkPath := "config.yaml"
				_ = os.Symlink("/etc/passwd", symlinkPath)

				return symlinkPath
			},
			expectError:  true,
			errorPattern: "symlink not allowed",
			description:  "should reject symlinks to sensitive files",
		},
		{
			name: "symlink to valid config in temp dir",
			setupSymlink: func() string {
				// Create valid config file
				realConfig := "real-config.yaml"
				configContent := `gommitlint:
  message:
    subject:
      max_length: 50`
				_ = os.WriteFile(realConfig, []byte(configContent), 0600)

				// Create symlink to it
				symlinkPath := "config.yaml"
				_ = os.Symlink(realConfig, symlinkPath)

				return symlinkPath
			},
			expectError:  true,
			errorPattern: "symlink not allowed",
			description:  "should reject all symlinks for security",
		},
		{
			name: "broken symlink",
			setupSymlink: func() string {
				symlinkPath := "config.yaml"
				_ = os.Symlink("nonexistent-file.yaml", symlinkPath)

				return symlinkPath
			},
			expectError:  true,
			errorPattern: "symlink not allowed",
			description:  "should reject broken symlinks",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			configPath := testCase.setupSymlink()
			defer os.Remove(configPath)

			// Create CLI app context
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "gommitconfig"},
				},
			}

			err = app.Set("gommitconfig", configPath)
			require.NoError(t, err)

			// Test config loading
			_, err = LoadConfigFromCommand(app)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}

// TestGommitConfigDirectoryTraversal tests directory traversal scenarios.
func TestGommitConfigDirectoryTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalDir) }()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "nested", "dir")
	err = os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(nestedDir)
	require.NoError(t, err)

	// Create config file in parent directory
	parentConfig := filepath.Join(tmpDir, "parent-config.yaml")
	configContent := `gommitlint:
  message:
    subject:
      max_length: 50`
	err = os.WriteFile(parentConfig, []byte(configContent), 0600)
	require.NoError(t, err)

	tests := []struct {
		name         string
		configPath   string
		expectError  bool
		errorPattern string
		description  string
	}{
		{
			name:         "access parent directory",
			configPath:   "../../parent-config.yaml",
			expectError:  true,
			errorPattern: "path traversal",
			description:  "should prevent access to parent directories",
		},
		{
			name:         "complex traversal pattern",
			configPath:   "../../../tmp/../" + filepath.Base(tmpDir) + "/parent-config.yaml",
			expectError:  true,
			errorPattern: "path traversal",
			description:  "should prevent complex traversal patterns",
		},
		{
			name:         "current directory access",
			configPath:   "./local-config.yaml",
			expectError:  true, // File doesn't exist, but path is valid
			errorPattern: "not found",
			description:  "should allow current directory access",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create CLI app context
			app := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "gommitconfig"},
				},
			}

			err = app.Set("gommitconfig", testCase.configPath)
			require.NoError(t, err)

			// Test config loading
			_, err = LoadConfigFromCommand(app)

			if testCase.expectError {
				require.Error(t, err, testCase.description)

				if testCase.errorPattern != "" {
					require.Contains(t, err.Error(), testCase.errorPattern, testCase.description)
				}
			} else {
				require.NoError(t, err, testCase.description)
			}
		})
	}
}
