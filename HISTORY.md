# History

## 2018-07-28

- Finished with unit tests.
- Added configuration to the repository to run tests using TravisCI and check coverage with Coveralls.
- Repository configured on Github to display badges and do some CI on pull request changes.
- Adding some documentation to help with usage and changes.

## 2018-07-17

Continuing to implement unit tests. Facing an issue for `sync_test.go`, because the tests seem to stop executing before the goroutines have all been executed.

- Created mocks for new interfaces `jira.Client` and `store.Store`
- Extracted shared code for store to `store/store.go`
- Renamed `store.Store` to `store.PGStore` to clarify its a Postgres backend

## 2018-07-15

**Adding unit tests**

Also performed some modifications to the structure of the code to make it more testable. For example:

- Refactored `store` to perform the deletion of previous records and insertion of new records for a given issue in a single operation, wrapped in a transaction. This simplifies the tests while making the operation more robust.
- Extracted `jira.Client` to a specific `jira/client` package, adding an interface in `jira/client.go` to enable the addition of a `MockClient` to test the synchronization code easily.

## 2018-07-14

- [x] Handle incremental synchronization to enable the synchronization to be performed regularly (e.g. every 10 minutes).
  - [x] Fetch only issues updated after max `issue_updated_at`
  - [x] Drop previous states and events for the updated issues in an incremental sync
- [ ] Store a local cache copy of issue to enable faster reprocessing when changing only the mapping --> WON'T DO
  Abandoned because storing the raw issue as JSON is not reversible (unmarshaling back to a `jira.Issue` fails, because the dates are not marshaled), so it would be too much work to cache the ra
w issue (it would require either to fetch directly from the API, without using `go-jira`, or changing the marshaling).
- [ ] Synchronize transparently with a DB table switch --> WON'T DO
  Doesn't seem necessary with the incremental sync, for now at least.

## Earlier

### Data model

- [x] For `jira_issue_events` replace `..._author` columns by a single `event_author` one
- [x] Add appropriate columns to `jira_issues_events` to simplify analysis (e.g. tribe, project, issue_type...)

### Bugs

- [x] Fix synchronization bug

Identified issue causing the bug: `DI-146`

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x12511bf]

goroutine 21 [running]:
main.issueEventsFromIssue(0xc421645080, 0xc420656346, 0x6, 0xc421645080)
        /Users/rchampourlier/dev/agilizer-source-jira/jiramapping.go:51 +0xbf
main.(*jiraClient).performSync.func1(0x1285200, 0xc42071c630, 0x0, 0x0)
        /Users/rchampourlier/dev/agilizer-source-jira/sync.go:29 +0xd6
github.com/Jeffail/tunny.(*closureWorker).Process(0xc42000e140, 0x1285200, 0xc42071c630, 0x0, 0x0)
        /Users/rchampourlier/go/src/github.com/Jeffail/tunny/tunny.go:73 +0x3d
github.com/Jeffail/tunny.(*workerWrapper).run(0xc420186660)
        /Users/rchampourlier/go/src/github.com/Jeffail/tunny/worker.go:101 +0x341
created by github.com/Jeffail/tunny.newWorkerWrapper
        /Users/rchampourlier/go/src/github.com/Jeffail/tunny/worker.go:70 +0x141
panic: worker was closed

goroutine 1346 [running]:
github.com/Jeffail/tunny.(*Pool).Process(0xc420024ac0, 0x1285200, 0xc42071c630, 0x0, 0x0)
        /Users/rchampourlier/go/src/github.com/Jeffail/tunny/tunny.go:163 +0x13b
created by main.(*jiraClient).performSync.func2
        /Users/rchampourlier/dev/agilizer-source-jira/sync.go:40 +0xcb
```

- [x] Fix DB bug (it seems it was an issue with the number of connections to the DB)

```
2018/06/17 18:50:50 error in `insertIssueEvent`: pq: SSL is not enabled on the server
exit status 1
```

### Inserting issues (state)

- [x] Fix resolutiondate parsing
- [x] Insert issues in the DB (state, not events), just like the existing source.
- [x] Add custom fields to issue states

### Misc

- [x] Fix bug, pool closed once the search is complete, without waiting the fetches to finish
- [X] Convert event's `time` to a datetime column

### Connection with AgilizerSource PostgreSQL db

- [x] DB connection
- [x] DB initialization (create tables...)
- [x] Write events to DB instead of logging

