# Integration tests

Our integration tests are designed to be independent. To handle the creation, destruction and seeding of the test 
database, we use [golang-migrate](https://github.com/golang-migrate/migrate).

## Running

Before running the tests, we need to set up our test environment by running:

```bash
./scripts/run_dev.sh
```

Then to run the integration tests, we can run:

```bash
go test github.com/nickbadlose/muzz/test -count=1 
```

> To skip these tests add the `-short` flag to the `go test` command, for example `go test ./... -count=10 -short`. This
is handy when wanting to run unit tests only or with multiple counts, as integration tests can take a long time.

## Tests

#### Integration tests

We have two main integration table tests:

- TestPublicRoutes: for running tests against any public endpoints that don't require authentication.
- TestPrivateRoutes: for running tests against any private endpoints that require authentication.

#### Constraint tests

These tests are designed for testing any constraints set up in migration files.

## Issues / gotchas

List any encountered issues with integration testing here to help future users debug.

#### Running multiple migration folders against the same database

Golang-migrate requires multiple schema tables to run separate migration folders against the same database, which we
need as we want to run our application migrations first to set up the DB and then seed any test data from a different 
migration folder. [Here](https://github.com/golang-migrate/migrate/issues/395#issuecomment-867133636) is how to achieve 
this.

#### Dirty database version

If you encounter any dirty database errors from the migrator, sometimes these can be due to tests ungracefully 
shutting down, you'll need to [fix the database version](https://github.com/golang-migrate/migrate/blob/master/FAQ.md#what-does-dirty-database-mean) 
to run them again without errors. 

The easiest way to force fix this, in dev and test envs at least, is to just run `./scripts/reset_db.sh` to completely 
rest the database docker service to its original state and run `go test ./... -count=1` again.

There is also the very real possibility that your up or down migrations are configured incorrectly, in which case, 
write them properly ;)

#### GeoIP requests

Since we have a free account, with limited requests per month, we mock calling this service. This unfortunately means 
we cannot just call api.NewServer to run integration tests, and we have to manually build the handlers with a 
helper function :(

#### Running tests in parallel

Despite only testing a few endpoints, our tests were taking upwards of 20 seconds per run already, so we want to be able
to speed that process up. By running with t.parallel, the run time of the test package was reduced to 5 seconds, which
is a big improvement.

There is a downside in setup complexity, all tests must create a unique DB name to connect to, since we are running in
parallel as we don't want requests to be editing the same data. Since the database is dynamic, the migration files must
be dynamically created too.
