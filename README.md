# Agilizer Source for Jira

[![Go Report Card](https://goreportcard.com/badge/github.com/rchampourlier/agilizer-source-jira)](https://goreportcard.com/report/github.com/rchampourlier/agilizer-source-jira)
[![Build Status](https://travis-ci.com/rchampourlier/agilizer-source-jira.svg?branch=master)](https://travis-ci.com/rchampourlier/agilizer-source-jira)
[![Coverage Status](https://coveralls.io/repos/github/rchampourlier/agilizer-source-jira/badge.svg)](https://coveralls.io/github/rchampourlier/agilizer-source-jira)

## Objective

The goal of this project is to enable teams using Jira to perform analyses on Jira projects through a SQL database, enabling the use of user-friendly tools like Metabase, more powerful ones like Superset or even advanced data analysis toolsets (Python, R...).

## Use cases

Some examples of easy requests that can be done using Metabase.

### Identify issues with many comments

![](doc/metabase-example-issues-with-many-comments.png)

### Track number of bugs created over time

![](doc/metabase-example-created-bugs-per-priority-over-time.png)

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

#### 0. Clone the repo

```
go get github.com/rchampourlier/agilizer-source-jira
```

NB: you should follow Go conventions (e.g. `GOPATH`, dependencies management, etc.) or at least have your environment working and know it well enough.

#### 1. Create your `.env` file

```
cp .env.example .env
```

Now, edit the file and set the following values:

- `JIRA_USERNAME`: a valid Jira username (you should probably have a dedicated user for this and not use your personal user)
- `JIRA_PASSWORD`: the corresponding password
- `DB_URL`: the URL to the database where you want to push Jira records to (it should look like this: `postgres://USER:PASSWORD@HOST:5432/DB_NAME`)

NB: if you face Postgres SSL-related issues, try adding `?sslmode=disable` at the end of your `DB_URL`.

#### 2. DB initialization and initial synchronization

```
source .env
go run *.go reset
```

#### 3. Incremental synchronization

```
source .env
go run *.go sync
```

_NB: the DB must have been initialized and a first synchronization done._

### How to change the generated state and event records

#### Add a new field to the _Jira Issue States_

1. In `store/pgstore.go`:
  - In `CreateTables(..)`, add the column for the new field to the `jira_issues_states` table.
  - In `insertIssueState(..)`, add the new value in the `INSERT`.
2. In `store/store.go`:
  - Change the `IssueState struct` to add the new field.
3. [Optional] If you want to add the field to the tests (necessary if the field is mandatory or you do some operation - e.g. mapping or conversion), in `store/mockstore.go`:
  - Update `ReplaceIssueStateAndEvents(..)` to check the value for the new field.
4. In `jira/mapping.go`:
  - Change `issueStateFromIssue(..)` to generate the correct `store.IssueState` for your issue, adding the new field. (This is where you will do the mapping with custom fields.)

NB: you can use the `explore-custom-fields` action on the command line to get custom fields mappings.

#### Generate new kinds of _Jira Issue Events_

For now, the following events are generated from the issue's data:

- `created`
- `comment_added`
- `status_changed`
- `assignee_changed`

If you want to add new kinds of events:

1. In `jira/mapping.go`:
  - Edit `issueEventsFromIssue(..)` to generate your new events for each issue processed. You can see how the existing events are generated.
2. [Optional] If you need to change the _Jira Issue Events_ structure to add columns related to your new events, you can follow the instructions for _Jira Issue States_ above, there is not much difference (unless you should look for event-related functions!).

## Troubleshooting

### SSL issues with Postgres

Add `?sslmode=disable` at the end of your DB URL.

## Contribution

The project is not opened to contribution at the moment.

## License

MIT

