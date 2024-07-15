# Muzz Tech Test

- [Running the Application](#running-the-application)
- [Testing the Application](#testing-the-application)
- [Endpoints](#endpoints)
- [Application Architecture](#application-architecture)
- [Code Architecture](#code-architecture)
- [Internal Packages](#internal-packages)
- [Config](#config)
- [Discover Query](#discover-query)
- [Location Data](#location-data)
- [Tools](#tools)

## Running the Application

Then to start the application, run:

```bash
./scripts/run.sh 
```

You should now be able to head to [endpoints section](#endpoints) and make some requests.

#### Local and Debugging

To run the app locally, run:

```bash
./scripts/run_dev.sh 
```

And then run the api locally using whichever IDE you prefer. This allows you to use the debugger to add break points 
etc. since this is a hassle in a container.

## Testing the Application

Before running integration tests, we must set up our dev environment by running:

```bash
./scripts/run_dev.sh
```

To run the whole test suite, we can run:

```bash
go test ./... -count=1
```

#### Integration Tests

See our integration tests [README](test/README.md) for more information.

#### Unit Tests

To run unit tests only, run with the `-short` flag.

```bash
go test ./... -count=1 -short 
```

#### Generating Mocks

We use [mockery](https://github.com/vektra/mockery) for generating our mock interfaces and run them as part of our 
`go generate./...` command. It gives us a clear and explicit way of writing expectation from dependencies in unit 
tests. See `internal/service/` tests for an example of using this, `gen.go` contains the generation logic and the 
`/mocks` package holds the mock interfaces.

## Endpoints

I have set up a [postman collection](https://www.postman.com/nickbadlose/workspace/muzz-api/collection/13188383-3d2cd57a-d0c4-43bb-ad64-0333f8a67deb?action=share&creator=13188383) 
with the list of available endpoints configured.

**Collection variables:**

*base_url* - is the host of the server to make requests against. Initial value.

*jwt_token* - is the authentication token that should be attached to requests to private endpoints. A successful postman
call to `/login` will set this value for other requests.

*true_client_ip* - is the IP address of the client making the request, postman doesn't attach one, so we have to 
manually populate this value when making requests through postman. Edit this value if you wish to change the location 
of the authenticated user when making discover requests. 

You can edit these as necessary.

#### /user/create

Creates a user record in the database, validating the request and encrypting the password on storage. 

Unlike other requests, which get the user location from their IP address, this endpoint accepts location coordinates.
This is really just for ease of use when adding user data for varying location to test, in reality I would use a 
consistent method across all requests, whether supplied by the client or retrieved from the IP.

example request:

```json
{
    "email": "<string>",
    "password": "<string>",
    "name": "<string>",
    "gender": "<string>",
    "age": "<integer>",
    "location": {
        "longitude": "<float>",
        "latitude": "<float>"
    }
}
```

example response:

```json
{
    "result": {
        "id": "<integer>",
        "email": "<string>",
        "password": "<string>",
        "name": "<string>",
        "gender": "<string>",
        "age": "<integer>",
        "location": {
            "longitude": "<float>",
            "latitude": "<float>"
        }
    }
}
```

I would not return a password from this request, however it was in the specs.

#### /login

Login authenticates the user against our database user records, on successful login, we update the users current 
location in our database, using the user IP address attached to the request.

example request:

```json
{
    "email": "<string>",
    "password": "<string>"
}
```

example response:

```json
{
    "token": "<string>"
}
```

#### /discover

Discover returns a list of users that are potential matches for the authenticated user. The distanceFromMe field is 
calculated using the authenticated users IP address as their location from the stored location values of the users in 
the database. Distance is returned in kilometres.

To filter results, we can pass query parameters:

E.g. `/discover?sort=attractiveness&maxAge=30&minAge=20&genders=male,unspecified,female`

Available filters:

1. sort: <distance|attractiveness>
2. masAge: <int>
3. minAge: <int>
4. genders: <male|female|unspecified>

example response:

```json
{
    "result": {
        "id": "<integer>",
        "name": "<string>",
        "gender": "<string>",
        "age": "<integer>",
        "distanceFromMe": "<float>"
    }
}
```

#### /swipe

Swipe performs a swipe action against the given user, returning a match if both users have stated a swipe preference 
of true.

example request:

```json
{
    "userID": "<int>",
    "preference": "<bool>"
}
```

example response: 

```json
{
    "result": {
        "matched": "<bool>",
        "matchID": "<integer>"
    }
}
```

## Application Architecture

Monolith architecture currently, whilst microservices bring a lot of benefits, they also bring a lot of initial 
overhead, slowing development speed and increasing complexity. There is no need to add this complexity from the get go, 
we can migrate to a more scalable architecture when we need to scale.

Code should be structured with scalability in mind, so when we grow as a company, we can decide to migrate to more 
scalable architectures such as microservice with minimal impact, there should be clear boundaries for us to build our 
services from. We could easily isolate all user logic into a single microservice for example.

#### Database Choice

For storing the main data, the initial thought process is SQL vs NoSQL, the specs provided have stated 3 clear domain 
entities, Users, Swipes and Matches. Where swipes and matches both relate to the user. So the schema is almost laid 
out already and is relational data. 

The specs declare we can search users by filtering related swipes, sorting by either distance or attractiveness, 
which SQL handles all of this well.

If there were `LIKE` based searches that I could see in the specs. Then I would start to consider a NoSQL options, 
particularly elasticsearch is great for these types of searches, whereas SQL performs pretty poorly in comparison for 
these. So another reason not to use NoSQL right at this point in time. 

There are no payment items currently, but if we wish to monetize the app, we will want SQL databases to utilise 
the data integrity and consistency provided by transactions. So for now, even though there is an argument for 
availability over consistency, it seems a safe bet to use SQL over NoSQL. 

I have decided to use PostgreSQL, as it offers all the functionality we need, the PostGis extension offers 
excellent geospatial querying and since we will be doing a lot of both read and write operations, Postgres handles both
of these well with good performance.

#### Caching

For caching we use redis, it's superfast and is ideal for use as an in memory cache. Allowing us to speed up any time 
heavy processes, such as requests to external services like the location IP service.

## Code Architecture

Monorepo with a layered design, domain logic at the top level (`muzz` package), interface logic in `api` package, 
service logic in `internal/service` and database logic in `internal/database/...` etc.

**Dependency inversion and the adapter pattern:**

A good example of this is the io.Reader interface. The dependant package specifies the interface it requires. For 
example, our `internal/service` service structs define the repository interfaces they need to use for storing data. 
This allows us to:

- Completely rewrite one layer of the application with zero impact in any other layers. (we could switch to MongoDB 
  and only the code in the `internal/database` package would change. The service, interface and domain layers would 
  be unaware.)
- Unit testing is extremely easy to do thoroughly, we can mock our interfaces using the `go generate ./...` command
  and isolate unit tests to the current layer only. 
- It boosts maintainability, readability and development speed, especially in the long run by setting up our code like 
  this.

## Internal Packages

I generally create packages for reusable code, such as database, cache and logger logic. Even though the packages are 
only really being used in the two places right now (monolith), main and integrations tests. We may need to use them 
elsewhere, such as scripts or if we migrate to microservices, we can easily reuse this code.

Benefits:

- Reusable code.
- We can decorate the package in one place, such as adding extra methods (`logger.MaybeError`), tracing and debug modes.
- Test the package in one place, so we know it's safe to use elsewhere.
- Isolate the external dependency to one place, if wrapping another lib. So if we need to migrate away, the impact is 
  minimal.

The functional options pattern: I often like to configure these reusable packages using this pattern if viable, as it's 
a flexible, customisable and maintainable pattern going forward. It allows using a factory method to set defaults and 
add fields without breaking changes to the client, which is really nice. 

Some packages are not necessarily reusable but are useful for structure and abstracting away all related logic 
into one place, such as the auth and location packages. This removes clutter from the main application and makes code 
more readable.

#### Tracer

We use jaeger for tracing, to view your traces, run the app, make a request and go to http://localhost:16686/search and
click `Find Traces`. You can see the all the traced steps of the request.

When `DEBUG_ENABLED` is set to true, all request and response information at the boundary of each layer is traced,
with arguments and parameters also. This obviously shouldn't be set to true for production, however for local
development and tests servers, it can be extremely useful for debugging. Especially when using microservices.

Try making a request with `DEBUG_ENABLED` set to both true and false and see the difference in the tags on the given
spans. If you make a `/login` request the following queries will have valuable tags on them:
- HTTP POST /login
- reddis.command/GET
- reddis.command/SETEX
- sql:query
- sql:exec

You should be able to see how useful this would be for debugging issues. Trace IDs are returned on the response header 
under the `Otel-Trace-Id` header. This trace ID is logged in any request related logs. 

#### Database 

We use an SQL builder package, [upper/db](https://upper.io/v4/) to build SQL queries. I prefer using an SQL builder for building queries 
especially when filtering is involved as we can cleanly use logic to build the query. It also means we can freely 
migrate between any of the SQL variants supported by the lib without breaking changes.

The sub-package `adapter` is where we adapt our database to the interface required by our `internal/service` package.

#### Log 

Wrapping the [uber zap logger](https://github.com/uber-go/zap) package. Its generally just a logger I have used in the 
past and provides everything I feel I need from a logger.

It also allows us to present logger a global, which for logging I prefer globals, as it means we can log in any deeply 
nested places with ease and without parameter drilling or attaching multiple methods to the struct like so.

With composite logger:
```go
package mypackage

import "go.uber.org/zap"

type Server struct {
	logger *zap.Logger
}

func (s *Server) surfaceLevelFunc() {
	nestedFunc(s.logger)
} 

func nestedFunc(l *zap.Logger) {
	furtherNestedFunc(l)
}

func furtherNestedFunc(l *zap.Logger) {
	// we want to log in here
	l.Error(...)
}
```

With global logger:
```go
// no need to parameter drill logger
func furtherNestedFunc() {
	// we want to log in here
	logger.Error(...)
}
```

We want to make logging as easy as possible for the user, the more difficult it is, the less likely they are to log 
any valuable information in pesky places.

## Config

Usually I would not push an env file to version control, however for this project it makes some sense, given it isn't
a production project and will stop me having to provide env vars.

If you wish to edit any env vars, go to the `config/default.env` and edit them accordingly in there. This env file is
used as the source of our docker containers environment variables.

> **NOTE** If we change the database or cache credentials in the default.env file after a docker volume has been
> created, restarting the containers will update the env vars, but not the auth. You will need to clear the volumes
> to do this by running `./scripts/clear_data.sh`. Then you should be good to go.


## Discover Query

Attractiveness adds a lot of overhead to the initial query, our attractiveness sorting algorithm:

total_preferred_swipes / total_swipes
 - total_preferred_swipes - swipes on the user where preferred = true
 - total_swipes - total swipes on the user of either preference

This gives us an attractiveness percentage, between 0 and 1.

To start out I just wrote this query in the way seemed to make the most logical sense to me:

```sql
SELECT "u"."id", "u"."name", "u"."gender", "u"."age",
       (u.location::geography <-> ST_SetSRID(ST_MakePoint(-2.244644,53.4808),4326)::geography) / 1000 AS distance,
       NULLIF((SELECT COUNT(swiped_user_id) FROM swipe WHERE swiped_user_id = u.id AND preference = true),0)::float / (SELECT COUNT(swiped_user_id) FROM swipe WHERE swiped_user_id = u.id)::float AS attractiveness
FROM "user" AS "u" WHERE (u.id != 1 AND u.id NOT IN (SELECT swiped_user_id FROM swipe WHERE user_id = 1))
ORDER BY "attractiveness" DESC;
```

Simply get all the swipes per user and all the swipes per user where preference = true in two separate subqueries and 
calculate our attractiveness value.

Whilst fine for tests, this doesn't look viable for production code, as I think it's doing two full 'swipe' table scans 
per user row returned, so I will investigate the performance of the queries by seeding test data and analysing the 
queries. [This legend](https://gis.stackexchange.com/a/247131) has provided me with a formula to seed random locations 
across the world, the only other real issue was avoiding unique constraints in the swipe table. See the seeding test 
migrations file [here](scripts/go/migrate/migrations/seed_test_data/1_seed_random_data.up.sql).

Let's seed 10000 dummy users with swipe data and run the query to see if any performance optimisation is needed at all, 
before prematurely optimising:

Query:
```sql
SELECT "u"."id", "u"."name", "u"."gender", "u"."age",
       (u.location::geography <-> ST_SetSRID(ST_MakePoint(-2.244644,53.4808),4326)::geography) / 1000 AS distance,
       NULLIF((SELECT COUNT(swiped_user_id) FROM swipe WHERE swiped_user_id = u.id AND preference = true),0)::float / (SELECT COUNT(swiped_user_id) FROM swipe WHERE swiped_user_id = u.id)::float AS attractiveness
FROM "user" AS "u" WHERE (u.id != 1 AND u.id NOT IN (SELECT swiped_user_id FROM swipe WHERE user_id = 1))
ORDER BY "attractiveness" DESC;
```

`EXPLAIN ANALYSE` output:
```
Sort  (cost=17588999.47..17589005.61 rows=2455 width=36) (actual time=46756.375..46756.627 rows=4904 loops=1)
"  Sort Key: (((NULLIF((SubPlan 1), 0))::double precision / ((SubPlan 2))::double precision)) DESC"
  Sort Method: quicksort  Memory: 512kB
"  ->  Seq Scan on ""user"" u  (cost=4.82..17588861.23 rows=2455 width=36) (actual time=653.382..46752.757 rows=4904 loops=1)"
"        Filter: ((id <> 1) AND (NOT (hashed SubPlan 3)) AND (age >= 25) AND (age <= 40) AND (gender = ANY ('{male,female,unspecified}'::text[])))"
        Rows Removed by Filter: 5096
        SubPlan 1
          ->  Aggregate  (cost=3582.00..3582.01 rows=1 width=8) (actual time=5.025..5.025 rows=1 loops=4904)
                ->  Seq Scan on swipe  (cost=0.00..3581.98 rows=10 width=4) (actual time=0.515..5.023 rows=10 loops=4904)
                      Filter: (preference AND (swiped_user_id = u.id))
                      Rows Removed by Filter: 199988
        SubPlan 2
          ->  Aggregate  (cost=3582.03..3582.04 rows=1 width=8) (actual time=4.374..4.374 rows=1 loops=4904)
                ->  Seq Scan on swipe swipe_1  (cost=0.00..3581.98 rows=20 width=4) (actual time=0.224..4.372 rows=20 loops=4904)
                      Filter: (swiped_user_id = u.id)
                      Rows Removed by Filter: 199978
        SubPlan 3
          ->  Index Only Scan using unique_swiped_user_per_user on swipe swipe_2  (cost=0.42..4.77 rows=20 width=4) (actual time=13.310..13.318 rows=20 loops=1)
                Index Cond: (user_id = 1)
                Heap Fetches: 0
Planning Time: 40.149 ms
JIT:
  Functions: 33
"  Options: Inlining true, Optimization true, Expressions true, Deforming true"
"  Timing: Generation 6.929 ms, Inlining 57.465 ms, Optimization 266.883 ms, Emission 272.362 ms, Total 603.639 ms"
Execution Time: 46852.974 ms
```

Execution time - 46 seconds, with just 10000 users, yikes :( as suspected, this is not viable for production code.

From the plan we can see that roughly 25 seconds (~5*4904) was spent on subquery 1 - `Aggregate  (cost=3582.00..3582.01 rows=1 
width=8) (actual time=5.025..5.025 rows=1 loops=4904)` and roughly 20 seconds (~4.3*4904) was spent on subquery 2 
`Aggregate  (cost=3582.03..3582.04 rows=1 width=8) (actual time=4.374..4.374 rows=1 loops=4904)`. So since that consists 
of ~99% of the time taken, that is where we can focus our attention for optimising the query.

By utilising a JOIN statement with grouping and logical aggregates, we should be able to get our desired 
total_preferred_swipes and total_swipes values with a single JOIN. 

Updated query:
```sql
SELECT
    "u"."id", "u"."name", "u"."gender", "u"."age",
    (u.location <-> ST_SetSRID(ST_MakePoint(-2.244644,53.4808),4326)) / 1000 AS distance
FROM "user" AS "u" LEFT JOIN swipe AS "s" ON u.id = s.swiped_user_id
WHERE (u.id != 1 AND u.id NOT IN (SELECT swiped_user_id FROM swipe WHERE user_id = 1) AND u.age >= 20 AND u.age <= 30 AND "u"."gender" IN ('male', 'female', 'unspecified'))
GROUP BY u.id
ORDER BY (NULLIF(sum(case when s.preference then 1 else 0 end),0)::float / COUNT(u.id)::float) DESC LIMIT 50;
```

Updated `EXPLAIN ANALYSE` output:
```
Limit  (cost=4866.87..4880.62 rows=50 width=36) (actual time=114.972..115.008 rows=50 loops=1)
  ->  Result  (cost=4866.87..6241.87 rows=5000 width=36) (actual time=114.969..115.002 rows=50 loops=1)
        ->  Sort  (cost=4866.87..4879.37 rows=5000 width=60) (actual time=114.955..114.959 rows=50 loops=1)
"              Sort Key: (((NULLIF(sum(CASE WHEN s.preference THEN 1 ELSE 0 END), 0))::double precision / (count(u.id))::double precision)) DESC"
              Sort Method: top-N heapsort  Memory: 35kB
              ->  HashAggregate  (cost=4600.77..4700.77 rows=5000 width=60) (actual time=112.550..113.866 rows=9979 loops=1)
                    Group Key: u.id
                    Batches: 1  Memory Usage: 2193kB
                    ->  Hash Right Join  (cost=461.32..3888.27 rows=95000 width=53) (actual time=8.449..68.642 rows=189553 loops=1)
                          Hash Cond: (s.swiped_user_id = u.id)
                          ->  Seq Scan on swipe s  (cost=0.00..2927.99 rows=189999 width=5) (actual time=0.012..12.896 rows=189999 loops=1)
                          ->  Hash  (cost=398.82..398.82 rows=5000 width=52) (actual time=8.415..8.416 rows=9979 loops=1)
                                Buckets: 16384 (originally 8192)  Batches: 1 (originally 1)  Memory Usage: 944kB
"                                ->  Seq Scan on ""user"" u  (cost=4.82..398.82 rows=5000 width=52) (actual time=0.094..5.091 rows=9979 loops=1)"
                                      Filter: ((id <> 1) AND (NOT (hashed SubPlan 1)))
                                      Rows Removed by Filter: 21
                                      SubPlan 1
                                        ->  Index Only Scan using unique_swiped_user_per_user on swipe  (cost=0.42..4.77 rows=20 width=4) (actual time=0.046..0.055 rows=20 loops=1)
                                              Index Cond: (user_id = 1)
                                              Heap Fetches: 0
Planning Time: 0.853 ms
Execution Time: 115.240 ms
```

115 ms, a huge improvement from 46 seconds. Let's try with 1000000 users, to check if it is really viable for 
production as 10000 users is still a pretty small amount. 

`EXPLAIN ANALYSE` output:
```
Limit  (cost=1061839.73..1061853.48 rows=50 width=37) (actual time=5605.415..5654.843 rows=50 loops=1)
  ->  Result  (cost=1061839.73..1199339.73 rows=500000 width=37) (actual time=5449.376..5498.800 rows=50 loops=1)
        ->  Sort  (cost=1061839.73..1063089.73 rows=500000 width=61) (actual time=5449.201..5498.592 rows=50 loops=1)
"              Sort Key: (((NULLIF(sum(CASE WHEN s.preference THEN 1 ELSE 0 END), 0))::double precision / (count(u.id))::double precision)) DESC"
              Sort Method: top-N heapsort  Memory: 36kB
              ->  Finalize GroupAggregate  (cost=911055.28..1045230.09 rows=500000 width=61) (actual time=4902.223..5384.951 rows=999979 loops=1)
                    Group Key: u.id
                    ->  Gather Merge  (cost=911055.28..1027730.09 rows=1000000 width=69) (actual time=4901.923..5149.877 rows=1184458 loops=1)
                          Workers Planned: 2
                          Workers Launched: 2
                          ->  Sort  (cost=910055.25..911305.25 rows=500000 width=69) (actual time=4791.287..4844.261 rows=394819 loops=3)
                                Sort Key: u.id
                                Sort Method: external merge  Disk: 29984kB
                                Worker 0:  Sort Method: external merge  Disk: 29976kB
                                Worker 1:  Sort Method: external merge  Disk: 34936kB
                                ->  Partial HashAggregate  (cost=755837.88..842216.33 rows=500000 width=69) (actual time=4005.100..4710.823 rows=394819 loops=3)
                                      Group Key: u.id
                                      Planned Partitions: 16  Batches: 17  Memory Usage: 8337kB  Disk Usage: 404504kB
                                      Worker 0:  Batches: 17  Memory Usage: 8337kB  Disk Usage: 440520kB
                                      Worker 1:  Batches: 17  Memory Usage: 8337kB  Disk Usage: 440408kB
                                      ->  Parallel Hash Right Join  (cost=35289.02..315743.18 rows=4166577 width=54) (actual time=1637.000..2734.343 rows=6666510 loops=3)
                                            Hash Cond: (s.swiped_user_id = u.id)
                                            ->  Parallel Seq Scan on swipe s  (cost=0.00..191440.54 rows=8333154 width=5) (actual time=0.068..415.209 rows=6666667 loops=3)
                                            ->  Parallel Hash  (cost=30649.86..30649.86 rows=208333 width=53) (actual time=670.853..670.870 rows=333326 loops=3)
                                                  Buckets: 131072 (originally 131072)  Batches: 16 (originally 8)  Memory Usage: 6624kB
"                                                  ->  Parallel Seq Scan on ""user"" u  (cost=5.86..30649.86 rows=208333 width=53) (actual time=393.912..443.086 rows=333326 loops=3)"
                                                        Filter: ((id <> 1) AND (NOT (hashed SubPlan 1)))
                                                        Rows Removed by Filter: 7
                                                        SubPlan 1
                                                          ->  Index Only Scan using unique_swiped_user_per_user on swipe  (cost=0.44..5.68 rows=71 width=4) (actual time=13.038..13.045 rows=20 loops=3)
                                                                Index Cond: (user_id = 1)
                                                                Heap Fetches: 0
Planning Time: 0.989 ms
JIT:
  Functions: 111
"  Options: Inlining true, Optimization true, Expressions true, Deforming true"
"  Timing: Generation 21.997 ms, Inlining 153.070 ms, Optimization 696.329 ms, Emission 605.248 ms, Total 1476.645 ms"
Execution Time: 5674.014 ms
```

5.6 seconds which is pretty heavy, but definitely viable, so we can tackle that optimisation when we are celebrating 
1 million users :) This query will definitely suffice for the time being. Also adding filters will improve the 
execution time as we filter the users before joining with the swipe table.

I've also added a LIMIT of 50 despite not being in the specs, since returning all users is bad, the nice thing about 
the discover endpoint is that pagination isn't necessary since we exclude swiped users, it basically self paginates if 
you simply add a limit. In reality, this limit would be client provided.

> **NOTE** - if you wish to test queries with more users, you can edit this [file](scripts/go/migrate/migrations/seed_test_data/1_seed_random_data.up.sql)
to use another number instead of the 10000 users we are using for analysis, the golang-migrate library we are using for
migrations doesn't support query parameters, so you will need to manually edit it.

A gotcha for analysing the queries is data is cached for subsequent runs, so we need to reset and re-seed the entire
database before each `EXPLAIN ANALYSE`.

To do this, run:
```bash
./scripts/clean_db.sh true 
```

## Location Data

To calculate distances, we need to store location data, there were two obvious options to me for retrieving this data, 
either the client would get the data and send it in requests, like our `/user/create` endpoint does, or our api can 
retrieve the location data from the request IP address, like our `/login` and `/discover` endpoints do.

I would lean towards handling it on the server via the request IP address, as it isn't much hassle and we can easily 
cache requests like we have done to reduce any overhead. Consistency is key though and I would use the same method 
for all endpoints, I have only decided to send the location data with the `/user/create` endpoint for ease of 
adding random data locations since this isn't a production app.

We use http://api.ipstack.com to get location data as it is a popular service which has a free tier for our purposes.

## Tools

A quick reference for any tools we are using.

- Traces are hosted at http://localhost:16686/search. Make some requests and check out the tags.
- [Postman collection](https://www.postman.com/nickbadlose/workspace/muzz-api/collection/13188383-3d2cd57a-d0c4-43bb-ad64-0333f8a67deb?action=share&creator=13188383)
  for making requests against the api. See the [endpoints section](#endpoints) for more information.
- Adminer is hosted at http://localhost:8080/?pgsql=db&username=muzz_admin&db=muzz for viewing the database. Log in 
  and do what you need.
- Migration tool is hosted [here.](scripts/go/migrate/README.md)
- Scripts are hosted [here.](scripts/README.md)
- For linting we use [golangci-lint](https://golangci-lint.run/welcome/install/#local-installation). 
  `golangci-lint run` to run it. The config file is in the `.golangci.yml`. See [the docs](https://golangci-lint.run/usage/configuration)
  for available configurations.