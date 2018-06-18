package jira

import (
	"log"
	"sync"
	"time"

	"github.com/Jeffail/tunny"

	"github.com/rchampourlier/agilizer-source-jira/db"
)

func (c *client) PerformSync(db *db.DB, poolSize int) {
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
		is := issueStateFromIssue(i)
		for _, ie := range issueEventsFromIssue(i) {
			db.InsertIssueEvent(ie, is)
		}
		db.InsertIssueState(is)
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

func (c *client) PerformSyncForIssueKey(db *db.DB, issueKey string) {
	beforeSync := time.Now()
	log.Printf("Sync starting\n")

	i := c.getIssue(issueKey)
	is := issueStateFromIssue(i)
	for _, ie := range issueEventsFromIssue(i) {
		db.InsertIssueEvent(ie, is)
	}
	db.InsertIssueState(is)

	log.Printf("Sync done in %f minutes\n", time.Since(beforeSync).Minutes())
}
