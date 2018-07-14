package main

import (
	"fmt"
	"os"

	"github.com/rchampourlier/agilizer-source-jira/jira"
	"github.com/rchampourlier/agilizer-source-jira/store"
)

const poolSize = 10

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
// ### reset
//
// Drops all Store tables and indexes used by this source.
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
	store := store.NewStore()
	defer store.Close()

	switch os.Args[1] {

	case "init":
		store.Reset()

	case "sync":
		jira.NewClient().PerformIncrementalSync(store, poolSize)

	case "sync-full":
		store.Reset()
		jira.NewClient().PerformSync(store, poolSize)

	case "sync-issue":
		if len(os.Args) < 3 {
			usage()
		}
		jira.NewClient().PerformSyncForIssueKey(store, os.Args[2])

	case "reset":
		store.Drop()

	case "explore-custom-fields":
		if len(os.Args) < 3 {
			usage()
		}
		jira.NewClient().ExploreCustomFields(os.Args[2])

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
  - reset
  - explore-custom-fields <issue-key>
`)
	os.Exit(1)
}
