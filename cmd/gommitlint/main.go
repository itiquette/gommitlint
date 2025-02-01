// SPDX-FileCopyrightText: itiquette
//
// SPDX-License-Identifier: MPL-2.0

package main

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	Execute(version, commit, date)
}
