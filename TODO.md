# TODO

## Ongoing

- [ ] Fix bug, pool closed once the search is complete, without waiting the fetches to finish
- [ ] Insert issues in the DB (state, not events), just like the existing source

## Future improvements

### Implement OAuth authentication

Replace BasicAuth authentication with OAuth.

## Done 

- [X] Convert event's `time` to a datetime column

### Connection with AgilizerSource PostgreSQL db

- [x] DB connection
- [x] DB initialization (create tables...)
- [x] Write events to DB instead of logging

