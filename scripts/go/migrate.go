package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nickbadlose/muzz/config"
)

func main() {
	cfg := config.MustLoad()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.DatabaseUser(),
		cfg.DatabasePassword(),
		cfg.DatabaseHost(),
		cfg.Database(),
	)
	m, err := migrate.New(
		"file://./migrations",
		dsn,
	)
	if err != nil {
		log.Fatalf("initializing migrator: %s", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("running migrations: %s", err)
	}

	fmt.Printf("migrations successfully ran against the %s database\n", cfg.Database())
}
