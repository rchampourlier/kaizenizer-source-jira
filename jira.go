package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andygrunwald/go-jira"
)

type jiraClient struct {
	*jira.Client
}

func newJiraClient() *jiraClient {
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USERNAME"),
		Password: os.Getenv("JIRA_PASSWORD"),
	}
	client, err := jira.NewClient(tp.Client(), "https://jobteaser.atlassian.net")
	if err != nil {
		log.Fatalln(fmt.Errorf("error in `newJiraClient`: %s", err))
	}
	return &jiraClient{client}
}

func (c *jiraClient) searchIssues(issueKeys chan string) {
	jso := jira.SearchOptions{
		MaxResults: 100,
		StartAt:    0,
	}
	for {
		pIssues, res, err := c.Issue.Search("order by created DESC", &jso)
		if err != nil {
			// TODO: instead of crashing, should handle the error and retry
			log.Fatalln(fmt.Errorf("error in `searchIssues`: %s", err))
		}
		log.Printf("Search: StartAt=%d Total=%d MaxResults=%d\n", res.StartAt, res.Total, res.MaxResults)
		jso.MaxResults = res.MaxResults
		jso.StartAt += res.MaxResults
		if len(pIssues) == 0 {
			log.Printf("Search: done\n")
			close(issueKeys)
			break
		}
		for _, pi := range pIssues {
			issueKeys <- pi.Key
		}
	}
}

func (c *jiraClient) getIssue(issueKey string) *jira.Issue {
	log.Printf("Fetching issue %s\n", issueKey)
	i, _, err := c.Issue.Get(issueKey, &jira.GetQueryOptions{
		Expand:       "names,schema,changelog",
		FieldsByKeys: true,
	})
	if err != nil {
		// TODO: instead of crashing, should handle the error and retry
		log.Fatalln(fmt.Errorf("error in `getIssue`: %s", err))
	}
	return i
}

func (c *jiraClient) exploreCustomFields(issueKey string) {
	i := newJiraClient().getIssue(issueKey)
	customFields := i.Fields.Unknowns
	for n, v := range customFields {
		fmt.Printf("%s -> %s\n", n, v)
	}
}
