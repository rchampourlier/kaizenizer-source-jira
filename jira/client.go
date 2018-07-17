package jira

import (
	"github.com/andygrunwald/go-jira"
)

// Client is the interface for Jira clients used by the
// application.
//
// It currently has two implementations:
//
//   - `jira/client.APIClient`, which wraps `go-jira`'s client
//   - `jira/client.MockClient`, a mock for tests
type Client interface {
	SearchIssues(query string, issueKeys chan string)
	GetIssue(issueKey string) *jira.Issue
}
