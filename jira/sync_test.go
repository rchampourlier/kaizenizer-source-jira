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

// TODO: expectations should be checked (ensure they have been
// verified)

type mapperMock struct{}

func (m *mapperMock) IssueEventsFromIssue(i *extJira.Issue) []store.IssueEvent {
	return []store.IssueEvent{store.IssueEvent{}}
}

func (m *mapperMock) IssueStateFromIssue(i *extJira.Issue) store.IssueState {
	return store.IssueState{}
}

func TestPerformIncrementalSync(t *testing.T) {
	refTime := time.Now()
	issueKeys := []string{"PJ-1", "PJ-2", "PJ-3"}

	c := client.NewMockClient(t)
	s := store.NewMockStore(t)

	s.ExpectGetRestartFromUpdatedAt().WillReturn(refTime)

	// Perform a search with `updated > 'max issue_updated_at'`
	expectedJiraQuery := fmt.Sprintf(
		"updated > '%d.*%d.*%d.*%d.*%d.*' ORDER BY updated ASC",
		refTime.Year(),
		refTime.Month(),
		refTime.Day(),
		refTime.Hour(),
		refTime.Minute(),
	)
	c.ExpectSearchIssues(expectedJiraQuery).WillRespondWithIssueKeys(issueKeys)

	// for each issue in the search results
	for _, k := range issueKeys {
		// perform a get issue
		c.ExpectGetIssue(k).WillRespondWithIssue(&extJira.Issue{})

		// replace previous records of the issue with new
		//   state and events records
		s.ExpectReplaceIssueStateAndEvents().
			WithIssueKey(k).
			WithIssueState(&store.IssueState{}).
			WithIssueEvents([]*store.IssueEvent{&store.IssueEvent{}}).
			WillReturnError(nil)
	}

	jira.PerformIncrementalSync(c, s, 10, &mapperMock{})
}

func TestPerformSync(t *testing.T) {
	issueKeys := []string{"PJ-1", "PJ-2", "PJ-3"}

	c := client.NewMockClient(t)
	s := store.NewMockStore(t)

	// Perform a search with `updated > 'max issue_updated_at'`
	c.ExpectSearchIssues("ORDER BY updated ASC").WillRespondWithIssueKeys(issueKeys)

	// for each issue in the search results
	for _, k := range issueKeys {
		// perform a get issue
		c.ExpectGetIssue(k).WillRespondWithIssue(&extJira.Issue{})

		// replace previous records of the issue with new
		//   state and events records
		s.ExpectReplaceIssueStateAndEvents().
			WithIssueKey(k).
			WithIssueState(&store.IssueState{}).
			WithIssueEvents([]*store.IssueEvent{&store.IssueEvent{}}).
			WillReturnError(nil)
	}

	jira.PerformSync(c, s, 10, &mapperMock{})
}

func TestPerformSyncForIssueKey(t *testing.T) {
	k := "PJ-1"

	c := client.NewMockClient(t)
	s := store.NewMockStore(t)

	c.ExpectGetIssue(k).WillRespondWithIssue(&extJira.Issue{})

	// replace previous records of the issue with new
	//   state and events records
	s.ExpectReplaceIssueStateAndEvents().
		WithIssueKey(k).
		WithIssueState(&store.IssueState{}).
		WithIssueEvents([]*store.IssueEvent{&store.IssueEvent{}}).
		WillReturnError(nil)

	jira.PerformSyncForIssueKey(c, s, k, &mapperMock{})
}

func timeAsStr(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-0700")
}
