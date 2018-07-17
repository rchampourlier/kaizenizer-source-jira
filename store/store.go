package store

import (
	"time"
)

// Store is an interface for the application's store
type Store interface {
	ReplaceIssueStateAndEvents(k string, is IssueState, ies []IssueEvent) (err error)
	GetMaxUpdatedAt() *time.Time
	CreateTables()
	DropTables()
}

// IssueState represents the state of an issue to be stored
// in the DB.
type IssueState struct {
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Key               string
	Project           *string
	Status            *string
	ResolvedAt        *time.Time
	Priority          *string
	Summary           *string
	Description       *string
	Type              *string
	Labels            *string
	Reporter          *string
	Assignee          *string
	DeveloperBackend  *string
	DeveloperFrontend *string
	Reviewer          *string
	ProductOwner      *string
	BugCause          *string
	Epic              *string
	Tribe             *string
	Components        *string
	FixVersions       *string
}

// IssueEvent represents a change event on an issue to be stored
// in the DB.
type IssueEvent struct {
	EventTime        time.Time
	EventKind        string
	EventAuthor      string
	IssueKey         string
	CommentBody      *string
	StatusChangeFrom *string
	StatusChangeTo   *string
}
