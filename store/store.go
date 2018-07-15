package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Store represents the application's store and
// contains the connection to the DB.
type Store struct {
	*sql.DB
}

// NewStore returns a `Store` storing the specified DB.
// The passed DB should already be open and ready to
// receive queries.
func NewStore(db *sql.DB) *Store {
	return &Store{db}
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

// ReplaceIssueStateAndEvents replace the existing state and
// events records for the specified issue key, then inserts
// the new state and events records.
//
// The operations are performed atomically using a DB transaction.
func (s *Store) ReplaceIssueStateAndEvents(k string, is IssueState, ies []IssueEvent) (err error) {
	tx, err := s.Begin()

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	if err = dropAllForIssueKey(tx, k); err != nil {
		return
	}
	if err = insertIssueState(tx, is); err != nil {
		return
	}
	if err = insertIssueEvents(tx, ies, is); err != nil {
		return
	}

	return
}

// GetMaxUpdatedAt returns the max value of `issue_updated_at`
// from the `jira_issues_states` table.
func (s *Store) GetMaxUpdatedAt() *time.Time {
	var maxUpdatedAt time.Time
	err := s.QueryRow("SELECT MAX(issue_updated_at) FROM jira_issues_states").Scan(&maxUpdatedAt)
	switch {
	case err != nil:
		log.Fatal(err)
	default:
		return &maxUpdatedAt
	}
	return nil
}

// CreateTables creates the `jira_issues_events` and
// `jira_issues_states` tables used by this
// application.
func (s *Store) CreateTables() {
	queries := []string{
		`CREATE TABLE "jira_issues_states" (
			"id" SERIAL PRIMARY KEY NOT NULL,
			"inserted_at" TIMESTAMP(6) NOT NULL DEFAULT statement_timestamp(),
			"issue_created_at" TIMESTAMP NOT NULL,
			"issue_updated_at" TIMESTAMP NOT NULL,
			"issue_key" TEXT NOT NULL,
			"issue_project" TEXT NOT NULL,
			"issue_status" TEXT NOT NULL,
			"issue_resolved_at" TIMESTAMP,
			"issue_priority" TEXT NOT NULL,
			"issue_summary" TEXT NOT NULL,
			"issue_description" TEXT,
			"issue_type" TEXT NOT NULL,
			"issue_labels" TEXT,
			"issue_assignee" TEXT,
			"issue_developer_backend" TEXT,
			"issue_developer_frontend" TEXT,
			"issue_reviewer" TEXT,
			"issue_product_owner" TEXT,
			"issue_bug_cause" TEXT,
			"issue_epic" TEXT,
			"issue_tribe" TEXT,
			"issue_components" TEXT,
			"issue_fix_versions" TEXT
		);`,
		`CREATE TABLE "jira_issues_events" (
			"id" serial primary key not null,
			"inserted_at" TIMESTAMP(6) NOT NULL DEFAULT statement_timestamp(),
			"event_time" TIMESTAMP NOT NULL,
			"event_kind" TEXT NOT NULL,
			"event_author" TEXT NOT NULL,
			"issue_created_at" TIMESTAMP NOT NULL,
			"issue_updated_at" TIMESTAMP NOT NULL,
			"issue_key" TEXT NOT NULL,
			"issue_project" TEXT NOT NULL,
			"issue_status" TEXT NOT NULL,
			"issue_resolved_at" TIMESTAMP,
			"issue_priority" TEXT NOT NULL,
			"issue_summary" TEXT NOT NULL,
			"issue_description" TEXT,
			"issue_type" TEXT NOT NULL,
			"issue_labels" TEXT,
			"issue_assignee" TEXT,
			"issue_developer_backend" TEXT,
			"issue_developer_frontend" TEXT,
			"issue_reviewer" TEXT,
			"issue_product_owner" TEXT,
			"issue_bug_cause" TEXT,
			"issue_epic" TEXT,
			"issue_tribe" TEXT,
			"issue_components" TEXT,
			"issue_fix_versions" TEXT,
			"comment_body" TEXT,
			"status_change_from" TEXT,
			"status_change_to" TEXT
		);`,
	}
	err := s.exec(queries)
	if err != nil {
		log.Fatalln(fmt.Errorf("error in `Reset`: %s", err))
	}
}

// Drop drops the tables used by this source
// (`jira_issues_events` and `jira_issues_states`)
func (s *Store) DropTables() {
	queries := []string{
		`DROP TABLE IF EXISTS "jira_issues_states";`,
		`DROP TABLE IF EXISTS "jira_issues_events";`,
	}
	err := s.exec(queries)
	if err != nil {
		log.Fatalln(fmt.Errorf("error in `Drop()`: %s", err))
	}
}

// insertIssueEvents inserts the specified events in the store in
// the passed transaction. The passed `IssueState` is used to enrich
// the event records.
func insertIssueEvents(tx *sql.Tx, ies []IssueEvent, is IssueState) (err error) {
	for _, ie := range ies {
		if err = insertIssueEvent(tx, ie, is); err != nil {
			return err
		}
	}
	return
}

// insertIssueEvent inserts an issue event in the store through
// the specified transaction
func insertIssueEvent(tx *sql.Tx, ie IssueEvent, is IssueState) (err error) {
	query := `
	INSERT INTO jira_issues_events (
		event_time,
		event_kind,
		event_author,
		comment_body,
		status_change_from,
		status_change_to,
		issue_key,
		issue_created_at,
		issue_updated_at,
		issue_project,
		issue_status,
		issue_resolved_at,
		issue_priority,
		issue_summary,
		issue_description,
		issue_type,
		issue_labels,
		issue_assignee,
		issue_developer_backend,
		issue_developer_frontend,
		issue_reviewer,
		issue_product_owner,
		issue_bug_cause,
		issue_epic,
		issue_tribe,
		issue_components,
		issue_fix_versions
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27);
	`

	_, err = tx.Exec(
		query,
		ie.EventTime,
		ie.EventKind,
		ie.EventAuthor,
		ie.CommentBody,
		ie.StatusChangeFrom,
		ie.StatusChangeTo,
		ie.IssueKey,
		is.CreatedAt,
		is.UpdatedAt,
		is.Project,
		is.Status,
		is.ResolvedAt,
		is.Priority,
		is.Summary,
		is.Description,
		is.Type,
		is.Labels,
		is.Assignee,
		is.DeveloperBackend,
		is.DeveloperFrontend,
		is.Reviewer,
		is.ProductOwner,
		is.BugCause,
		is.Epic,
		is.Tribe,
		is.Components,
		is.FixVersions,
	)
	return
}

// insertIssueState inserts a new `IssueState` record in the store within
// the specified transaction
func insertIssueState(tx *sql.Tx, is IssueState) (err error) {
	query := `
	INSERT INTO jira_issues_states (
		issue_created_at,
		issue_updated_at,
		issue_key,
		issue_project,
		issue_status,
		issue_resolved_at,
		issue_priority,
		issue_summary,
		issue_description,
		issue_type,
		issue_labels,
		issue_assignee,
		issue_developer_backend,
		issue_developer_frontend,
		issue_reviewer,
		issue_product_owner,
		issue_bug_cause,
		issue_epic,
		issue_tribe,
		issue_components,
		issue_fix_versions
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);
	`
	_, err = tx.Exec(
		query,
		is.CreatedAt,
		is.UpdatedAt,
		is.Key,
		is.Project,
		is.Status,
		is.ResolvedAt,
		is.Priority,
		is.Summary,
		is.Description,
		is.Type,
		is.Labels,
		is.Assignee,
		is.DeveloperBackend,
		is.DeveloperFrontend,
		is.Reviewer,
		is.ProductOwner,
		is.BugCause,
		is.Epic,
		is.Tribe,
		is.Components,
		is.FixVersions,
	)
	return
}

// dropAllForIssueKey drops all records from `jira_issues_states` and
// `jira_issues_events` that match the specified issue key.
func dropAllForIssueKey(tx *sql.Tx, issueKey string) (err error) {
	_, err = tx.Exec("DELETE FROM jira_issues_events WHERE issue_key = '" + issueKey + "';")
	if err != nil {
		return
	}
	_, err = tx.Exec("DELETE FROM jira_issues_states WHERE issue_key = '" + issueKey + "';")
	return
}

// exec executes the passed SQL commands on the DB using `Exec`.
func (s *Store) exec(cmds []string) (err error) {
	for _, c := range cmds {
		_, err = s.Exec(c)
		if err != nil {
			return
		}
	}
	return
}
