// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package interfaces

// Check defines a commit check.
type Check interface {
	Status() string
	//	Name() string
	Message() string
	Errors() []error
}
