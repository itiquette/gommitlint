// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides a streamlined approach to configuration management in gommitlint.
// It consolidates the various configuration interfaces and structures into a cohesive, value-based system
// with a consistent access pattern.
package config

import (
	"github.com/itiquette/gommitlint/internal/config/types"
)

type Config = types.Config
type SubjectConfig = types.SubjectConfig
type BodyConfig = types.BodyConfig
type ConventionalConfig = types.ConventionalConfig
type RulesConfig = types.RulesConfig
type SecurityConfig = types.SecurityConfig
type RepositoryConfig = types.RepositoryConfig
type OutputConfig = types.OutputConfig
type SpellCheckConfig = types.SpellCheckConfig
type JiraConfig = types.JiraConfig
