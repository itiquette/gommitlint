// SPDX-FileCopyrightText: itiquette
//
// SPDX-License-Identifier: MPL-2.0

package main

import "github.com/itiquette/gommitlint/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
