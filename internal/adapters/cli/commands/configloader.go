// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itiquette/gommitlint/internal/adapters/config"
	"github.com/itiquette/gommitlint/internal/adapters/signing"
	configTypes "github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/urfave/cli/v3"
)

// ConfigResult holds both the config and information about where it came from.
type ConfigResult struct {
	Config configTypes.Config
	Source string
}

// LoadConfigFromCommand loads configuration based on command flags.
func LoadConfigFromCommand(cmd *cli.Command) (ConfigResult, error) {
	configPath := cmd.String("gommitconfig")
	ignoreConfig := cmd.Bool("ignore-config")
	repoPath := cmd.Root().String("repo-path")

	if configPath != "" && ignoreConfig {
		return ConfigResult{}, errors.New("cannot specify both --gommitconfig and --ignore-config flags")
	}

	if ignoreConfig {
		// Load only defaults, no file config
		cfg := config.LoadDefaultConfig()
		err := error(nil)

		return ConfigResult{
			Config: cfg,
			Source: "defaults (--ignore-config)",
		}, err
	}

	if configPath != "" {
		// --gommitconfig takes highest precedence
		// Perform security validation first
		if err := secureConfigPathValidation(configPath); err != nil {
			return ConfigResult{}, err
		}

		cfg, err := config.LoadConfigFromPath(configPath)

		return ConfigResult{
			Config: cfg,
			Source: configPath + " (--gommitconfig)",
		}, err
	}

	// Validate repo path for security if specified
	validatedRepoPath := ""

	if repoPath != "" {
		var err error

		validatedRepoPath, err = signing.ValidateGitRepoPath(repoPath)
		if err != nil {
			return ConfigResult{}, fmt.Errorf("invalid repo path: %w", err)
		}
	}

	// Use validated repo-path for config discovery
	cfg, err := config.LoadConfigWithRepoPath(validatedRepoPath)
	if err != nil {
		return ConfigResult{}, err
	}

	// Determine the source by checking which config file was actually loaded
	source := "defaults"

	if foundConfigFile := findExistingConfigFileInRepo(validatedRepoPath); foundConfigFile != "" {
		if validatedRepoPath != "" {
			source = foundConfigFile + " (--repo-path)"
		} else {
			source = foundConfigFile
		}
	}

	return ConfigResult{
		Config: cfg,
		Source: source,
	}, nil
}

// findExistingConfigFile finds the first existing config file using the same logic as the config loader.
func findExistingConfigFile() string {
	return findExistingConfigFileInRepo("")
}

// findExistingConfigFileInRepo finds the first existing config file in repository-specific search paths.
func findExistingConfigFileInRepo(repoPath string) string {
	var paths []string

	// Determine base directory for config files
	baseDir := "."
	if repoPath != "" {
		baseDir = repoPath
	}

	// Add local config files (in repository or current directory)
	paths = []string{
		filepath.Join(baseDir, ".gommitlint.yaml"),
		filepath.Join(baseDir, ".gommitlint.yml"),
		filepath.Join(baseDir, ".gommitlint.toml"),
	}

	// Add XDG config paths if XDG_CONFIG_HOME is set and directory exists
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		// Validate XDG_CONFIG_HOME path for security
		cleanXDG := filepath.Clean(xdgConfigHome)
		if filepath.IsAbs(cleanXDG) && !strings.Contains(cleanXDG, "..") {
			gommitlintDir := filepath.Join(cleanXDG, "gommitlint")
			if _, err := os.Stat(gommitlintDir); err == nil {
				paths = append(paths,
					filepath.Join(gommitlintDir, "config.yaml"),
					filepath.Join(gommitlintDir, "config.yml"),
					filepath.Join(gommitlintDir, "config.toml"),
				)
			}
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
