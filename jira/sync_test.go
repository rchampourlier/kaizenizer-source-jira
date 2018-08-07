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

func TestPerformIncrementalSync(t *testing.T) {
	issueMocks := []struct {
		key        string
		changelogs []string
	}{
		{"PJ-1", []string{}},
		{"PJ-2", []string{"status", "Open", "In Dev"}},
		{"PJ-3", []string{"assignee", "Me", "You"}},
	}
	var issueKeys []string
	for _, issueMock := range issueMocks {
		issueKeys = append(issueKeys, issueMock.key)
	}

	refTime := time.Now()

	c := client.NewMockClient(t)
	s := store.NewMockStore(t)

	s.ExpectGetMaxUpdatedAt().WillReturn(refTime)

	// Perform a search with `updated > 'max issue_updated_at'`
	expectedJiraQuery := fmt.Sprintf("updated > '%s' ORDER BY created ASC", timeAsStr(refTime))
	c.ExpectSearchIssues(expectedJiraQuery).
		WillRespondWithIssueKeys(issueKeys)

	// for each issue in the search results
	for _, im := range issueMocks {
		i, is, ies := mockIssue(im.key, refTime, im.changelogs)

		// perform a get issue
		c.ExpectGetIssue(im.key).WillRespondWithIssue(i)

		// replace previous records of the issue with new
		//   state and events records
		s.ExpectReplaceIssueStateAndEvents().
			WithIssueKey(im.key).
			WithIssueState(is).
			WithIssueEvents(ies).
			WillReturnError(nil)
	}

	jira.PerformIncrementalSync(c, s, 10)
}

// mockIssue mocks a Jira issue. It returns the mocked `extJira.Issue` as well
// as the corresponding `store.IssueState` and `store.IssueEvent`s that are to
// be expected for this issue.
func mockIssue(issueKey string, refTime time.Time, changelogs []string) (*extJira.Issue, *store.IssueState, []*store.IssueEvent) {
	issueFields := extJira.IssueFields{
		Type:           extJira.IssueType{Name: "Bug"},
		Project:        extJira.Project{Key: "PJ", Name: "Project"},
		Status:         &extJira.Status{Name: "In Review"},
		Resolutiondate: extJira.Time(refTime),
		Created:        extJira.Time(refTime),
		Duedate:        extJira.Date(refTime),
		Updated:        extJira.Time(refTime),
		Priority:       &extJira.Priority{Name: "Major"},
		Summary:        "summary",
		Description:    "description",
	}
	changelog := extJira.Changelog{}
	if len(changelogs) > 0 {
		switch changelogs[0] { // kind of changelog
		case "assignee":
			h := extJira.ChangelogHistory{
				Author: extJira.User{
					Name: "assignee_change_author",
				},
				Created: timeAsStr(refTime),
				Items: []extJira.ChangelogItems{
					extJira.ChangelogItems{
						Field:      "assignee",
						FieldType:  "string",
						From:       changelogs[1],
						FromString: changelogs[1],
						To:         changelogs[2],
						ToString:   changelogs[2],
					},
				},
			}
			changelog.Histories = append(changelog.Histories, h)
		case "status":
			h := extJira.ChangelogHistory{
				Author: extJira.User{
					Name: "status_change_author",
				},
				Created: timeAsStr(refTime),
				Items: []extJira.ChangelogItems{
					extJira.ChangelogItems{
						Field:      "status",
						FieldType:  "string",
						From:       changelogs[1],
						FromString: changelogs[1],
						To:         changelogs[2],
						ToString:   changelogs[2],
					},
				},
			}
			changelog.Histories = append(changelog.Histories, h)
		default:
		}
	}

	i := &extJira.Issue{
		Key:       issueKey,
		Fields:    &issueFields,
		Changelog: &changelog,
	}
	is := &store.IssueState{
		CreatedAt:   refTime,
		UpdatedAt:   refTime,
		Key:         issueKey,
		Project:     strAddr("Project"),
		ResolvedAt:  timeAddr(refTime),
		Summary:     strAddr("summary"),
		Description: strAddr("description"),
	}
	ies := []*store.IssueEvent{
		{
			EventTime:   refTime,
			EventKind:   "created",
			EventAuthor: "N/A",
			// "N/A" because the issue's reporter is not set, so the required string is replaced by N/A
			IssueKey:         issueKey,
			CommentBody:      nil,
			StatusChangeFrom: nil,
			StatusChangeTo:   nil,
		},
	}
	if len(changelogs) > 0 {
		switch changelogs[0] {
		case "assignee":
			ies = append(ies, &store.IssueEvent{
				EventTime:          refTime,
				EventKind:          "assignee_changed",
				EventAuthor:        "assignee_change_author",
				IssueKey:           issueKey,
				CommentBody:        nil,
				StatusChangeFrom:   nil,
				StatusChangeTo:     nil,
				AssigneeChangeFrom: strAddr(changelogs[1]),
				AssigneeChangeTo:   strAddr(changelogs[2]),
			})
		case "status":
			ies = append(ies, &store.IssueEvent{
				EventTime:          refTime,
				EventKind:          "status_changed",
				EventAuthor:        "status_change_author",
				IssueKey:           issueKey,
				CommentBody:        nil,
				StatusChangeFrom:   strAddr(changelogs[1]),
				StatusChangeTo:     strAddr(changelogs[2]),
				AssigneeChangeFrom: nil,
				AssigneeChangeTo:   nil,
			})
		}
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
