// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package model

// Options defines the configuration for commit linting.
type Options struct {
	MsgFromFile   *string
	CommitRef     string
	RevisionRange string
}

// NewOptions creates an Options instance with default values.
func NewOptions() *Options {
	return &Options{
		MsgFromFile:   nil,
		CommitRef:     "",
		RevisionRange: "",
	}
}
