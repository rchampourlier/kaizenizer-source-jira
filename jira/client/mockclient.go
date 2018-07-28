package client

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/andygrunwald/go-jira"
)

// MockClient is a mock to fake a client to Jira API. It
// implements the `jira.Client` interface.
type MockClient struct {
	*testing.T
	expectations []Expectation
	mutex        sync.Mutex
}

// Expectation is a specific interface for structs representing
// expectations for the mock. They implement a `Describe` method
// that can be used by the mock to display when there is a
// mismatch between the expected call and the call it received.
type Expectation interface {
	Describe() string
}

// NewMockClient returns a new `MockClient` with a default
// behaviour.
func NewMockClient(t *testing.T) *MockClient {
	return &MockClient{
		T:     t,
		mutex: sync.Mutex{},
	}
}

// SearchIssues fakes a search issues query to the Jira API.
// The `query` parameter is ignored. The list of issue keys
// passed when initializing the mock is sent through the
// `issueKeys` channel. When all keys have been sent, the
// channel is closed.
func (c *MockClient) SearchIssues(query string, issueKeys chan string) {
	e := c.popExpectation()
	if e == nil {
		c.Errorf("mock received `SearchIssues` but no expectation was set")
	}
	esi, ok := e.(*ExpectedSearchIssues)
	if !ok {
		c.Errorf("mock received `SearchIssues` but was expecting %s\n", e.Describe())
	}
	for _, ik := range esi.issueKeys {
		issueKeys <- ik
	}
	close(issueKeys)
}

// GetIssue fakes fetching the issue specified by its key.
// It returns a fake `jira.Issue`.
func (c *MockClient) GetIssue(issueKey string) *jira.Issue {
	e := c.popExpectation()
	if e == nil {
		msg := "mock received `GetIssue` but no expectation was set"
		log.Fatalln(msg)
	}
	ee, ok := e.(*ExpectedGetIssue)
	if !ok {
		msg := fmt.Sprintf("mock received `GetIssue` but was expecting `%s`", e.Describe())
		log.Fatalln(msg)
	}
	// Check `issueKey` matches `e.issueKey`
	return ee.issue
}

// ============
// Expectations
// ============

// SearchIssues
// ------------

// ExpectedSearchIssues is an expectation for `SearchIssues`
type ExpectedSearchIssues struct {
	query     string
	issueKeys []string
}

// ExpectSearchIssues indicates the mock should expect a call to
// `SearchIssues` with the specified query.
//
// NB: the query is matched exactly.
func (c *MockClient) ExpectSearchIssues(query string) *ExpectedSearchIssues {
	e := ExpectedSearchIssues{query: query}
	c.expectations = append(c.expectations, &e)
	return &e
}

// Describe describes the `SearchIssues` expectation
func (e *ExpectedSearchIssues) Describe() string {
	return fmt.Sprintf("SearchIssues with query `%s`", e.query)
}

// WillRespondWithIssueKeys indicates `ExpectedSearchIssues`
// expectation should send the specified issue keys when
// called.
func (e *ExpectedSearchIssues) WillRespondWithIssueKeys(issueKeys []string) {
	e.issueKeys = issueKeys
}

// GetIssue
// --------

// ExpectedGetIssue represents an expectation to receive a
// `GetIssue` call
type ExpectedGetIssue struct {
	issueKey string
	issue    *jira.Issue
}

// ExpectGetIssue indicates the mock is expected to receive a
// `GetIssue` call with the specified issue key
func (c *MockClient) ExpectGetIssue(issueKey string) *ExpectedGetIssue {
	e := ExpectedGetIssue{issueKey: issueKey}
	c.expectations = append(c.expectations, &e)
	return &e
}

// WillRespondWithIssue specified that the `ExpectedGetIssue`
// expectation should respond with the passed issue.
func (e *ExpectedGetIssue) WillRespondWithIssue(issue *jira.Issue) {
	e.issue = issue
}

// Describe describes the `GetIssue` expectation
func (e *ExpectedGetIssue) Describe() string {
	return fmt.Sprintf("ExpectedGetIssue with key `%s`", e.issueKey)
}

// Other
// -----

func (c *MockClient) popExpectation() Expectation {
	c.mutex.Lock()
	if len(c.expectations) == 0 {
		return nil
	}
	e := c.expectations[0]
	c.expectations = c.expectations[1:]
	c.mutex.Unlock()
	return e
}
