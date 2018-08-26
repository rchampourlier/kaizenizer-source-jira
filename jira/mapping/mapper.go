package mapping

import (
	"log"
	"sort"
	"time"

	extJira "github.com/andygrunwald/go-jira"

	"github.com/rchampourlier/kaizenizer-source-jira/store"
)

// Mapper is the implementation of the `jira.Mapper` interface.
// Using an interface and a type with methods is only used to
// enable dependency-injection for the synchronization functions
// so they can be tested in isolation from the mapping.
type Mapper struct{}

// IssueEventsFromIssue generates and returns the `IssueEvent`
// records corresponding to the passed issue.
//
// The following events are generated:
//
// - `created`: represents the issue creation
// - `status_changed`: for each status change in the issue's changelogs
// - `assignee_changed`: idem, for assignee changes
// - `comment_added`: for each comment in the issue
func (m *Mapper) IssueEventsFromIssue(i *extJira.Issue) []store.IssueEvent {
	issueEvents := make([]store.IssueEvent, 0)

	issueEvents = append(issueEvents, store.IssueEvent{
		EventTime:          time.Time(i.Fields.Created),
		EventKind:          "created",
		EventAuthor:        requiredString(reporterName(i)),
		IssueKey:           i.Key,
		CommentBody:        nil,
		StatusChangeFrom:   nil,
		StatusChangeTo:     nil,
		AssigneeChangeFrom: nil,
		AssigneeChangeTo:   nil,
	})

	if i.Fields.Comments != nil {
		for _, c := range i.Fields.Comments.Comments {
			issueEvents = append(issueEvents, store.IssueEvent{
				EventTime:        parseTime(c.Created),
				EventKind:        "comment_added",
				EventAuthor:      c.Author.Name,
				IssueKey:         i.Key,
				CommentBody:      &c.Body,
				StatusChangeFrom: nil,
				StatusChangeTo:   nil,
			})
		}
	}

	// If no assignee changelog, create a assignee_changed event with the current
	// assignee.
	// Do the same with status changed.

	hasChangelogOnStatus := false
	hasChangelogOnAssignee := false
	if i.Changelog != nil {
		for k := range i.Changelog.Histories {
			// We implement the loop using the index (k) to loop in reverse
			// order. Histories returned by Jira API are sorted by time
			// *descending* but the implementation is simpler if we
			// process them in *ascending* order.
			h := i.Changelog.Histories[len(i.Changelog.Histories)-k-1]

			for _, cli := range h.Items {
				switch cli.Field {
				case "status":
					from := cli.FromString
					to := cli.ToString
					if !hasChangelogOnStatus {
						// first changelog on status
						// => generate additional event with initial status
						issueEvents = append(issueEvents, store.IssueEvent{
							EventTime:        time.Time(i.Fields.Created),
							EventKind:        "status_changed",
							EventAuthor:      h.Author.Name,
							IssueKey:         i.Key,
							StatusChangeFrom: nil,
							StatusChangeTo:   &from,
						})
					}
					hasChangelogOnStatus = true
					issueEvents = append(issueEvents, store.IssueEvent{
						EventTime:        parseTime(h.Created),
						EventKind:        "status_changed",
						EventAuthor:      h.Author.Name,
						IssueKey:         i.Key,
						StatusChangeFrom: &from,
						StatusChangeTo:   &to,
					})

				case "assignee":
					from := cli.FromString
					to := cli.ToString
					if !hasChangelogOnAssignee {
						// first changelog on assignee
						// => generate additional event with initial assignee
						issueEvents = append(issueEvents, store.IssueEvent{
							EventTime:          time.Time(i.Fields.Created),
							EventKind:          "assignee_changed",
							EventAuthor:        h.Author.Name,
							IssueKey:           i.Key,
							AssigneeChangeFrom: nil,
							AssigneeChangeTo:   &from,
						})
					}
					hasChangelogOnAssignee = true
					issueEvents = append(issueEvents, store.IssueEvent{
						EventTime:          parseTime(h.Created),
						EventKind:          "assignee_changed",
						EventAuthor:        h.Author.Name,
						IssueKey:           i.Key,
						AssigneeChangeFrom: &from,
						AssigneeChangeTo:   &to,
					})
				default:
					continue
				}
			}
		}
	}

	// If there was no `status_changed` event created, and the issue has a status,
	// add a `status_changed` event for the initial status.
	if !hasChangelogOnStatus {
		author := "N/A"
		if rn := reporterName(i); rn != nil {
			author = *rn
		}
		issueEvents = append(issueEvents, store.IssueEvent{
			EventTime:        time.Time(i.Fields.Created),
			EventKind:        "status_changed",
			EventAuthor:      author,
			IssueKey:         i.Key,
			StatusChangeFrom: nil,
			StatusChangeTo:   &(i.Fields.Status.Name),
		})
	}

	// Do the same for the assignee.
	// NB: an issue may have no assignee.
	if !hasChangelogOnAssignee && i.Fields.Assignee != nil {
		author := "N/A"
		if rn := reporterName(i); rn != nil {
			author = *rn
		}
		issueEvents = append(issueEvents, store.IssueEvent{
			EventTime:          time.Time(i.Fields.Created),
			EventKind:          "assignee_changed",
			EventAuthor:        author,
			IssueKey:           i.Key,
			AssigneeChangeFrom: nil,
			AssigneeChangeTo:   &(i.Fields.Assignee.Name),
		})
	}

	sort.Sort(store.IssueEventsByTime(issueEvents))
	return issueEvents
}

// IssueStateFromIssue creates a `store.IssueState` from a Jira issue
func (m *Mapper) IssueStateFromIssue(i *extJira.Issue) store.IssueState {
	return store.IssueState{
		CreatedAt:         time.Time(i.Fields.Created),
		UpdatedAt:         time.Time(i.Fields.Updated),
		Key:               i.Key,
		Project:           &i.Fields.Project.Name,
		Status:            &i.Fields.Status.Name,
		ResolvedAt:        resolvedAt(i),
		Priority:          &i.Fields.Priority.Name,
		Summary:           &i.Fields.Summary,
		Description:       &i.Fields.Description,
		Type:              &i.Fields.Type.Name,
		Labels:            labels(i),
		Reporter:          reporterName(i),
		Assignee:          assigneeName(i),
		DeveloperBackend:  userNameFromCustomField(i, "customfield_10600"),
		DeveloperFrontend: userNameFromCustomField(i, "customfield_12403"),
		Reviewer:          userNameFromCustomField(i, "customfield_10601"),
		ProductOwner:      userNameFromCustomField(i, "customfield_11200"),
		BugCause:          valueFromCustomField(i, "customfield_11101"),
		Epic:              epicFromCustomField(i),
		Tribe:             valueFromCustomField(i, "customfield_12100"),
		Components:        components(i),
		FixVersions:       fixVersions(i),
	}
}

// requiredString returns a string for the specified string pointer,
// even if it's nil. In this case, returns `"N/A"`.
func requiredString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// Returns the issue's reporter name or nil if there is no
// reporter.
func reporterName(i *extJira.Issue) *string {
	if i.Fields.Reporter == nil {
		return nil
	}
	return &i.Fields.Reporter.Name
}

func assigneeName(i *extJira.Issue) *string {
	if i.Fields.Assignee == nil {
		return nil
	}
	return &i.Fields.Assignee.Name
}

func components(i *extJira.Issue) *string {
	var components string
	for _, c := range i.Fields.Components {
		components = components + c.Name
	}
	return &components
}

func labels(i *extJira.Issue) *string {
	var labels string
	for _, l := range i.Fields.Labels {
		labels = labels + l
	}
	return &labels
}

func resolvedAt(i *extJira.Issue) *time.Time {
	t := time.Time(i.Fields.Resolutiondate)
	if t.IsZero() {
		return nil
	}
	return &t
}

func fixVersions(i *extJira.Issue) *string {
	var fixVersions string
	for _, fv := range i.Fields.FixVersions {
		fixVersions = fixVersions + fv.Name
	}
	return &fixVersions
}

func userNameFromCustomField(i *extJira.Issue, field string) *string {
	cf := i.Fields.Unknowns[field]
	if cf == nil {
		return nil
	}
	name := cf.(map[string]interface{})["name"].(string)
	return &name
}

func epicFromCustomField(i *extJira.Issue) *string {
	cf := i.Fields.Unknowns["customfield_10009"]
	if cf == nil {
		return nil
	}
	e, ok := cf.(*string)
	if !ok {
		return nil
	}
	return e
}

func valueFromCustomField(i *extJira.Issue, field string) *string {
	cf := i.Fields.Unknowns[field]
	if cf == nil {
		return nil
	}
	cfMap, ok := cf.(map[string]interface{})
	if !ok {
		log.Printf("customfield `%s` / issue %s\n", field, i.Key) // to give some context
		_ = cf.(map[string]interface{})                           // to make it fail with explicit message about type issue
	}
	cfValue, ok := cfMap["value"].(string)
	if !ok {
		log.Printf("customfield `%s` / issue %s\n", field, i.Key)
		_ = cfMap["value"].(string)
	}
	return &cfValue
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", s)
	if err != nil {
		log.Fatalf("failed to parse time `%s`", s)
	}
	return t
}
