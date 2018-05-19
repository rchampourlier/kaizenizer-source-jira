package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func insertEvent(db *sql.DB, e event) {
	query := `
	INSERT INTO jira_issues_events (
		time,
		kind,
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
		e.time,
		e.kind,
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

func openDB() *sql.DB {
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

func initDB(db *sql.DB) {
	queries := []string{
		`DROP TABLE IF EXISTS "jira_issues_events";`,
		`CREATE TABLE "jira_issues_events" (
		  "id" serial primary key not null,
		  "time" timestamp NOT NULL,
		  "kind" text NOT NULL,
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
	err := doQueries(db, queries)
	if err != nil {
		log.Fatalln(err)
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
