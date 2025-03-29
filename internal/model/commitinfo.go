// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "github.com/go-git/go-git/v5/plumbing/object"

// CommitInfo holds information about a commit.
type CommitInfo struct {
	Message   string         // Full commit message
	Subject   string         // First line of commit message
	Body      string         // Rest of commit message after first line
	Signature string         // Signature
	RawCommit *object.Commit // Gives access to the full commit object
}
