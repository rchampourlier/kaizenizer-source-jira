package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andygrunwald/go-jira"
)

// APIClient represents an interface to Jira API. It embeds
// `go-jira`'s `jira.APIClient`.
type APIClient struct {
	*jira.Client
}

// NewAPIClient returns an usable `jira.client` usable to access Jira
// API. It embeds a `jira.APIClient`.
func NewAPIClient() *APIClient {
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USERNAME"),
		Password: os.Getenv("JIRA_PASSWORD"),
	}
	c, err := jira.NewClient(tp.Client(), "https://jobteaser.atlassian.net")
	if err != nil {
		log.Fatalln(fmt.Errorf("error in `NewAPIClient`: %s", err))
	}
	return &APIClient{c}
}

// SearchIssues perform a search on Jira API using the specified
// JQL `query` and sends the keys of the issues in the response
// through the `issueKeys` channel.
func (c *APIClient) SearchIssues(query string, issueKeys chan string) {
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

// GetIssue fetches the issue specified by the key from the Jira
// API using `go-jira` and returns a `jira.Issue`.
func (c *APIClient) GetIssue(issueKey string) *jira.Issue {
	i, r, err := c.Issue.Get(issueKey, &jira.GetQueryOptions{
		Expand:       "names,schema,changelog",
		FieldsByKeys: true,
	})
	if err != nil {
		// TODO: instead of crashing, should handle the error and retry
		log.Fatalln(fmt.Errorf("error in `GetIssue` for `%s`: %s -- response: %v", issueKey, err, r))
	}
	log.Printf("Fetched issue %s (updated: %s)\n", issueKey, time.Time(i.Fields.Updated))
	return i
}

// ExploreRawIssue prints the raw data fetched from Jira.
// This can be used to get the structure of an issue to
// implement new features.
func (c *APIClient) ExploreRawIssue(issueKey string) {
	i := NewAPIClient().GetIssue(issueKey)
	fmt.Printf("issue:\n")
	fmt.Println(i)
	fmt.Println("---")
	fmt.Printf("fields:\n")
	fmt.Println(i.Fields)
	fmt.Println("---")
	if i.Changelog != nil {
		fmt.Printf("changelog histories:\n")
		for j, h := range i.Changelog.Histories {
			fmt.Printf("  -- history %d (%s)\n", j, h.Created)
			for _, item := range h.Items {
				fmt.Printf("   |-> %s: %s -> %s\n", item.Field, item.FromString, item.ToString)
			}
		}
		fmt.Println("---")
	}
}

// ExploreCustomFields prints information about custom fields
// by fetching the specified issue. This can be used to
// retrieve the custom fields IDs by fetching an issue with
// identifiable values for these fields.
func (c *APIClient) ExploreCustomFields(issueKey string) {
	i := NewAPIClient().GetIssue(issueKey)
	customFields := i.Fields.Unknowns
	for n, v := range customFields {
		fmt.Printf("%s -> %s\n", n, v)
	}
}
