package jira

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Jeffail/tunny"

	"github.com/rchampourlier/agilizer-source-jira/store"
)

// PerformIncrementalSync fetches only newly updated issues and performs
// the same processing as `PerformSync` on each issue.
//
// ### Implementation
//
// - Updated issue are fetched by performing a JQL query where
//   `updated` is greater than the max of
//   `jira_issues_states.issue_updated_at`.
// - For each updated issue, the records already in the store are
//   dropped (e.g. the issue's state and events) so they can be
//   recreated.
func PerformIncrementalSync(c Client, store store.Store, poolSize int) {
	beforeSync := time.Now()
	log.Printf("Sync starting\n")

	// Using a chan of issue keys and a wait group for synchronization
	issueKeys := make(chan string, 100)

	// Using a WaitGroup to synchronize issue fetches and wait
	// until all are done (even if all searches have been done)
	var wg sync.WaitGroup

	// Initialize a pool of workers to fetch issues
	p := tunny.NewFunc(poolSize, func(key interface{}) interface{} {
		log.Printf("Take pool work for issue key %v\n", key)
		defer wg.Done()

		i := c.GetIssue(key.(string))
		store.ReplaceIssueStateAndEvents(key.(string), issueStateFromIssue(i), issueEventsFromIssue(i))
		return nil
	})
	defer p.Close()

	go func() {
		for issueKey := range issueKeys {
			log.Printf("Add pool work for issue key %s\n", issueKey)
			wg.Add(1)
			go p.Process(issueKey)
		}
	}()

	maxUpdatedAt := store.GetMaxUpdatedAt()
	q := fmt.Sprintf("updated > '%d/%d/%d %d:%d' ORDER BY created ASC",
		maxUpdatedAt.Year(),
		maxUpdatedAt.Month(),
		maxUpdatedAt.Day(),
		maxUpdatedAt.Hour(),
		maxUpdatedAt.Minute())
	c.SearchIssues(q, issueKeys)

	// Wait until all fetches are done
	wg.Wait()

	log.Printf("Sync done in %f minutes\n", time.Since(beforeSync).Minutes())
}

// PerformSync fetches issue identifiers from the attached Jira instance
// (using the Jira _searchIssues_ endpoint) and then fetches all
// issues (using the _get_ endpoint).
//
// Each fetched issue is then processed to generate `IssueState` and
// `IssueEvent` records that are stored in the application's store.
func PerformSync(c Client, store store.Store, poolSize int) {
	beforeSync := time.Now()
	log.Printf("Sync starting\n")

	// Using a chan of issue keys and a wait group for synchronization
	issueKeys := make(chan string, 100)

	// Using a WaitGroup to synchronize issue fetches and wait
	// until all are done (even if all searches have been done)
	var wg sync.WaitGroup

	// Initialize a pool of workers to fetch issues
	p := tunny.NewFunc(poolSize, func(key interface{}) interface{} {
		defer wg.Done()

		i := c.GetIssue(key.(string))
		store.ReplaceIssueStateAndEvents(key.(string), issueStateFromIssue(i), issueEventsFromIssue(i))
		return nil
	})
	defer p.Close()

	go func() {
		for issueKey := range issueKeys {
			wg.Add(1)
			go p.Process(issueKey)
		}
	}()

	c.SearchIssues("order by created ASC", issueKeys)

	// Wait until all fetches are done
	wg.Wait()

	log.Printf("Sync done in %f minutes\n", time.Since(beforeSync).Minutes())
}

// PerformSyncForIssueKey is the same as `PerformSync` but for a single
// issue specified by its key.
func PerformSyncForIssueKey(c Client, store store.Store, issueKey string) {
	beforeSync := time.Now()
	log.Printf("Sync starting\n")

	i := c.GetIssue(issueKey)
	store.ReplaceIssueStateAndEvents(issueKey, issueStateFromIssue(i), issueEventsFromIssue(i))

	log.Printf("Sync done in %f minutes\n", time.Since(beforeSync).Minutes())
}
