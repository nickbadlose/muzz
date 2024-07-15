# Migrate

To set up our database for development, docker and test environments, we need to run our migrations' folder. The script 
uses the [golang-migrate](https://github.com/golang-migrate/migrate) tool to run these migrations.

## Usage

To run migrations, set up your database, configure the `default.env` file to store the credentials and then run

```bash
go run ./scripts/go/migrate/main.go 
```

## Flags

1. `db - <string>` - name of the database to run migrations against. Attempts to read the value from `DATABASE`
    environment variable, if it exists.
2. `migrationPath - <string>` - location of the migrations folder to run. Default to `./migrations`
3. `seed - <boolean>` - whether to seed the db with dummy data for testing.

## Config

Migrate scripts attempts to read environment from config env files located in the `config` directory. 
Using `default.env` for initial config.