package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func insertIssueEvent(db *sql.DB, e issueEvent) {
	query := `
	INSERT INTO jira_issues_events (
		event_time,
		event_kind,
		issue_key,
		issue_type,
		issue_project,
		issue_priority,
		issue_summary,
		issue_reporter,
		issue_components,
		comment_name,
		comment_body,
		comment_author,
		status_change_author,
		status_change_from,
		status_change_to
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);
	`

	rows, err := db.Query(
		query,
		e.EventTime,
		e.EventKind,
		e.issueKey,
		e.issueType,
		e.issueProject,
		e.issuePriority,
		e.issueSummary,
		e.issueReporter,
		e.issueComponents,
		e.commentName,
		e.commentBody,
		e.commentAuthor,
		e.statusChangeAuthor,
		e.statusChangeFrom,
		e.statusChangeTo,
	)
	if err != nil {
		log.Fatalln(err)
	}
	rows.Close()
}

func insertIssueState(db *sql.DB, s issueState) {
	query := `
	INSERT INTO jira_issues_states (
		created_at,
		updated_at,
		key,
		project,
		status,
		resolved_at,
		priority,
		summary,
		description,
		type,
		labels,
		assignee,
		developer_backend,
		developer_frontend,
		reviewer,
		product_owner,
		bug_cause,
		epic,
		tribe,
		components,
		fix_versions
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);
	`
	rows, err := db.Query(
		query,
		s.CreatedAt,
		s.UpdatedAt,
		s.Key,
		s.Project,
		s.Status,
		s.ResolvedAt,
		s.Priority,
		s.Summary,
		s.Description,
		s.Type,
		s.Labels,
		s.Assignee,
		s.DeveloperBackend,
		s.DeveloperFrontend,
		s.Reviewer,
		s.ProductOwner,
		s.BugCause,
		s.Epic,
		s.Tribe,
		s.Components,
		s.FixVersions,
	)
	if err != nil {
		log.Fatalln(err)
	}
	rows.Close()
}

func openDB() *sql.DB {
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

func initDB(db *sql.DB) {
	queries := make([]string, 0)
	queries = append(queries, queriesForJiraIssuesEvents()...)
	queries = append(queries, queriesForJiraIssuesStates()...)
	err := doQueries(db, queries)
	if err != nil {
		log.Fatalln(err)
	}
}

func queriesForJiraIssuesEvents() []string {
	return []string{
		`DROP TABLE IF EXISTS "jira_issues_events";`,
		`CREATE TABLE "jira_issues_events" (
		  "id" serial primary key not null,
		  "event_time" timestamp NOT NULL,
		  "event_kind" text NOT NULL,
		  "issue_key" text NOT NULL,
		  "issue_type" text,
		  "issue_project" text,
		  "issue_priority" text,
		  "issue_summary" text,
		  "issue_reporter" text,
		  "issue_components" text,
		  "comment_name" text,
		  "comment_body" text,
		  "comment_author" text,
		  "status_change_author" text,
		  "status_change_from" text,
		  "status_change_to" text,
		  "inserted_at" timestamp(6) NOT NULL DEFAULT statement_timestamp()
		);`,
	}
}

func queriesForJiraIssuesStates() []string {
	return []string{
		`DROP TABLE IF EXISTS "jira_issues_states";`,
		`CREATE TABLE "jira_issues_states" (
			"id" SERIAL PRIMARY KEY NOT NULL,
			"created_at" TIMESTAMP NOT NULL,
			"updated_at" TIMESTAMP NOT NULL,
			"key" TEXT NOT NULL,
			"project" TEXT NOT NULL,
			"status" TEXT NOT NULL,
			"resolved_at" TIMESTAMP,
			"priority" TEXT NOT NULL,
			"summary" TEXT NOT NULL,
			"description" TEXT,
			"type" TEXT NOT NULL,
			"labels" TEXT,
			"assignee" TEXT,
			"developer_backend" TEXT,
			"developer_frontend" TEXT,
			"reviewer" TEXT,
			"product_owner" TEXT,
			"bug_cause" TEXT,
			"epic" TEXT,
			"tribe" TEXT,
			"components" TEXT,
			"fix_versions" TEXT,
			"inserted_at" TIMESTAMP(6) NOT NULL DEFAULT statement_timestamp()
		);`,
	}
}

func dropDBTables(db *sql.DB) {
	queries := []string{
		`DROP TABLE IF EXISTS "jira_issues_events";`,
	}
	err := doQueries(db, queries)
	if err != nil {
		log.Fatalln(err)
	}
}

// doQueries performs the specified queries on the passed db.
// If an error occurs, it returns the error. This function can't
// be used for queries where you need the result rows.
func doQueries(db *sql.DB, queries []string) error {
	for _, q := range queries {
		rows, err := db.Query(q)
		if err != nil {
			return err
		}
		defer rows.Close()
		if err = rows.Err(); err != nil {
			return err
		}
	}
	return nil
}
