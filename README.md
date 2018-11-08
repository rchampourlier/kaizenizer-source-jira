# Kaizenizer Source for Jira

[![Go Report Card](https://goreportcard.com/badge/github.com/rchampourlier/agilizer-source-jira)](https://goreportcard.com/report/github.com/rchampourlier/agilizer-source-jira)
[![Build Status](https://travis-ci.com/rchampourlier/agilizer-source-jira.svg?branch=master)](https://travis-ci.com/rchampourlier/agilizer-source-jira)
[![Coverage Status](https://coveralls.io/repos/github/rchampourlier/agilizer-source-jira/badge.svg)](https://coveralls.io/github/rchampourlier/agilizer-source-jira)

## Objective

The goal of this project is to enable teams using Jira to perform analyses on Jira projects through a SQL database, enabling the use of user-friendly tools like Metabase, more powerful ones like Superset or even advanced data analysis toolsets (Python, R...).

You can this project in conjunction with the [Kaizenizer](https://github.com/rchampourlier/kaizenizer) project for advanced metrics and visualization like lead and cycle time, cumulative flow diagrams, etc.

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
go get github.com/rchampourlier/kaizenizer-source-jira
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

### How to contribute / customize

#### Run tests

```
make test
```

#### How to change the generated state and event records

##### Add a new field to the _Jira Issue States_

- **In `store/pgstore.go`**
  - In `CreateTables(..)`, add the column for the new field to the `jira_issues_states` table.
  - In `insertIssueState(..)`, add the new value in the `INSERT`.
- **In `store/store.go`**
  - Change the `IssueState struct` to add the new field.
- **[Optional] If you want to add the field to the tests (necessary if the field is mandatory or you do some operation - e.g. mapping or conversion), in `store/mockstore.go`**
  - Update `ReplaceIssueStateAndEvents(..)` to check the value for the new field.
- **In `jira/mapping/mapper.go`**
  - Change `IssueStateFromIssue(..)` to generate the correct `store.IssueState` for your issue, adding the new field. (This is where you will do the mapping with custom fields.)
- Run the tests and fix/update as necessary.

NB: you can use the `explore-custom-fields` action on the command line to get custom fields mappings.

##### Generate new kinds of _Jira Issue Events_

For now, the following events are generated from the issue's data:

- `created`
- `comment_added`
- `status_changed`
- `assignee_changed`

If you want to add new kinds of events:

- **In `jira/mapping/mapper.go`**
  - Edit `IssueEventsFromIssue(..)` to generate your new events for each issue processed. You can see how the existing events are generated.
  - Update the corresponding tests in `jira/mapping/mapper_test.go`
- [Optional] If you need to change the _Jira Issue Events_ structure to add columns related to your new events, you can follow the instructions for _Jira Issue States_ above, there is not much difference (unless you should look for event-related functions!).
- As always, do not forget to run tests and fix/update them if necessary.

## Troubleshooting

### SSL issues with Postgres

Add `?sslmode=disable` at the end of your DB URL.

### Postgres `no pg_hba.conf entry for host`

In this case, be sure not to have `?sslmode=disable` at the end!

## Contribution

The project is not opened to contribution at the moment.

## License

MIT

