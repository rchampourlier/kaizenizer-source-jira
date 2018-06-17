# TODO

## Ongoing

---

## Future improvements


### Deployment

- [ ] Deploy and schedule synchronization

### Assignment data

- [ ] Generate events for assignee change, with status information

### Optimization

- [ ] Store a local cache copy of issue to enable faster reprocessing when changing only the mapping
- [ ] Synchronize transparently with a DB table switch
- [ ] Handle incremental synchronization to enable the synchronization to be performed regularly (e.g. every 10 minutes).

### Source code data

- [ ] Extract Github pull requests from comments (see https://github.com/jobteaser/agilizer_source/blob/master/lib/agilizer/interface/jira/transformations/2_extract_github_pull_requests_from_comments.rb)
- [ ] Add source code enrichments (see https://github.com/jobteaser/agilizer_source/blob/master/lib/agilizer/enrichments/source_code_changes.rb)

### Implement OAuth authentication

- [ ] Replace BasicAuth authentication with OAuth.

---

## Done 

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

