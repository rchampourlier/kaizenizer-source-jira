package store_test

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/rchampourlier/kaizenizer-source-jira/store"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPGStore_ReplaceIssueStateAndEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	s := store.NewPGStore(db)

	// expect transaction begin
	mock.ExpectBegin()

	// expect drop state and events
	mock.ExpectExec("DELETE FROM jira_issues_events WHERE issue_key = 'key'").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("DELETE FROM jira_issues_states WHERE issue_key = 'key'").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// expect insert state
	mock.ExpectExec("INSERT INTO jira_issues_states").WithArgs(
		anyTime{},
		anyTime{},
		"key",
		"project",
		"status",
		anyTime{},
		"priority",
		"summary",
		"description",
		"type",
		"labels",
		"assignee",
		"developer_backend",
		"developer_frontend",
		"reviewer",
		"product_owner",
		"bug_cause",
		"epic",
		"tribe",
		"components",
		"fix_versions",
	).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO jira_issues_events").WithArgs(
		anyTime{},
		"kind",
		"author",
		"comment",
		"status_from",
		"status_to",
		"assignee_from",
		"assignee_to",
		"key",
		anyTime{},
		anyTime{},
		"project",
		"status",
		anyTime{},
		"priority",
		"summary",
		"description",
		"type",
		"labels",
		"assignee",
		"developer_backend",
		"developer_frontend",
		"reviewer",
		"product_owner",
		"bug_cause",
		"epic",
		"tribe",
		"components",
		"fix_versions",
	).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = s.ReplaceIssueStateAndEvents("key", mockIssueState(), []store.IssueEvent{mockIssueEvent()})
	if err != nil {
		t.Fatalf("unexpected error in `ReplaceIssueStateAndEvents`: %s\n", err)
	}

	// Checking expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStore_GetRestartFromUpdatedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := store.NewPGStore(db)

	timeV := time.Now()
	rows := sqlmock.NewRows([]string{"max"}).
		AddRow(timeV)

	mock.ExpectQuery("SELECT MIN\\(issue_updated_at\\) FROM \\( SELECT issue_updated_at FROM jira_issues_states ORDER BY issue_updated_at DESC LIMIT \\$1 \\)").
		WillReturnRows(rows)

	r := s.GetRestartFromUpdatedAt(10)
	if *r != timeV {
		t.Errorf("unexpected result `%v`, expected `%v`\n", r, timeV)
	}
}

func TestPGStore_CreateTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE \"jira_issues_states\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE \"jira_issues_events\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := store.NewPGStore(db)
	s.CreateTables()
}

func TestPGStore_Drop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DROP TABLE IF EXISTS \"jira_issues_states\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DROP TABLE IF EXISTS \"jira_issues_events\"").
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := store.NewPGStore(db)
	s.DropTables()
}

func mockIssueState() store.IssueState {
	return store.IssueState{
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Key:               "key",
		Project:           stringAddr("project"),
		Status:            stringAddr("status"),
		ResolvedAt:        timeAddr(time.Now()),
		Priority:          stringAddr("priority"),
		Summary:           stringAddr("summary"),
		Description:       stringAddr("description"),
		Type:              stringAddr("type"),
		Labels:            stringAddr("labels"),
		Reporter:          stringAddr("reporter"),
		Assignee:          stringAddr("assignee"),
		DeveloperBackend:  stringAddr("developer_backend"),
		DeveloperFrontend: stringAddr("developer_frontend"),
		Reviewer:          stringAddr("reviewer"),
		ProductOwner:      stringAddr("product_owner"),
		BugCause:          stringAddr("bug_cause"),
		Epic:              stringAddr("epic"),
		Tribe:             stringAddr("tribe"),
		Components:        stringAddr("components"),
		FixVersions:       stringAddr("fix_versions"),
	}
}

func mockIssueEvent() store.IssueEvent {
	return store.IssueEvent{
		EventTime:          time.Now(),
		EventKind:          "kind",
		EventAuthor:        "author",
		IssueKey:           "key",
		CommentBody:        stringAddr("comment"),
		StatusChangeFrom:   stringAddr("status_from"),
		StatusChangeTo:     stringAddr("status_to"),
		AssigneeChangeFrom: stringAddr("assignee_from"),
		AssigneeChangeTo:   stringAddr("assignee_to"),
	}
}

type anyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func stringAddr(s string) *string {
	return &s
}

func timeAddr(t time.Time) *time.Time {
	return &t
}
