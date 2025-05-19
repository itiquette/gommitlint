// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package types defines the core configuration data structures for gommitlint.
//
// All types follow value semantics and immutability principles. Configuration
// changes are made through functional transformation methods that return new
// instances rather than modifying existing ones.
//
// The types are separated from the main config package to avoid import cycles
// and provide a clean interface for configuration data.
//
// Structure:
//   - Config: Root configuration structure
//   - SubjectConfig: Subject line validation configuration
//   - BodyConfig: Commit body validation configuration
//   - ConventionalConfig: Conventional commit format configuration
//   - RulesConfig: Rule enable/disable configuration
//   - SecurityConfig: Security and signature configuration
//   - RepositoryConfig: Git repository configuration
//   - OutputConfig: Output formatting configuration
//   - SpellCheckConfig: Spell checking configuration
//   - JiraConfig: JIRA reference validation configuration
//
// All types provide With* methods for functional transformations:
//
//	cfg := cfg.WithSubject(cfg.Subject.WithMaxLength(80))
package types
