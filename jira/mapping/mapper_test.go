package mapping_test

import (
	"fmt"
	"testing"
	"time"

	extJira "github.com/andygrunwald/go-jira"
	"github.com/rchampourlier/golib/matchers"

	"github.com/rchampourlier/kaizenizer-source-jira/jira/mapping"
	"github.com/rchampourlier/kaizenizer-source-jira/store"
)

// This tests the `jira` package.
//
// `jira/mapping.go` is not unit tested but its tested in integration
// through these tests.

type changelogMockDef struct {
	field string
	from  string
	to    string
	time  time.Time
}

type issueMockDef struct {
	key        string
	refTime    time.Time
	assignee   *string // initial assignee, will be modified by changelogs
	status     string  // initial status, will be modified by changelogs
	changelogs []changelogMockDef
}

func TestIssueEventsFromIssue(t *testing.T) {
	assigneeName := "assignee"
	reporterName := "reporter"
	refTime := time.Now()
	m := mapping.Mapper{}

	t.Run("issue without changelog", func(t *testing.T) {
		key := "PJ-1"
		def := issueMockDef{
			key,
			refTime,
			&assigneeName,
			"Open",
			[]changelogMockDef{},
		}
		i := mockIssue(def)
		resultEvents := m.IssueEventsFromIssue(i)

		// Expects following events:
		//   - `created`
		//   - `assignee_changed` for the initial assignee
		//   - `status_changed` for the initial status
		matchers.MatchInt(t, "count of events", 3, len(resultEvents), i.Key)

		resultKinds := make([]string, 3)
		for j, e := range resultEvents {
			resultKinds[j] = e.EventKind
			matchers.MatchString(t, "event.IssueKey", key, e.IssueKey, e)

			switch e.EventKind {
			case "created":
				et := refTime.Add(-time.Hour) // event time should be issue creation time
				matchers.MatchTimeApprox(t, "event.EventTime", e.EventTime, et, 1, e)
				matchers.MatchString(t, "event.EventAuthor", e.EventAuthor, reporterName, e)

			case "status_changed":
				et := refTime.Add(-time.Hour)
				// event time is expected to be issue creation time
				// because the status_changed event is not created from
				// a changelog history but from the initial (and not
				// modified) status of the issue
				matchers.MatchTimeApprox(t, "event.EventTime", e.EventTime, et, 1, e)
				matchers.MatchStringPtr(t, "event.StatusChangeFrom", nil, e.StatusChangeFrom, e)
				matchers.MatchStringPtr(t, "event.StatusChangeTo", &def.status, e.StatusChangeTo, e)

			case "assignee_changed":
				et := refTime.Add(-time.Hour)
				// event time is expected to be issue creation time
				// because the status_changed event is not created from
				// a changelog history but from the initial (and not
				// modified) status of the issue
				matchers.MatchTimeApprox(t, "event.EventTime", e.EventTime, et, 1, e)
			}
		}
		matchers.MatchStringSlices(t, "events with kinds", []string{"created", "status_changed", "assignee_changed"}, resultKinds, i.Key)
	})

	t.Run("issue with status and assignee changelogs", func(t *testing.T) {
		key := "PJ-2"
		def := issueMockDef{
			key,
			refTime,
			nil,
			"Open",
			[]changelogMockDef{
				changelogMockDef{"assignee", "InitialAssignee", "ChangedAssignee", refTime.Add(2 * time.Hour)},
				changelogMockDef{"status", "Open", "In Dev", refTime.Add(1 * time.Hour)},
			},
		}
		i := mockIssue(def)
		resultEvents := m.IssueEventsFromIssue(i)
		resultEventsMap := groupAndSortEvents(resultEvents)

		// Expects following events:
		//   - `created`
		//   - `assignee_changed` for the initial assignee
		//   - `assignee_changed` for the changelog
		//   - `status_changed` for the initial status
		//   - `status_changed` for the changelog
		matchers.MatchInt(t, "count of events", len(resultEvents), 5, i.Key)

		// Match `created` event
		et := refTime.Add(-time.Hour) // event time should be issue creation time
		matchers.MatchInt(t, "count of `created` events", 1, len(resultEventsMap["created"]), i.Key)
		re := resultEventsMap["created"][0]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchString(t, "event.EventAuthor", re.EventAuthor, reporterName, i.Key)

		// Match `assignee_changed` events
		matchers.MatchInt(t, "count of `assignee_changed` events", 2, len(resultEventsMap["assignee_changed"]), i.Key)

		re = resultEventsMap["assignee_changed"][0]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchStringPtr(t, "event.AssigneeChangeFrom", nil, re.AssigneeChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.AssigneeChangeTo", strAddr("InitialAssignee"), re.AssigneeChangeTo, i.Key)

		et = refTime.Add(2 * time.Hour)
		re = resultEventsMap["assignee_changed"][1]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchStringPtr(t, "event.AssigneeChangeFrom", strAddr("InitialAssignee"), re.AssigneeChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.AssigneeChangeTo", strAddr("ChangedAssignee"), re.AssigneeChangeTo, i.Key)

		// Match `status_changed` events

		// Event #1
		matchers.MatchInt(t, "count of `status_changed` events", 2, len(resultEventsMap["status_changed"]), i.Key)

		et = refTime.Add(-time.Hour) // event time should be issue creation time
		re = resultEventsMap["status_changed"][0]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeFrom", nil, re.StatusChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeTo", strAddr("Open"), re.StatusChangeTo, i.Key)

		// Event #2
		et = refTime.Add(1 * time.Hour)
		re = resultEventsMap["status_changed"][1]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeFrom", strAddr("Open"), re.StatusChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeTo", strAddr("In Dev"), re.StatusChangeTo, i.Key)
	})

	t.Run("issue with multiple status changelogs and no assignee", func(t *testing.T) {
		key := "PJ-3"
		def := issueMockDef{
			key,
			refTime,
			nil,
			"Open",
			[]changelogMockDef{
				changelogMockDef{"status", "In Dev", "In Review", refTime.Add(2 * time.Hour)},
				changelogMockDef{"status", "Open", "In Dev", refTime.Add(1 * time.Hour)},
			},
		}
		i := mockIssue(def)
		resultEvents := m.IssueEventsFromIssue(i)
		resultEventsMap := groupAndSortEvents(resultEvents)

		// Expects following events:
		//   - `created`
		//   - `status_changed` for the initial status
		//   - `status_changed` for the changelog #1
		//   - `status_changed` for the changelog #2
		matchers.MatchInt(t, "count of events", len(resultEvents), 4, i.Key)

		// Match `created` event
		et := refTime.Add(-time.Hour) // event time should be issue creation time
		matchers.MatchInt(t, "count of `created` events", 1, len(resultEventsMap["created"]), i.Key)
		re := resultEventsMap["created"][0]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchString(t, "event.EventAuthor", re.EventAuthor, reporterName, i.Key)

		// Match `status_changed` events

		// Event #1
		matchers.MatchInt(t, "count of `status_changed` events", 3, len(resultEventsMap["status_changed"]), i.Key)

		et = refTime.Add(-time.Hour) // event time should be issue creation time
		re = resultEventsMap["status_changed"][0]
		matchers.MatchTimeApprox(t, "event.EventTime", re.EventTime, et, 1, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeFrom", nil, re.StatusChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeTo", strAddr("Open"), re.StatusChangeTo, i.Key)

		// Event #2
		et = refTime.Add(1 * time.Hour)
		re = resultEventsMap["status_changed"][1]
		matchers.MatchTimeApprox(t, "event.EventTime", et, re.EventTime, 1, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeFrom", strAddr("Open"), re.StatusChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeTo", strAddr("In Dev"), re.StatusChangeTo, i.Key)

		// Event #3
		et = refTime.Add(2 * time.Hour)
		re = resultEventsMap["status_changed"][2]
		matchers.MatchTimeApprox(t, "event.EventTime", et, re.EventTime, 1, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeFrom", strAddr("In Dev"), re.StatusChangeFrom, i.Key)
		matchers.MatchStringPtr(t, "event.StatusChangeTo", strAddr("In Review"), re.StatusChangeTo, i.Key)
	})
}

func TestIssueStateFromIssue(t *testing.T) {
	key := "PJ-1"
	assigneeName := "assignee"
	refTime := time.Now()
	m := mapping.Mapper{}
	def := issueMockDef{
		key,
		refTime,
		&assigneeName,
		"Open",
		[]changelogMockDef{},
	}
	i := mockIssue(def)

	resultState := m.IssueStateFromIssue(i)
	et := refTime.Add(-time.Hour)
	if !resultState.CreatedAt.Equal(et) {
		t.Errorf("expected CreatedAt to be `%s`, got `%s`", et, resultState.CreatedAt)
	}
	// TODO: implement other expectations
}

// mockIssue mocks a Jira issue. It returns the mocked `extJira.Issue` as well
// as the corresponding `store.IssueState` and `store.IssueEvent`s that are to
// be expected for this issue.
func mockIssue(def issueMockDef) *extJira.Issue {
	issueFields := extJira.IssueFields{
		Type:           extJira.IssueType{Name: "Bug"},
		Project:        extJira.Project{Key: "PJ", Name: "Project"},
		Status:         &extJira.Status{Name: def.status},
		Resolutiondate: extJira.Time(def.refTime),
		Reporter:       &extJira.User{Name: "reporter"},
		Created:        extJira.Time(def.refTime.Add(-time.Hour)),
		//Created:     extJira.Time(def.refTime),
		Duedate:     extJira.Date(def.refTime),
		Updated:     extJira.Time(def.refTime),
		Priority:    &extJira.Priority{Name: "Major"},
		Summary:     "summary",
		Description: "description",
	}
	if def.assignee != nil {
		issueFields.Assignee = &extJira.User{Name: *def.assignee}
	}
	changelog := extJira.Changelog{}
	for _, cl := range def.changelogs {
		h := extJira.ChangelogHistory{
			Author: extJira.User{
				Name: fmt.Sprintf("%s_change_author", cl.field),
			},
			Created: timeAsStr(cl.time),
			Items: []extJira.ChangelogItems{
				extJira.ChangelogItems{
					Field:      cl.field,
					FieldType:  "string",
					From:       cl.from,
					FromString: cl.from,
					To:         cl.to,
					ToString:   cl.to,
				},
			},
		}
		changelog.Histories = append(changelog.Histories, h)
	}

	i := &extJira.Issue{
		Key:       def.key,
		Fields:    &issueFields,
		Changelog: &changelog,
	}
	return i
}

func strAddr(s string) *string {
	return &s
}

func timeAddr(t time.Time) *time.Time {
	return &t
}

func timeAsStr(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-0700")
}

func groupAndSortEvents(events []store.IssueEvent) map[string][]store.IssueEvent {
	resultMap := make(map[string][]store.IssueEvent)
	for _, e := range events {
		if len(resultMap[e.EventKind]) == 0 {
			resultMap[e.EventKind] = []store.IssueEvent{e}
		} else {
			lastBeforeIdx := 0
			for _, me := range resultMap[e.EventKind] {
				if me.EventTime.After(e.EventTime) {
					break
				}
				lastBeforeIdx++
			}
			resultMap[e.EventKind] = append(append(resultMap[e.EventKind][0:lastBeforeIdx], e), resultMap[e.EventKind][lastBeforeIdx:]...)
		}
	}
	return resultMap
}
