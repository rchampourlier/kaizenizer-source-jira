package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	_ "github.com/lib/pq"
)

func (c *jiraClient) performSync(db *sql.DB) {
	beforeSync := time.Now()
	log.Printf("Sync starting\n")

	// Using a chan of issue keys and a wait group
	// for synchronization
	issueKeys := make(chan string, 100)

	// Using a WaitGroup to synchronize issue fetches and wait
	// until all are done (even if all searches have been done)
	var wg sync.WaitGroup

	// Initialize a pool of workers to fetch issues
	p := tunny.NewFunc(poolSize, func(key interface{}) interface{} {
		defer wg.Done()
		i := c.getIssue(key.(string))
		for _, ie := range issueEventsFromIssue(i) {
			insertIssueEvent(db, ie)
		}
		insertIssueState(db, issueStateFromIssue(i))
		return nil
	})
	defer p.Close()

	go func() {
		for issueKey := range issueKeys {
			wg.Add(1)
			go p.Process(issueKey)
		}
	}()

	c.searchIssues(issueKeys)

	// Wait until all fetches are done
	wg.Wait()

	log.Printf("Sync done in %f minutes\n", time.Since(beforeSync).Minutes())
}
