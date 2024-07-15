# Scripts

A list of the available scripts to run. If they haven't already been given execution permissions, you can do so 
by running:

```bash
chmod +x <path-to-script>
```

## migrate

See the migrate [README](go/migrate/README.md).

## run.sh

Runs the entire application with dependencies in docker and then runs application migrations. The api image is a fresh
build and dependency containers are restarted.

Migrations are also ran against the configured database.

## run_dev.sh

Tears down the `api` container if it is running and just runs the application dependencies in docker followed by the 
migrations.

This is useful for running local tests and debugging by running the api through your IDE.

## clean_run.sh

Tears down all containers, destroys all volumes and restarts all containers, use with caution, all database 
and cache data will be lost.

## clean_db.sh

Tears down database, destroys its volume and restarts database, use with caution, all database data will be lost.

parameters:
- seed: <boolean=false>. Flag to seed the database with dummy data.

example usage: 

```bash
./scripts/clean_db.sh true
```

## clean_data.sh

Tears down database and cache, destroys its volume and restarts them, use with caution, all data will be lost.

## clean_db.sh

Tears down cache, destroys its volume and restarts cache, use with caution, all cache data will be lost.