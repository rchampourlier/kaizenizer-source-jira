# Agilizer Source for Jira

[![Go Report Card](https://goreportcard.com/badge/github.com/rchampourlier/agilizer-source-jira)](https://goreportcard.com/report/github.com/rchampourlier/agilizer-source-jira)

## Objective

The goal of this project is to enable teams using Jira to perform analyses on Jira projects through a SQL database, enabling the use of user-friendly tools like Metabase, more powerful ones like Superset or even advanced data analysis toolsets (Python, R...).

## Use cases

### Operational

- Monitor (using a dashboard, for example in Metabase), issues with an high number of comments (e.g. > 3) and lasting more than the average cycle time.
- Monitor (using a dashboard) delivery-related KPIs, e.g.:
  - Number of issues created
  - Number of issues of type "Bug" created
  - Number of issues solved
- Send an alert when an issue is not done and has not been updated in the last 2 weeks.
- Send an alert when an issue has been sent back from review to development more than once.

### Team analytics

- Calculate team KPIs over all Jira projects, for example:
  - Mean Time Between Incidents (e.g. Blocker Bugs)
  - Cycle Time
  - Average number of issues solved per week 

## How it works

The tool will connect to Jira using the API and fetch all issues. For each issue:

- a simplified representation of the issue is stored in the `jira_issues_states` table,
- a set of events is created in the `jira_issues_events` to represent the updates that occurred on the issue (e.g. `created`, `comment_added`, `status_changed`).

The tool will perform a request to only retrieve the issues modified since the last synchronization, using the timestamp of the last event. All corresponding issues will be processed to generate new events as needed.

### Requirements

- A PostgreSQL database
- Jira username and password
- `cp .env.example` updated as necessary

### How to use

```
source .env
go run *.go sync
```

## Status

The project is currently in early stage of development (POC phase). As such, it is not usable yet and the code quality standards (e.g. testing) will not be respected yet.

## Contribution

The project is not opened to contribution at the moment.

## License

MIT

