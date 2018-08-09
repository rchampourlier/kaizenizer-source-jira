package jira_test

import (
	"fmt"
	"testing"
	"time"

	extJira "github.com/andygrunwald/go-jira"

	"github.com/rchampourlier/agilizer-source-jira/jira"
	"github.com/rchampourlier/agilizer-source-jira/jira/client"
	"github.com/rchampourlier/agilizer-source-jira/store"
)

// This tests the `jira` package.
//
// `jira/mapping.go` is not unit tested but its tested in integration
// through these tests.

type issueMockDef struct {
	key        string
	refTime    time.Time
	assignee   *string // initial assignee, will be modified by changelogs
	status     string  // initial status, will be modified by changelogs
	changelogs []struct {
		field string
		from  string
		to    string
	}
}

func TestPerformIncrementalSync(t *testing.T) {
	refTime := time.Now()
	assigneeName := "PM"
	issueMockDefs := []issueMockDef{
		issueMockDef{
			"PJ-1",
			refTime,
			&assigneeName,
			"Open",
			[]struct{ field, from, to string }{},
		},
		issueMockDef{
			"PJ-2",
			refTime,
			nil,
			"Open",
			[]struct{ field, from, to string }{
				{"status", "Open", "In Dev"},
				{"assignee", "PM", "Dev"},
			},
		},
		issueMockDef{
			"PJ-3",
			refTime,
			&assigneeName,
			"In Dev",
			[]struct{ field, from, to string }{
				{"assignee", "Me", "You"},
			},
		},
	}
	var issueKeys []string
	for _, issueMockDef := range issueMockDefs {
		issueKeys = append(issueKeys, issueMockDef.key)
	}

	c := client.NewMockClient(t)
	s := store.NewMockStore(t)

	s.ExpectGetRestartFromUpdatedAt().WillReturn(refTime)

	// Perform a search with `updated > 'max issue_updated_at'`
	expectedJiraQuery := fmt.Sprintf("updated > '%s' ORDER BY created ASC", timeAsStr(refTime))
	c.ExpectSearchIssues(expectedJiraQuery).
		WillRespondWithIssueKeys(issueKeys)

	// for each issue in the search results
	for _, def := range issueMockDefs {
		i, is, ies := mockIssue(def)

		// perform a get issue
		c.ExpectGetIssue(def.key).WillRespondWithIssue(i)

		// replace previous records of the issue with new
		//   state and events records
		s.ExpectReplaceIssueStateAndEvents().
			WithIssueKey(def.key).
			WithIssueState(is).
			WithIssueEvents(ies).
			WillReturnError(nil)
	}

	jira.PerformIncrementalSync(c, s, 10)
}

// mockIssue mocks a Jira issue. It returns the mocked `extJira.Issue` as well
// as the corresponding `store.IssueState` and `store.IssueEvent`s that are to
// be expected for this issue.
func mockIssue(def issueMockDef) (*extJira.Issue, *store.IssueState, []*store.IssueEvent) {
	issueFields := extJira.IssueFields{
		Type:           extJira.IssueType{Name: "Bug"},
		Project:        extJira.Project{Key: "PJ", Name: "Project"},
		Status:         &extJira.Status{Name: "In Review"},
		Resolutiondate: extJira.Time(def.refTime),
		Created:        extJira.Time(def.refTime),
		Duedate:        extJira.Date(def.refTime),
		Updated:        extJira.Time(def.refTime),
		Priority:       &extJira.Priority{Name: "Major"},
		Summary:        "summary",
		Description:    "description",
	}
	changelog := extJira.Changelog{}
	for _, cl := range def.changelogs {
		h := extJira.ChangelogHistory{
			Author: extJira.User{
				Name: fmt.Sprintf("%s_change_author", cl.field),
			},
			Created: timeAsStr(def.refTime),
			Items: []extJira.ChangelogItems{
				extJira.ChangelogItems{
					Field:      cl.field,
					FieldType:  "string",
					From:       cl.from,
					FromString: cl.from,
					To:         cl.to,
					ToString:   cl.to,
				},
			},
		}
		changelog.Histories = append(changelog.Histories, h)
	}

	i := &extJira.Issue{
		Key:       def.key,
		Fields:    &issueFields,
		Changelog: &changelog,
	}
	is := &store.IssueState{
		CreatedAt:   def.refTime,
		UpdatedAt:   def.refTime,
		Key:         def.key,
		Project:     strAddr("Project"),
		ResolvedAt:  timeAddr(def.refTime),
		Summary:     strAddr("summary"),
		Description: strAddr("description"),
	}
	ies := []*store.IssueEvent{
		{
			EventTime:          def.refTime,
			EventKind:          "created",
			EventAuthor:        "N/A", // the issue reporter is not set, required string is replaced by N/A
			IssueKey:           def.key,
			CommentBody:        nil,
			StatusChangeFrom:   nil,
			StatusChangeTo:     nil,
			AssigneeChangeFrom: nil,
			AssigneeChangeTo:   nil,
		},
	}
	for _, cl := range def.changelogs {
		ie := &store.IssueEvent{
			EventTime:   def.refTime,
			EventKind:   fmt.Sprintf("%s_changed", cl.field),
			EventAuthor: fmt.Sprintf("%s_change_author", cl.field),
			IssueKey:    def.key,
			CommentBody: nil,
		}
		switch cl.field {
		case "status":
			ie.StatusChangeFrom = strAddr(cl.from)
			ie.StatusChangeTo = strAddr(cl.to)
		case "assignee":
			ie.AssigneeChangeFrom = strAddr(cl.from)
			ie.AssigneeChangeTo = strAddr(cl.to)
		}
		ies = append(ies, ie)
	}

	return i, is, ies
}

func strAddr(s string) *string {
	return &s
}

func timeAddr(t time.Time) *time.Time {
	return &t
}

func timeAsStr(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-0700")
}
