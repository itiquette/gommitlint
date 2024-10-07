// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

//
//nolint:testpackage
package commit

import (
	"testing"
)

func TestCommit_ValidateJiraCheck(t *testing.T) {
	type fields struct {
		SpellCheck         *SpellCheck
		Conventional       *Conventional
		Header             *HeaderChecks
		Body               *BodyChecks
		DCO                bool
		GPG                bool
		MaximumOfOneCommit bool
		msg                string
	}

	type want struct {
		errorCount int
	}

	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Missing jira issue no type",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "invalid commit",
			},
			want: want{errorCount: 1},
		},
		{
			name: "Missing jira issue with type",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix: invalid commit",
			},
			want: want{errorCount: 1},
		},
		{
			name: "Valid commit",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix: [JIRA-1234] valid commit",
			},
			want: want{errorCount: 0},
		},
		{
			name: "Valid commit 2",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix: [PROJ-1234] valid commit",
			},
			want: want{errorCount: 0},
		},
		{
			name: "Invalid jira project",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix: [FALSE-1234] valid commit",
			},
			want: want{errorCount: 1},
		},
		{
			name: "Invalid jira issue number",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix: JIRA-0 valid commit",
			},
			want: want{errorCount: 1},
		},
		{
			name: "Valid commit with scope",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix(test): [PROJ-1234] valid commit",
			},
			want: want{errorCount: 0},
		},
		{
			name: "Valid commit without square brackets",
			fields: fields{
				Header: &HeaderChecks{
					Jira: &JiraChecks{
						Keys: []string{"JIRA", "PROJ"},
					},
				},
				msg: "fix: PROJ-1234 valid commit",
			},
			want: want{errorCount: 0},
		},
	}
	for _, tt := range tests {
		tabletest := tt
		t.Run(tabletest.name, func(t *testing.T) {
			commit := Commit{
				SpellCheck:   tabletest.fields.SpellCheck,
				Conventional: tabletest.fields.Conventional,
				Header:       tabletest.fields.Header,
				Body:         tabletest.fields.Body,
				DCO:          tabletest.fields.DCO,
				GPG: &GPG{
					Required: tabletest.fields.GPG,
				},
				MaximumOfOneCommit: tabletest.fields.MaximumOfOneCommit,
				msg:                tabletest.fields.msg,
			}
			got := commit.ValidateJiraCheck()

			if len(got.Errors()) != tabletest.want.errorCount {
				t.Errorf("Wanted %d errors but got %d errors: %v", tabletest.want.errorCount, len(got.Errors()), got.Errors())
			}
		})
	}
}
