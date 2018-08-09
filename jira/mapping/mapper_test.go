package mapping_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	extJira "github.com/andygrunwald/go-jira"

	"github.com/rchampourlier/agilizer-source-jira/jira/mapping"
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

func TestIssueEventsFromIssue(t *testing.T) {
	key := "PJ-1"
	assigneeName := "PM"
	refTime := time.Now()
	m := mapping.Mapper{}

	t.Run("issue without changelog", func(t *testing.T) {
		i := mockIssue(issueMockDef{
			key,
			refTime,
			&assigneeName,
			"Open",
			[]struct{ field, from, to string }{},
		})
		resultEvents := m.IssueEventsFromIssue(i)

		// TODO: implement expectations on resultEvents
		log.Println(resultEvents)
	})

	t.Run("issue with status and assignee changelogs", func(t *testing.T) {
		i := mockIssue(issueMockDef{
			"PJ-2",
			refTime,
			nil,
			"Open",
			[]struct{ field, from, to string }{
				{"status", "Open", "In Dev"},
				{"assignee", "PM", "Dev"},
			},
		})
		resultEvents := m.IssueEventsFromIssue(i)

		// TODO: implement expectations on resultEvents
		log.Println(resultEvents)
	})
}

func TestIssueStateFromIssue(t *testing.T) {
	key := "PJ-1"
	assigneeName := "PM"
	refTime := time.Now()
	m := mapping.Mapper{}
	i := mockIssue(issueMockDef{
		key,
		refTime,
		&assigneeName,
		"Open",
		[]struct{ field, from, to string }{},
	})
	resultState := m.IssueStateFromIssue(i)
	log.Println(resultState)
	// TODO: implement expectations
}

// mockIssue mocks a Jira issue. It returns the mocked `extJira.Issue` as well
// as the corresponding `store.IssueState` and `store.IssueEvent`s that are to
// be expected for this issue.
func mockIssue(def issueMockDef) *extJira.Issue {
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
	return i
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
