// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) error
		expectedTypes []string
		wantErr       bool
		errorMessage  string
	}{
		{
			name: "Load configuration from local file",
			setupFunc: func(dir string) error {
				content := `gommitlint:
  conventional-commit:
    types:
      - custom
      - types`

				return os.WriteFile(filepath.Join(dir, ".gommitlint.yaml"), []byte(content), 0600)
			},
			expectedTypes: []string{"custom", "types"},
			wantErr:       false,
		},
		{
			name: "Load configuration from XDG config home",
			setupFunc: func(dir string) error {
				xdgPath := filepath.Join(dir, "gommitlint")
				if err := os.MkdirAll(xdgPath, 0755); err != nil {
					return err
				}
				content := `gommitlint:
  conventional-commit:
    types:
      - xdg
      - config`
				configPath := filepath.Join(xdgPath, "gommitlint.yaml")
				if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
					return err
				}
				os.Setenv("XDG_CONFIG_HOME", dir)

				return nil
			},
			expectedTypes: []string{"xdg", "config"},
			wantErr:       false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Create a temporary directory for this test
			tmpDir := t.TempDir()

			// Change to temp directory for test
			err := os.Chdir(tmpDir)
			require.NoError(t, err)

			// Setup test environment
			err = tabletest.setupFunc(tmpDir)
			require.NoError(t, err, "Setup failed")

			// Create and use the config loader
			loader := DefaultConfigLoader{}
			config, err := loader.LoadConfiguration()

			if tabletest.wantErr {
				require.Error(t, err)

				if tabletest.errorMessage != "" {
					require.Equal(t, tabletest.errorMessage, err.Error())
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			require.Equal(t, tabletest.expectedTypes, config.GommitConf.ConventionalCommit.Types)
		})
	}
}

func TestReadConfigurationFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string) error
		wantErr   bool
		wantPanic bool
	}{
		{
			name: "No configuration files exist",
			setupFunc: func(string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "Valid configuration file",
			setupFunc: func(dir string) error {
				content := `
gommitlint:
  conventional:
    types:
      - test
      - valid
`

				return os.WriteFile(filepath.Join(dir, ".gommitlint.yaml"), []byte(content), 0600)
			},
			wantErr: false,
		},
		{
			name: "Invalid YAML that causes unmarshal panic",
			setupFunc: func(dir string) error {
				content := `
gommitlint: [invalid
  conventional:
    types:
      - test
`

				return os.WriteFile(filepath.Join(dir, ".gommitlint.yaml"), []byte(content), 0600)
			},
			wantErr: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Create a temporary directory for this test
			tmpDir := t.TempDir()

			// Change to temp directory for test
			err := os.Chdir(tmpDir)
			require.NoError(t, err)

			err = tabletest.setupFunc(tmpDir)
			require.NoError(t, err, "Setup failed")

			appConfig := &AppConf{}

			if tabletest.wantPanic {
				require.Panics(t, func() {
					_ = ReadConfigurationFile(appConfig, ".gommitlint.yaml")
				})

				return
			}

			err = ReadConfigurationFile(appConfig, ".gommitlint.yaml")

			if tabletest.wantErr {
				require.Error(t, err)
				require.Equal(t, "error loading config: yaml: line 1: did not find expected ',' or ']'", err.Error())

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestHasXDGConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string) error
		expectedExists bool
	}{
		{
			name: "XDG config file exists",
			setupFunc: func(dir string) error {
				configPath := filepath.Join(dir, "gommitlint", "gommitlint.yaml")
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					return err
				}

				return os.WriteFile(configPath, []byte("test"), 0600)
			},
			expectedExists: true,
		},
		{
			name: "XDG config file does not exist",
			setupFunc: func(_ string) error {
				return nil
			},
			expectedExists: false,
		},
		{
			name: "XDG_CONFIG_HOME not set",
			setupFunc: func(_ string) error {
				os.Unsetenv("XDG_CONFIG_HOME")

				return nil
			},
			expectedExists: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			err := tabletest.setupFunc(tmpDir)
			require.NoError(t, err, "Setup failed")

			t.Setenv("XDG_CONFIG_HOME", tmpDir)

			exists, path := hasXDGConfigFile("XDG_CONFIG_HOME", "/gommitlint/gommitlint.yaml")
			require.Equal(t, tabletest.expectedExists, exists)

			if tabletest.expectedExists {
				require.NotEmpty(t, path)
			}
		})
	}
}

func TestHasLocalConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(string) error
		configFile     string
		expectedExists bool
	}{
		{
			name: "Local config file exists",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".gommitlint.yaml"), []byte("test"), 0600)
			},
			configFile:     ".gommitlint.yaml",
			expectedExists: true,
		},
		{
			name: "Local config file does not exist",
			setupFunc: func(string) error {
				return nil
			},
			configFile:     "nonexistent.yaml",
			expectedExists: false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			err := os.Chdir(tmpDir)
			require.NoError(t, err)

			err = tabletest.setupFunc(tmpDir)
			require.NoError(t, err, "Setup failed")

			exists := hasLocalConfigFile(tabletest.configFile)
			require.Equal(t, tabletest.expectedExists, exists)
		})
	}
}
