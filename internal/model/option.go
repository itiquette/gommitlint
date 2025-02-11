// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package model

// Option is a functional option used to pass in arguments.
type Option func(*Options)

// Options defines the set of options available.
type Options struct {
	CommitMsgFile *string
	CommitRef     string
	RevisionRange string
}

// WithCommitMsgFile sets the path to the commit message file.
func WithCommitMsgFile(o *string) Option {
	return func(args *Options) {
		args.CommitMsgFile = o
	}
}

// WithCommitRef sets the ref to compare git policies against.
func WithCommitRef(o string) Option {
	return func(args *Options) {
		args.CommitRef = o
	}
}

// WithRevisionRange sets the revision range to compare git policies against.
func WithRevisionRange(o string) Option {
	return func(args *Options) {
		args.RevisionRange = o
	}
}

// NewDefaultOptions initializes a Options struct with default values.
func NewDefaultOptions(setters ...Option) *Options {
	opts := &Options{
		CommitMsgFile: nil,
	}

	for _, setter := range setters {
		setter(opts)
	}

	return opts
}
