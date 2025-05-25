// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package ports defines shared types used by port interfaces.
package ports

import "io"

// ReportOptions defines options for report generation.
// This is a port-level type to avoid adapters depending on application types.
type ReportOptions struct {
	// Format specifies the output format (text, json, github, gitlab).
	Format string
	// Verbose indicates whether to include detailed information.
	Verbose bool
	// ExtraVerbose indicates whether to include extra detailed information.
	ExtraVerbose bool
	// ShowHelp indicates whether to show help for rules.
	ShowHelp bool
	// RuleToShowHelp specifies a specific rule to show help for.
	RuleToShowHelp string
	// LightMode indicates whether to use light color scheme.
	LightMode bool
	// Writer is the output writer.
	Writer io.Writer
}
