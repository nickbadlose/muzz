package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nickbadlose/muzz/config"
)

const (
	driverName = "postgres"

	defaultTimeout = 10 * time.Second
)

var (
	db            = flag.String("db", "", "database to run migrations against")
	migrationPath = flag.String("migration-path", "./migrations", "location of the migrations folder to run.")
	seed          = flag.Bool("seed", false, "whether to seed the db with dummy data for testing")
)

type (
	migration struct {
		source          string
		migrationsTable string
		message         string
	}
)

func main() {
	cfg := config.MustLoad()
	flag.Parse()

	if db == nil || *db == "" {
		*db = cfg.Database()
	}

	migrations := []*migration{
		{
			source:          *migrationPath,
			migrationsTable: "schema_migrations",
			message:         "running application migrations",
		},
	}

	if seed != nil && *seed {
		migrations = append(migrations, &migration{
			source:          "./scripts/go/migrations/seed_test_data",
			migrationsTable: fmt.Sprintf("schema_migrations_%s", "seed_test_data"),
			message:         "seeding test data",
		})
	}

	dsn := fmt.Sprintf(
		"%s://%s:%s@%s/%s?sslmode=disable",
		driverName,
		cfg.DatabaseUser(),
		cfg.DatabasePassword(),
		cfg.DatabaseHost(),
		*db,
	)

	confirmDatabaseConnection(dsn)

	for _, mig := range migrations {
		printfGreen(mig.message)

		migrationDSN := fmt.Sprintf("%s&x-migrations-table=%s", dsn, mig.migrationsTable)

		m, err := migrate.New(
			fmt.Sprintf("file://%s", mig.source),
			migrationDSN,
		)
		if err != nil {
			log.Fatalf("initializing migrator: %s", err)
		}

		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("running migrations: %s", err)
		}

		sErr, dErr := m.Close()
		if sErr != nil {
			log.Fatalf("closing migrator, source error: %s", sErr)
		}
		if dErr != nil {
			log.Fatalf("closing migrator, database error: %s", dErr)
		}
	}

	printfSuccess("migrations successfully ran against the %s database", cfg.Database())
}

// confirm whether the database is accepting connections.
func confirmDatabaseConnection(dsn string) {
	var connected = make(chan struct{})

	go func() {
		log.Println("checking if database is accepting connections...")
		conn, err := sql.Open(driverName, dsn)
		if err != nil {
			log.Fatalf("could not open database: %s", err)
		}

		for {
			err = conn.PingContext(context.Background())
			if err != nil {
				log.Println("could not ping database: ", err)
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}

		if err = conn.Close(); err != nil {
			log.Println("closing initial db check connection: ", err)
		}

		connected <- struct{}{}
	}()

	select {
	case <-time.After(defaultTimeout):
		log.Fatal("timed out whilst waiting for database to accept connections")
	case <-connected:
		printfGreen("database is accepting connections")
		return
	}
}

// printfSuccess and exit with code 0.
func printfSuccess(msg string, args ...any) {
	successMsg := fmt.Sprintf(msg, args...)
	fmt.Printf("\x1b[32;1m%s\x1b[0m\n", successMsg)
	os.Exit(0)
}

// printfGreen prints a green message to stdout.
func printfGreen(msg string, args ...any) {
	m := fmt.Sprintf(msg, args...)
	fmt.Printf("\x1b[32;1m%s\x1b[0m\n", m)
}
