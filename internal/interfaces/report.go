// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package interfaces

// Rule defines a commit check.
type Rule interface {
	Name() string
	Message() string
	Errors() []error
}
