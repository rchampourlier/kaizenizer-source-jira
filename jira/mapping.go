package jira

import (
	"log"
	"time"

	"github.com/andygrunwald/go-jira"

	"github.com/rchampourlier/agilizer-source-jira/store"
)

func issueEventsFromIssue(i *jira.Issue) []store.IssueEvent {
	issueEvents := make([]store.IssueEvent, 0)

	issueEvents = append(issueEvents, store.IssueEvent{
		EventTime:        time.Time(i.Fields.Created),
		EventKind:        "created",
		EventAuthor:      requiredString(reporterName(i)),
		IssueKey:         i.Key,
		CommentBody:      nil,
		StatusChangeFrom: nil,
		StatusChangeTo:   nil,
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
	if i.Changelog != nil {
		for _, h := range i.Changelog.Histories {
			for _, cli := range h.Items {
				if cli.Field != "status" {
					continue
				}
				issueEvents = append(issueEvents, store.IssueEvent{
					EventTime:        parseTime(h.Created),
					EventKind:        "status_changed",
					EventAuthor:      h.Author.Name,
					IssueKey:         i.Key,
					CommentBody:      nil,
					StatusChangeFrom: &cli.FromString,
					StatusChangeTo:   &cli.ToString,
				})
			}
		}
	}

	return issueEvents
}

func issueStateFromIssue(i *jira.Issue) store.IssueState {
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
func reporterName(i *jira.Issue) *string {
	if i.Fields.Reporter == nil {
		return nil
	}
	return &i.Fields.Reporter.Name
}

func assigneeName(i *jira.Issue) *string {
	if i.Fields.Assignee == nil {
		return nil
	}
	return &i.Fields.Assignee.Name
}

func components(i *jira.Issue) *string {
	var components string
	for _, c := range i.Fields.Components {
		components = components + c.Name
	}
	return &components
}

func labels(i *jira.Issue) *string {
	var labels string
	for _, l := range i.Fields.Labels {
		labels = labels + l
	}
	return &labels
}

func resolvedAt(i *jira.Issue) *time.Time {
	t := time.Time(i.Fields.Resolutiondate)
	if t.IsZero() {
		return nil
	}
	return &t
}

func fixVersions(i *jira.Issue) *string {
	var fixVersions string
	for _, fv := range i.Fields.FixVersions {
		fixVersions = fixVersions + fv.Name
	}
	return &fixVersions
}

func userNameFromCustomField(i *jira.Issue, field string) *string {
	cf := i.Fields.Unknowns[field]
	if cf == nil {
		return nil
	}
	name := cf.(map[string]interface{})["name"].(string)
	return &name
}

func epicFromCustomField(i *jira.Issue) *string {
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

func valueFromCustomField(i *jira.Issue, field string) *string {
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
