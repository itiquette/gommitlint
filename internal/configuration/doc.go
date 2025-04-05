// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package configuration provides functionality for loading and managing
// configuration for git commit message linting in the gommitlint tool.
//
// The package implements a configuration system that follows the XDG Base
// Directory Specification, looking for configuration files in standard
// locations. It supports both global configuration via XDG_CONFIG_HOME and
// local per-project configuration via .gommitlint.yaml files.
//
// # Configuration Structure
//
// GommitLintConfig is the main configuration structure defining all validation
// rules applied to git commits. It's organized into several categories:
//
//   - Content validation: rules for commit message formatting (subject, body,
//     conventional commits format, spell checking)
//   - Security validation: rules for commit signatures and sign-offs
//   - Miscellaneous: additional validation rules
//
// # Configuration Loading
//
// Configuration is loaded from the following locations (in order of precedence):
//  1. Local .gommitlint.yaml in the current directory
//  2. Global configuration in $XDG_CONFIG_HOME/gommitlint/gommitlint.yaml
//
// Usage Example
//
//	// Load configuration with default loader
//	config, err := configuration.New()
//	if err != nil {
//		log.Fatalf("Failed to load configuration: %v", err)
//	}
//
//	// Access configuration values
//	if config.GommitConf.Subject != nil && config.GommitConf.Subject.MaxLength > 0 {
//		fmt.Printf("Maximum subject length: %d\n", config.GommitConf.Subject.MaxLength)
//	}
//
// YAML Configuration Example
//
//	gommitlint:
//	  subject:
//	    case: lower
//	    imperative: true
//	    max-length: 50
//	    invalid-suffixes: ".,"
//	  body:
//	    required: true
//	  conventional-commit:
//	    required: true
//	    types:
//	      - feat
//	      - fix
//	      - docs
//	    scopes:
//	      - ui
//	      - backend
//	  sign-off: true
package configuration
