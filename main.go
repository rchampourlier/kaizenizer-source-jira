package main

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/andygrunwald/go-jira"
	_ "github.com/lib/pq"
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
// DB tables (`jira_issues_events` for now, see `db.go` for details).
//
// NB: The corresponding tables are dropped before performing the sync since
// incremental sync is not supported.
//
// ### drop-db-tables
//
// Drops the tables used by this source (`jira_issues_events`).
//
func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "init-db":
		db := openDB()
		defer db.Close()
		initDB(db)
	case "sync":
		db := openDB()
		defer db.Close()
		initDB(db) // reset of the DB before sync
		performSync(db)
	case "drop-db":
		db := openDB()
		defer db.Close()
		dropDBTables(db)
	default:
		usage()
	}
}

func usage() {
	log.Fatalln("Usage: go run main.go (init-db|drop-db|sync)")
}

func performSync(db *sql.DB) {
	c := getJiraClient()

	// Using a chan of issue keys and a wait group
	// for synchronization
	issueKeys := make(chan string, 100)

	// Using a WaitGroup to synchronize issue fetches and wait
	// until all are done (even if all searches have been done)
	var wg sync.WaitGroup

	// Initialize a pool of workers to fetch issues
	p := tunny.NewFunc(poolSize, func(key interface{}) interface{} {
		defer wg.Done()
		getIssue(c, db, key.(string))
		return nil
	})
	defer p.Close()

	go func() {
		for issueKey := range issueKeys {
			wg.Add(1)
			go p.Process(issueKey)
		}
	}()

	searchIssues(c, issueKeys)

	// Wait until all fetches are done
	wg.Wait()
}

func getJiraClient() *jira.Client {
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USERNAME"),
		Password: os.Getenv("JIRA_PASSWORD"),
	}
	client, err := jira.NewClient(tp.Client(), "https://jobteaser.atlassian.net")
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

func searchIssues(c *jira.Client, issueKeys chan string) {
	jso := jira.SearchOptions{
		MaxResults: 100,
		StartAt:    0,
	}
	for {
		pIssues, res, err := c.Issue.Search("order by created DESC", &jso)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("res: StartAt=%d Total=%d MaxResults=%d\n", res.StartAt, res.Total, res.MaxResults)
		jso.MaxResults = res.MaxResults
		jso.StartAt += res.MaxResults
		if len(pIssues) == 0 {
			log.Printf("All issues fetched\n")
			close(issueKeys)
			break
		}
		for _, pi := range pIssues {
			issueKeys <- pi.Key
		}
	}
}

func getIssue(client *jira.Client, db *sql.DB, issueKey string) {
	log.Printf("Fetching issue %s\n", issueKey)
	i, _, err := client.Issue.Get(issueKey, &jira.GetQueryOptions{
		Expand:       "names,schema,changelog",
		FieldsByKeys: true,
	})
	if err != nil {
		log.Fatalln(err)
	}
	for _, evt := range eventsFromIssue(i) {
		insertEvent(db, evt)
	}
}

type event struct {
	time               time.Time
	kind               string
	issueKey           string
	issueType          *string
	issueProject       *string
	issuePriority      *string
	issueSummary       *string
	issueReporter      *string
	issueComponents    *string
	commentName        *string
	commentBody        *string
	commentAuthor      *string
	statusChangeAuthor *string
	statusChangeFrom   *string
	statusChangeTo     *string
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", s)
	if err != nil {
		log.Fatalf("Failed to parse time `%s`", s)
	}
	return t
}

func eventsFromIssue(i *jira.Issue) []event {
	events := make([]event, 0)

	var components string
	for _, c := range i.Fields.Components {
		components = components + c.Name
	}

	events = append(events, event{
		time:               time.Time(i.Fields.Created),
		kind:               "created",
		issueKey:           i.Key,
		issueType:          &i.Fields.Type.Name,
		issueProject:       &i.Fields.Project.Name,
		issuePriority:      &i.Fields.Priority.Name,
		issueSummary:       &i.Fields.Summary,
		issueReporter:      reporterName(i),
		issueComponents:    &components,
		commentName:        nil,
		commentBody:        nil,
		commentAuthor:      nil,
		statusChangeAuthor: nil,
		statusChangeFrom:   nil,
		statusChangeTo:     nil,
	})

	if i.Fields.Comments != nil {
		for _, c := range i.Fields.Comments.Comments {
			events = append(events, event{
				time:               parseTime(c.Created),
				kind:               "status_changed",
				issueKey:           i.Key,
				issueType:          nil,
				issueProject:       nil,
				issuePriority:      nil,
				issueSummary:       nil,
				issueReporter:      nil,
				issueComponents:    nil,
				commentName:        &c.Name,
				commentBody:        &c.Body,
				commentAuthor:      &c.Author.Name,
				statusChangeAuthor: nil,
				statusChangeFrom:   nil,
				statusChangeTo:     nil,
			})
		}
	}
	if i.Changelog != nil {
		for _, h := range i.Changelog.Histories {
			for _, cli := range h.Items {
				if cli.Field != "status" {
					continue
				}
				events = append(events, event{
					time:               parseTime(h.Created),
					kind:               "status_changed",
					issueKey:           i.Key,
					issueProject:       nil,
					issuePriority:      nil,
					issueSummary:       nil,
					issueReporter:      nil,
					issueComponents:    nil,
					commentName:        nil,
					commentBody:        nil,
					commentAuthor:      nil,
					statusChangeAuthor: &h.Author.Name,
					statusChangeFrom:   &cli.FromString,
					statusChangeTo:     &cli.ToString,
				})
			}
		}
	}

	return events
}

// Returns the issue's reporter name or nil if there is no
// reporter.
func reporterName(i *jira.Issue) *string {
	if i.Fields.Reporter == nil {
		return nil
	}
	return &i.Fields.Reporter.Name
}
