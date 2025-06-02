// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// LoadFromEnv applies environment variable overrides to the configuration.
// Environment variables follow the pattern GOMMITLINT_SECTION_KEY.
func LoadFromEnv(cfg config.Config) config.Config {
	// Output format
	if val := os.Getenv("GOMMITLINT_OUTPUT"); val != "" {
		cfg.Output = val
	}

	// Subject max length
	if val := os.Getenv("GOMMITLINT_SUBJECT_MAX_LENGTH"); val != "" {
		if length, err := strconv.Atoi(val); err == nil && length > 0 {
			cfg.Message.Subject.MaxLength = length
		}
	}

	// Conventional types
	if val := os.Getenv("GOMMITLINT_CONVENTIONAL_TYPES"); val != "" {
		cfg.Conventional.Types = strings.Split(val, ",")
	}

	// Reference branch
	if val := os.Getenv("GOMMITLINT_REPO_REFERENCE_BRANCH"); val != "" {
		cfg.Repo.ReferenceBranch = val
	}

	// Signing requirements
	if val := os.Getenv("GOMMITLINT_SIGNING_REQUIRE_SIGNATURE"); val != "" {
		cfg.Signing.RequireSignature = val == "true"
	}

	// Rule enablement
	if val := os.Getenv("GOMMITLINT_RULES_ENABLED"); val != "" {
		cfg.Rules.Enabled = strings.Split(val, ",")
	}

	if val := os.Getenv("GOMMITLINT_RULES_DISABLED"); val != "" {
		cfg.Rules.Disabled = strings.Split(val, ",")
	}

	return cfg
}
