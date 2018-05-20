# TODO

## Ongoing

## Future improvements

### Optimization

- [ ] Pool the DB connections (with 30 workers it fails, 10 is ok)
- [ ] Synchronize transparently with a DB table switch
- [ ] Handle incremental synchronization to enable the synchronization to be performed regularly (e.g. every 10 minutes).

### Implement OAuth authentication

Replace BasicAuth authentication with OAuth.

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

