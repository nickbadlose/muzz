# muzz
Muzz tech test

## TODO
- makefile for docker up / restart db and clear data. Migrations should be run in these too
- Postgis vs mysql https://stackoverflow.com/a/22576304/22675476
- Using geography type https://postgis.net/docs/manual-3.4/using_postgis_dbmanagement.html#PostGIS_GeographyVSGeometry
- Postgis over posrtgres as global distances are curved and not plane. Postgis handles these calculations
- https://postgis.net/docs/using_postgis_dbmanagement.html#Create_Geography_Tables
- Get long/lat on FE vs BE:
  - FE will only have to get once and can send on all subsequent queries. BE will need to get from IP on each request, can cache but still not ideal.
  - State in README that I would probably discuss with peers to make decision on where I get it
- GitHub actions with tests and push to dockerhub / ecr if available
- https://postgis.net/workshops/postgis-intro/
- Looks like IP is coming from ISP, may be better to just pass it in, in general, and use IP of request as backup?
- add limit, no need for pagination as we exclude swiped users so will always want the first x amount.
- List of tools
- postgis gis index for distance calculations, see docs
- Think of edge cases for discover query, if any exist
- Added a limit as responses for many users were too large
- State validations are business logic.
- TODO docs for generating mocks
- Document how we would break app into separate sections as it grows, user section, with user subrouter and handlers, then eventually it's own microservice
- Do some sort of docs, if README or swagger or something else
- Test New funcs throughout
- Check the specs before sending.
- Make sure we can set up from scratch and run using docker only, not goland.
- Check all make functions and use correct methods ie len or cap with append or [i]
- Use sql for transactions for match records, delete on cascade for user records and related data, even other users data such as swipes and matches.
- Document providing IP in postman and manually providing for create user, so they don't all use your IP address.

## Postman

[Postman collection](https://www.postman.com/nickbadlose/workspace/muzz-api/collection/13188383-3d2cd57a-d0c4-43bb-ad64-0333f8a67deb?action=share&creator=13188383)

## Linting

To run the linter, you need to install [golangci-lint](https://golangci-lint.run/welcome/install/#local-installation).

To run the linter, run:

```bash
golangci-lint run 
```

Linter configurations can be edited to suit desired project needs in `.golangci.yml`. See [here](https://golangci-lint.run/usage/configuration) 
for available configurations.

## Database Package

We use an SQL builder package, [upper/db](https://upper.io/v4/) to build SQL queries. I prefer using an SQL builder for
building queries especially when filtering is involved as we can easily use logic to build the query. It also means 
we can freely migrate between any of the SQL variants supported by the lib without breaking changes.

TODO use this following reasoning for general package info in README, it's not relevant to just DB package, it just 
has an example with the migrate script utilising it.

Having a package may seem extreme for a monorepo, but it allows us to isolate testing to the package and re-use 
in multiple places, such as our `cmd/main` package and `scripts/go/migrate.go` package. Also, if we are successful 
and need to start migrating to microservices from this monorepo, we can easily initialise our DB  in microservices with 
a call to `database.New` with the peace of mind of using fully tested code, rather than rewriting all the 
connection code. 

It also allows us to decorate the package with things such as debug modes for test/dev environments etc. TODO more on this...

TODO
- package level doc - doc.go
- package level tests
- state in README how unit testing is much easier with adapter pattern. We can use complicated sql mock on just repository methods and mock repositories for unit tests of service.

## Log Package

Wrapping the uber zap logger package with our own allows us to decorate logs with extra information, such as traceIDs
and decorate the package with extra methods, such as logger.MaybeError.

It also allows us to present the zap logging package as a global logger, which for logging I prefer global
loggers, as it means we can log in any deeply nested places with ease and without parameter drilling a logger. Or
without attaching multiple methods to the struct like so.

It also utilises the functional options pattern, which allows us to add configurations without breaking changes.

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

TODO
- package level doc - doc.go
- package level tests

## Discover query

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
migrations file [here](./scripts/go/migrations/seed_test_data/1_seed_random_data.up.sql).

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

**NOTE** - if you wish to test queries with more users, you can edit this [file](./scripts/go/migrations/seed_test_data/1_seed_random_data.up.sql)
to use another number instead of the 10000 users we are using for analysis, the golang-migrate library we are using for
migrations doesn't support query parameters, so you will need to manually edit it.

A gotcha for analysing the queries is data is cached for subsequent runs, so we need to reset and re-seed the entire
database before each `EXPLAIN ANALYSE`.

To do this, run:
```bash
./scripts/reset_db.sh true 
```

This will clear all database volumes and reseed test data from scratch. The `true` flag denotes to seed test data.
 
I have decorated our go `scripts/go/migrate.go` script with optional seeding, this would allow any other users to 
utilise the seeding feature for testing their own future feature queries etc. Without the hassle of figuring it out 
themselves.