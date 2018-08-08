package store

import (
	"fmt"
	"log"
	"sync"
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
	mutex        sync.Mutex
}

// NewMockStore returns an instance of `MockStore`
func NewMockStore(t *testing.T) *MockStore {
	return &MockStore{
		T:     t,
		mutex: sync.Mutex{},
	}
}

// ReplaceIssueStateAndEvents implements the method for the mock.
//
// To match the args, use:
//   - `WithIssueKey` to match `k`
//   - `WithIssueState` to match `is`
//   - `withIssueEvents` to match `ies`
//
// This mock is used in a context where the order in which
// `ReplaceIssueStateAndEvents` is called is not important, so it
// will proceed differently for these expectations to ignore
// order.
func (m *MockStore) ReplaceIssueStateAndEvents(ik string, is IssueState, ies []IssueEvent) (err error) {
	ee := m.popExpectedReplaceIssueStateAndEventsForIssueKey(ik)
	if ee == nil {
		log.Fatalf("mock received `ReplaceIssueStateAndEvents` but no `ReplaceIssueStateAndEvents` for the issue key `%s` was found", ik)
	}

	// Verify arguments
	if ik != ee.issueKey {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with issue key `%s` but was expecting `%s`\n", ik, ee.issueKey)
	}

	// Check issue state
	if is.CreatedAt != ee.issueState.CreatedAt {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.CreatedAt=`%v` but was expecting `%v`\n", is.CreatedAt, ee.issueState.CreatedAt)
	}
	if is.UpdatedAt != ee.issueState.UpdatedAt {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.UpdatedAt=`%v` but was expecting `%v`\n", is.UpdatedAt, ee.issueState.UpdatedAt)
	}
	if is.Key != ee.issueState.Key {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.Key=`%v` but was expecting `%v`\n", is.Key, ee.issueState.Key)
	}
	if *is.Project != *ee.issueState.Project {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.Project=`%v` but was expecting `%v`\n", *is.Project, *ee.issueState.Project)
	}
	if *is.ResolvedAt != *ee.issueState.ResolvedAt {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.ResolvedAt=`%v` but was expecting `%v`\n", *is.ResolvedAt, *ee.issueState.ResolvedAt)
	}
	if *is.Summary != *ee.issueState.Summary {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.Summary=`%v` but was expecting `%v`\n", *is.Summary, *ee.issueState.Summary)
	}
	if *is.Description != *ee.issueState.Description {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with state.Description=`%v` but was expecting `%v`\n", *is.Summary, *ee.issueState.Description)
	}

	// Check count of events
	if len(ies) != len(ee.issueEvents) {
		m.Errorf("mock received `ReplaceIssueStateAndEvents` with %d events but was expecting %d (%v)\n", len(ies), len(ee.issueEvents), ies)
	} else {

		// Check events
		for i, ie := range ies {
			eie := ee.issueEvents[i]
			if !timesAlmostEqual(ie.EventTime, eie.EventTime) {
				m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].EventTime=`%v` but was expecting `%v`\n", i, ie.EventTime, eie.EventTime)
			}
			if ie.EventKind != eie.EventKind {
				m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].EventKind=`%v` but was expecting `%v`\n", i, ie.EventKind, eie.EventKind)
			}
			if ie.EventAuthor != eie.EventAuthor {
				m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].EventAuthor=`%v` but was expecting `%v`\n", i, ie.EventAuthor, eie.EventAuthor)
			}
			if ie.IssueKey != eie.IssueKey {
				m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].IssueKey=`%v` but was expecting `%v`\n", i, ie.IssueKey, eie.IssueKey)
			}
			if eie.CommentBody != nil {
				if ie.CommentBody == nil {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].CommentBody=nil but was expecting `%v`\n", i, *eie.CommentBody)
				} else if *ie.CommentBody != *eie.CommentBody {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].CommentBody=`%v` but was expecting `%v`\n", i, *ie.CommentBody, *eie.CommentBody)
				}
			}
			if eie.StatusChangeFrom != nil {
				if ie.StatusChangeFrom == nil {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].StatusChangeFrom=nil but was expecting `%v`\n", i, *eie.StatusChangeFrom)
				} else if *ie.StatusChangeFrom != *eie.StatusChangeFrom {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].StatusChangeFrom=`%v` but was expecting `%v`\n", i, *ie.StatusChangeFrom, *eie.StatusChangeFrom)
				}
			}
			if eie.StatusChangeTo != nil {
				if ie.StatusChangeTo == nil {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].StatusChangeTo=nil but was expecting `%v`\n", i, *eie.StatusChangeTo)
				} else if *ie.StatusChangeTo != *eie.StatusChangeTo {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].StatusChangeTo=`%v` but was expecting `%v`\n", i, *ie.StatusChangeTo, *eie.StatusChangeTo)
				}
			}
			if eie.AssigneeChangeFrom != nil {
				if ie.AssigneeChangeFrom == nil {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].AssigneeChangeFrom=nil but was expecting `%v`\n", i, *eie.AssigneeChangeFrom)
				} else if *ie.AssigneeChangeFrom != *eie.AssigneeChangeFrom {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].AssigneeChangeFrom=`%v` but was expecting `%v`\n", i, *ie.AssigneeChangeFrom, *eie.AssigneeChangeFrom)
				}
			}
			if eie.AssigneeChangeTo != nil {
				if ie.AssigneeChangeTo == nil {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].AssigneeChangeTo=nil but was expecting `%v`\n", i, *eie.AssigneeChangeTo)
				} else if *ie.AssigneeChangeTo != *eie.AssigneeChangeTo {
					m.Errorf("mock received `ReplaceIssueStateAndEvents` with event[%d].AssigneeChangeTo=`%v` but was expecting `%v`\n", i, *ie.AssigneeChangeTo, *eie.AssigneeChangeTo)
				}
			}
		}
	}
	return ee.err
}

// GetRestartFromUpdatedAt returns the value specified with the
// expectation.
//
// To position an expectation, use `ExpectGetRestartFromUpdatedAt(..)`
func (m *MockStore) GetRestartFromUpdatedAt(n int) *time.Time {
	e := m.popExpectation()
	if e == nil {
		m.Errorf("mock received `GetRestartFromUpdatedAt` but no expectation was set")
	}
	ee, ok := e.(*ExpectedGetRestartFromUpdatedAt)
	if !ok {
		m.Errorf("mock received `GetRestartFromUpdatedAt` but was expecting `%s`\n", e.Describe())
	}
	// Implement the necessary mocking
	return &ee.t
}

// ExpectGetRestartFromUpdatedAt sets an expectation on the
// `GetRestartFromUpdatedAt` method.
func (m *MockStore) ExpectGetRestartFromUpdatedAt() *ExpectedGetRestartFromUpdatedAt {
	e := ExpectedGetRestartFromUpdatedAt{}
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

// ExpectedGetRestartFromUpdatedAt
// -----------------------

// ExpectedGetRestartFromUpdatedAt represents an expectation for the
// `GetRestartFromUpdatedAt` method.
type ExpectedGetRestartFromUpdatedAt struct {
	t time.Time
}

// WillReturn can be used to specify which value the
// `GetRestartFromUpdatedAt` should return.
func (e *ExpectedGetRestartFromUpdatedAt) WillReturn(t time.Time) *ExpectedGetRestartFromUpdatedAt {
	e.t = t
	return e
}

// Describe describes the expectation
func (e *ExpectedGetRestartFromUpdatedAt) Describe() string {
	return fmt.Sprintf("GetRestartFromUpdatedAt")
}

// Other
// -----

func (m *MockStore) popExpectation() Expectation {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.expectations) == 0 {
		return nil
	}
	e := m.expectations[0]
	m.expectations = m.expectations[1:]
	return e
}

func (m *MockStore) popExpectedReplaceIssueStateAndEventsForIssueKey(ik string) *ExpectedReplaceIssueStateAndEvents {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.expectations) == 0 {
		return nil
	}

	for i, e := range m.expectations {
		ee, ok := e.(*ExpectedReplaceIssueStateAndEvents)
		if ok {
			// Check issue key matches
			if ik == ee.issueKey {
				if i == 0 {
					m.expectations = m.expectations[1:]
				} else if i == len(m.expectations)-1 {
					m.expectations = m.expectations[0:i]
				} else {
					m.expectations = append(m.expectations[0:i], m.expectations[i+1:]...)
				}
				return ee
			}
		}
	}

	return nil
}

func timesAlmostEqual(t1 time.Time, t2 time.Time) bool {
	return t1.Sub(t2) < time.Millisecond
}
