package store

import (
	"fmt"
	"time"
)

// Store is an interface for the application's store
type Store interface {
	ReplaceIssueStateAndEvents(k string, is IssueState, ies []IssueEvent) (err error)
	GetRestartFromUpdatedAt(n int) *time.Time
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
	EventTime          time.Time
	EventKind          string
	EventAuthor        string
	IssueKey           string
	CommentBody        *string
	StatusChangeFrom   *string
	StatusChangeTo     *string
	AssigneeChangeFrom *string
	AssigneeChangeTo   *string
}

func (ie IssueEvent) String() string {
	var from, to string
	switch ie.EventKind {
	case "status_changed":
		if ie.StatusChangeFrom != nil {
			from = *ie.StatusChangeFrom
		}
		if ie.StatusChangeTo != nil {
			to = *ie.StatusChangeTo
		}
	case "assignee_changed":
		if ie.AssigneeChangeFrom != nil {
			from = *ie.AssigneeChangeFrom
		}
		if ie.AssigneeChangeTo != nil {
			to = *ie.AssigneeChangeTo
		}
	default:
		return fmt.Sprintf("<IssueEvent:%s: time=%s author=%s issueKey=%s>", ie.EventKind, ie.EventTime, ie.EventAuthor, ie.IssueKey)
	}
	return fmt.Sprintf("<IssueEvent:%s `%s` -> `%s`: time=%s author=%s issueKey=%s>", ie.EventKind, from, to, ie.EventTime, ie.EventAuthor, ie.IssueKey)
}

// IssueEventsByTime implements sort.Interface for []IssueEvent based on
// the EventTime field.
type IssueEventsByTime []IssueEvent

func (a IssueEventsByTime) Len() int           { return len(a) }
func (a IssueEventsByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a IssueEventsByTime) Less(i, j int) bool { return a[i].EventTime.Before(a[j].EventTime) }
