// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

type CommitRule interface {
	Name() string
	Result() string
	Errors() []error
}
