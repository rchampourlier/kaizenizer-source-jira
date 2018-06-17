package main

import (
	"fmt"
	"os"
)

const poolSize = 10

// Main program
//
// ### init-db
//
// Initializes the connected database. Drops the existing tables if
// exist and create new ones according to the necessary schema.
//
// ### sync
//
// Performs the synchronization. Current strategy is a full synchronization
// by fetching all issues and inserting the appropriate rows in the
// DB tables (`jira_issues_issueEvents` for now, see `db.go` for details).
//
// NB: The corresponding tables are dropped before performing the sync since
// incremental sync is not supported.
//
// ### drop-db-tables
//
// Drops the tables used by this source (`jira_issues_issueEvents`).
//
// ### explore-custom-fields <issue key>
//
// Displays custom field information for the issue specified by the
// entered key.
//
func main() {
	if len(os.Args) < 2 {
		usage()
	}
	db := openDB()
	defer db.Close()

	switch os.Args[1] {

	case "init-db":
		initDB(db)

	case "sync":
		initDB(db) // reset of the DB before sync
		newJiraClient().performSync(db)

	case "sync-issue":
		if len(os.Args) < 3 {
			usage()
		}
		newJiraClient().performSyncForIssueKey(db, os.Args[2])

	case "drop-db":
		dropDBTables(db)

	case "explore-custom-fields":
		if len(os.Args) < 3 {
			usage()
		}
		newJiraClient().exploreCustomFields(os.Args[2])

	default:
		usage()
	}
}

func usage() {
	fmt.Printf(`Usage: go run main.go <action>

Available actions:
  - init-db
  - drop-db
  - sync
  - sync-issue <issue-key>
  - explore-custom-fields <issue-key>

`)
	os.Exit(1)
}
