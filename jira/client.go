package jira

import (
	"fmt"
	"log"
	"os"

	"github.com/andygrunwald/go-jira"
)

// Client represents an interface to Jira API. It embeds
// `go-jira`'s `jira.Client`.
type Client struct {
	*jira.Client
}

// NewClient returns an usable `jira.client` usable to access Jira
// API. It embeds a `jira.Client`.
func NewClient() *Client {
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USERNAME"),
		Password: os.Getenv("JIRA_PASSWORD"),
	}
	c, err := jira.NewClient(tp.Client(), "https://jobteaser.atlassian.net")
	if err != nil {
		log.Fatalln(fmt.Errorf("error in `NewClient`: %s", err))
	}
	return &Client{c}
}

func (c *Client) searchIssues(query string, issueKeys chan string) {
	jso := jira.SearchOptions{
		MaxResults: 100,
		StartAt:    0,
	}
	for {
		pIssues, res, err := c.Issue.Search(query, &jso)
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

func (c *Client) getIssue(issueKey string) *jira.Issue {
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

// ExploreCustomFields prints information about custom fields
// by fetching the specified issue. This can be used to
// retrieve the custom fields IDs by fetching an issue with
// identifiable values for these fields.
func (c *Client) ExploreCustomFields(issueKey string) {
	i := NewClient().getIssue(issueKey)
	customFields := i.Fields.Unknowns
	for n, v := range customFields {
		fmt.Printf("%s -> %s\n", n, v)
	}
}
