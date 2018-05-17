package main

import (
	"log"
	"os"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/andygrunwald/go-jira"
)

const poolSize = 10

func main() {
	c := getJiraClient()

	// Using a chan of issue keys and a wait group
	// for synchronization
	issueKeys := make(chan string, 100)

	// Initialize a pool of workers to fetch issues
	p := tunny.NewFunc(poolSize, func(key interface{}) interface{} {
		getIssue(c, key.(string))
		return nil
	})
	defer p.Close()

	go func() {
		for issueKey := range issueKeys {
			go p.Process(issueKey)
		}
	}()

	searchIssues(c, issueKeys)
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

func getIssue(client *jira.Client, issueKey string) {
	log.Printf("Fetching issue %s\n", issueKey)
	i, _, err := client.Issue.Get(issueKey, &jira.GetQueryOptions{
		Expand:       "names,schema,changelog",
		FieldsByKeys: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	var resolution string
	if i.Fields.Resolution != nil {
		resolution = i.Fields.Resolution.Name
	}
	var components string
	for _, c := range i.Fields.Components {
		components = components + c.Name
	}
	pushIssueCreatedEvent(IssueCreatedEvent{
		Time:       time.Time(i.Fields.Created),
		Key:        i.Key,
		Type:       i.Fields.Type.Name,
		Project:    i.Fields.Project.Name,
		Resolution: resolution,
		Priority:   i.Fields.Priority.Name,
		Summary:    i.Fields.Summary,
		Reporter:   i.Fields.Reporter.Name,
		Components: components,
	})
	if i.Fields.Comments != nil {
		for _, c := range i.Fields.Comments.Comments {
			pushIssueCommentAddedEvent(IssueCommentAddedEvent{
				Time:   c.Created,
				Key:    i.Key,
				Author: c.Author.Name,
				Name:   c.Name,
				Body:   c.Body,
			})
		}
	}
	if i.Changelog != nil {
		for _, h := range i.Changelog.Histories {
			for _, cli := range h.Items {
				if cli.Field != "status" {
					continue
				}
				pushIssueStatusChangedEvent(IssueStatusChangedEvent{
					Time:   h.Created,
					Key:    i.Key,
					Author: h.Author.Name,
					From:   cli.FromString,
					To:     cli.ToString,
				})
			}
		}
	}
}

// IssueCreatedEvent reflects the creation of a new
// issue.
type IssueCreatedEvent struct {
	Time       time.Time
	Key        string
	Type       string
	Project    string
	Resolution string
	Priority   string
	Summary    string
	Reporter   string
	Components string
}

// IssueCommentAddedEvent reflects the addition of a comment
// to the issue.
type IssueCommentAddedEvent struct {
	Time   string
	Key    string
	Author string
	Name   string
	Body   string
}

// IssueStatusChangedEvent reflects a change of status
// in the issue.
type IssueStatusChangedEvent struct {
	Time   string
	Key    string
	Author string
	From   string
	To     string
}

func pushIssueCreatedEvent(e IssueCreatedEvent) {
	log.Printf(
		"IssueCreated %v - %s - type=%s project=%s resolution=%s priority=%s summary=%s reporter=%s components=%s\n",
		e.Time,
		e.Key,
		e.Type,
		e.Project,
		e.Resolution,
		e.Priority,
		e.Summary,
		e.Reporter,
		e.Components,
	)
}

func pushIssueCommentAddedEvent(e IssueCommentAddedEvent) {
	log.Printf(
		"IssueCommentAdded %v - %s - author=%s name=%s body=%s\n",
		e.Time,
		e.Key,
		e.Author,
		e.Name,
		e.Body,
	)
}

func pushIssueStatusChangedEvent(e IssueStatusChangedEvent) {
	log.Printf(
		"IssueStatusChanged %v - %s - author=%s from=%s to=%s\n",
		e.Time,
		e.Key,
		e.Author,
		e.From,
		e.To,
	)
}
