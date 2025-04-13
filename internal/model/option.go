// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package model

// Options stores command-line options for validation.
type Options struct {
	MsgFromFile    *string
	RevisionRange  string
	CommitRef      string
	Verbose        bool   // Added for verbose output
	ShowHelp       bool   // Added for detailed rule help
	RuleToShowHelp string // Added to track which rule's help to show
	LightMode      bool   // Added to track which color scheme
}

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return &Options{
		MsgFromFile:    nil,
		RevisionRange:  "",
		CommitRef:      "",
		Verbose:        false,
		ShowHelp:       false,
		RuleToShowHelp: "",
		LightMode:      false,
	}
}
