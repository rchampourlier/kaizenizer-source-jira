package store

import (
	"fmt"
	"testing"
	"time"
)

// MockStore implements the `Store` interface for tests
//
// TODO: should implement `ExpectationsWereMet()` to
// verify all expectations were met.
type MockStore struct {
	*testing.T
	expectations []Expectation
}

// NewMockStore returns an instance of `MockStore`
func NewMockStore(t *testing.T) *MockStore {
	return &MockStore{T: t}
}

// ReplaceIssueStateAndEvents implements the method for
// the mock.
//
// To match the args, use:
//   - `WithIssueKey` to match `k`
//   - `WithIssueState` to match `is`
//   - `withIssueEvents` to match `ies`
func (m *MockStore) ReplaceIssueStateAndEvents(k string, is IssueState, ies []IssueEvent) (err error) {
	e := m.popExpectation()
	if e == nil {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` but no expectation was set")
	}
	ee, ok := e.(*ExpectedReplaceIssueStateAndEvents)
	if !ok {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` but was expecting `%s`\n", e.Describe())
	}

	// Verify arguments
	if k != ee.issueKey {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with issue key `%s` but was expecting `%s`\n", k, ee.issueKey)
	}
	if is != *ee.issueState {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state `%s` but was expecting `%s`\n", is, ee.issueState)
	}
	if len(ies) != len(ee.issueEvents) {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with %d events `%s` but was expecting %d\n", len(ies), len(ee.issueEvents))
	}

	return ee.err
}

// GetMaxUpdatedAt returns the value specified with the
// expectation.
//
// To position an expectation, use `ExpectGetMaxUpdatedAt(..)`
func (m *MockStore) GetMaxUpdatedAt() *time.Time {
	e := m.popExpectation()
	if e == nil {
		m.Errorf("mock received `GetMaxUpdatedAt` but no expectation was set")
	}
	ee, ok := e.(*ExpectedGetMaxUpdatedAt)
	if !ok {
		m.Errorf("mock received `GetMaxUpdatedAt` but was expecting `%s`\n", e.Describe())
	}
	// Implement the necessary mocking
	return &ee.t
}

// ExpectGetMaxUpdatedAt sets an expectation on the
// `GetMaxUpdatedAt` method.
func (m *MockStore) ExpectGetMaxUpdatedAt() *ExpectedGetMaxUpdatedAt {
	e := ExpectedGetMaxUpdatedAt{}
	m.expectations = append(m.expectations, &e)
	return &e
}

// CreateTables does nothing
func (m *MockStore) CreateTables() {
}

// DropTables does nothing
func (m *MockStore) DropTables() {
}

// ============
// Expectations
// ============

// Expectation represents an expectation for the mock
type Expectation interface {
	Describe() string
}

// ExpectReplaceIssueStateAndEvents adds a new expectation to the
// mock for the `ReplaceIssueStateAndEvents` method.
//
// Use `With...` and `Will...` methods on the returned
// `ExpectedReplaceIssueStateAndEvents` expectation to
// specify expected arguments and return value.
func (m *MockStore) ExpectReplaceIssueStateAndEvents() *ExpectedReplaceIssueStateAndEvents {
	e := ExpectedReplaceIssueStateAndEvents{}
	m.expectations = append(m.expectations, &e)
	return &e
}

// ExpectedReplaceIssueStateAndEvents
// ----------------------------------

// ExpectedReplaceIssueStateAndEvents represents an expectation for the
// `ReplaceIssueStateAndEvents` method.
type ExpectedReplaceIssueStateAndEvents struct {
	issueKey    string
	issueState  *IssueState
	issueEvents []*IssueEvent
	err         error
}

// Describe describes the expectation
func (e *ExpectedReplaceIssueStateAndEvents) Describe() string {
	return fmt.Sprintf("ReplaceIssueStateAndEvents for issue `%s`", e.issueKey)
}

// WithIssueKey specified the issue key argument the
// `ReplaceIssueStateAndEvents` method is expected
// to receive for this expectation.
func (e *ExpectedReplaceIssueStateAndEvents) WithIssueKey(ik string) *ExpectedReplaceIssueStateAndEvents {
	e.issueKey = ik
	return e
}

// WithIssueState specifies the `*IssueState` argument
// the `ReplaceIssueStateAndEvents` method is expected
// to receive for this expectation
func (e *ExpectedReplaceIssueStateAndEvents) WithIssueState(is *IssueState) *ExpectedReplaceIssueStateAndEvents {
	e.issueState = is
	return e
}

// WithIssueEvents specifies the `[]*IssueEvent` argument
// the `ReplaceIssueStateAndEvents` method is expected
// to receive for this expectation
func (e *ExpectedReplaceIssueStateAndEvents) WithIssueEvents(ies []*IssueEvent) *ExpectedReplaceIssueStateAndEvents {
	e.issueEvents = ies
	return e
}

// WillReturnError can be used to specify which value the
// `ReplaceIssueStateAndEvents` should return.
func (e *ExpectedReplaceIssueStateAndEvents) WillReturnError(err error) *ExpectedReplaceIssueStateAndEvents {
	e.err = err
	return e
}

// ExpectedGetMaxUpdatedAt
// -----------------------

// ExpectedGetMaxUpdatedAt represents an expectation for the
// `GetMaxUpdatedAt` method.
type ExpectedGetMaxUpdatedAt struct {
	t time.Time
}

// WillReturn can be used to specify which value the
// `GetMaxUpdatedAt` should return.
func (e *ExpectedGetMaxUpdatedAt) WillReturn(t time.Time) *ExpectedGetMaxUpdatedAt {
	e.t = t
	return e
}

// Describe describes the expectation
func (e *ExpectedGetMaxUpdatedAt) Describe() string {
	return fmt.Sprintf("GetMaxUpdatedAt")
}

// Other
// -----

func (m *MockStore) popExpectation() Expectation {
	if len(m.expectations) == 0 {
		return nil
	}
	e := m.expectations[0]
	m.expectations = m.expectations[1:]
	return e
}
