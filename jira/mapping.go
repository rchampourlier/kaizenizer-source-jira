package jira

import (
	extJira "github.com/andygrunwald/go-jira"
	"github.com/rchampourlier/kaizenizer-source-jira/store"
)

// Mapper transforms Jira issues into states and events
// records.
type Mapper interface {
	IssueEventsFromIssue(i *extJira.Issue) []store.IssueEvent
	IssueStateFromIssue(i *extJira.Issue) store.IssueState
}
