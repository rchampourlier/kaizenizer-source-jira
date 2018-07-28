package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/rchampourlier/agilizer-source-jira/jira"
	"github.com/rchampourlier/agilizer-source-jira/jira/client"
	"github.com/rchampourlier/agilizer-source-jira/store"
)

const poolSize = 10

// MaxOpenConns defines the maximum number of open connections
// to the DB.
const MaxOpenConns = 5 // for Heroku Postgres

// Main program
//
// ### init
//
// Initializes the connected database. Drops the existing tables if
// exist and create new ones according to the necessary schema.
//
// ### sync
//
// Performs an incremental sync, only fetching issues updated after
// the maximum `updated_at` of issues already stored in the application.
//
// ### sync-full
//
// This fetches all issues from the Jira instance and generates both states
// and events to the application's store.
//
// ### sync-issue <issue key>
//
// Synchronizes only the issue specified by the passed key.
//
// ### explore-custom-fields <issue key>
//
// Displays custom field information for the issue specified by the
// entered key.
//
// ### cleanup
//
// Drops all store tables and indexes used by this source.
//
func main() {
	if len(os.Args) < 2 {
		usage()
	}

	db := openDB()
	defer db.Close()
	store := store.NewPGStore(db)

	switch os.Args[1] {

	case "init":
		store.DropTables()
		store.CreateTables()

	case "sync":
		c := client.NewAPIClient()
		jira.PerformIncrementalSync(c, store, poolSize)

	case "sync-full":
		store.DropTables()
		store.CreateTables()
		c := client.NewAPIClient()
		jira.PerformSync(c, store, poolSize)

	case "sync-issue":
		if len(os.Args) < 3 {
			usage()
		}
		c := client.NewAPIClient()
		jira.PerformSyncForIssueKey(c, store, os.Args[2])

	case "explore-custom-fields":
		if len(os.Args) < 3 {
			usage()
		}
		client.NewAPIClient().ExploreCustomFields(os.Args[2])

	case "cleanup":
		store.DropTables()

	default:
		usage()
	}
}

func usage() {
	fmt.Printf(`Usage: go run main.go <action>

Available actions:
  - init
  - sync
  - sync-full
  - sync-issue <issue-key>
  - issue-to-xml <issue-key>
  - explore-custom-fields <issue-key>
  - cleanup
`)
	os.Exit(1)
}

func openDB() *sql.DB {
	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	db.SetMaxOpenConns(MaxOpenConns)
	if err != nil {
		log.Fatalln(fmt.Errorf("error in `openDB`: %s", err))
	}
	return db
}
