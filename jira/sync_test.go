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

func TestPerformIncrementalSync(t *testing.T) {
	issueKeys := []string{"PJ-1", "PJ-2", "PJ-3"}
	refTime := time.Now()
	c := client.NewMockClient(t)
	s := store.NewMockStore(t)

	s.ExpectGetMaxUpdatedAt().WillReturn(refTime)

	// Perform a search with `updated > 'max issue_updated_at'`
	expectedJiraQuery := fmt.Sprintf("updated > '%d/%d/%d %d:%d' ORDER BY created ASC", refTime.Year(), refTime.Month(), refTime.Day(), refTime.Hour(), refTime.Minute())
	c.ExpectSearchIssues(expectedJiraQuery).
		WillRespondWithIssueKeys(issueKeys)

	// for each issue in the search results
	for _, ik := range issueKeys {
		i, is, ies := mockIssue(ik, refTime)

		// perform a get issue
		c.ExpectGetIssue(ik).
			WillRespondWithIssue(i)

		// replace previous records of the issue with new
		//   state and events records
		s.ExpectReplaceIssueStateAndEvents().
			WithIssueState(is).
			WithIssueEvents(ies).
			WillReturnError(nil)
	}

	jira.PerformIncrementalSync(c, s, 10)
}

func mockIssue(issueKey string, refTime time.Time) (*extJira.Issue, *store.IssueState, []*store.IssueEvent) {
	issueFields := extJira.IssueFields{
		Type:           extJira.IssueType{Name: "Bug"},
		Project:        extJira.Project{Key: "PJ", Name: "Project"},
		Resolutiondate: extJira.Time(refTime),
		Created:        extJira.Time(refTime),
		Duedate:        extJira.Date(refTime),
		Updated:        extJira.Time(refTime),
		Description:    "description",
		Summary:        "summary",
	}
	changelog := extJira.Changelog{}
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
	ies := []*store.IssueEvent{}
	return i, is, ies
}

func strAddr(s string) *string {
	return &s
}

func timeAddr(t time.Time) *time.Time {
	return &t
}
