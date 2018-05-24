# TODO

## Ongoing

---

## Future improvements

### Deployment

- [ ] Deploy and schedule synchronization

### Assignment data

- [ ] Generate events for assignee change, with status information

### Optimization

- [ ] Pool the DB connections (with 30 workers it fails, 10 is ok)
- [ ] Synchronize transparently with a DB table switch
- [ ] Handle incremental synchronization to enable the synchronization to be performed regularly (e.g. every 10 minutes).

### Source code data

- [ ] Extract Github pull requests from comments (see https://github.com/jobteaser/agilizer_source/blob/master/lib/agilizer/interface/jira/transformations/2_extract_github_pull_requests_from_comments.rb)
- [ ] Add source code enrichments (see https://github.com/jobteaser/agilizer_source/blob/master/lib/agilizer/enrichments/source_code_changes.rb)

### Implement OAuth authentication

- [ ] Replace BasicAuth authentication with OAuth.

---

## Done 

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

