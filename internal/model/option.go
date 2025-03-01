// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package model

// Options defines the configuration for commit linting.
type Options struct {
	CommitMsgFile *string
	CommitRef     string
	RevisionRange string
}

// NewOptions creates an Options instance with default values.
func NewOptions() *Options {
	return &Options{
		CommitMsgFile: nil,
		CommitRef:     "",
		RevisionRange: "",
	}
}
