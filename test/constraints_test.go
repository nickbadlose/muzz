package test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/nickbadlose/muzz/config"
	"github.com/stretchr/testify/require"
)

func TestConstraints(t *testing.T) {
	cases := []struct {
		name       string
		errMessage string
	}{
		{
			name:       "user unique email",
			errMessage: "duplicate key value violates unique constraint \"unique_email\", Key (email)=(test@test.com) already exists.",
		},
		{
			name:       "swipe unique swiped user per user",
			errMessage: "duplicate key value violates unique constraint \"unique_swiped_user_per_user\", Key (user_id, swiped_user_id)=(1, 2) already exists.",
		},
		{
			name:       "swipe no matching user ids",
			errMessage: "new row for relation \"swipe\" violates check constraint \"no_matching_user_ids\", Failing row contains (1, 1, 1, t).",
		},
		{
			name:       "match unique matched user per user",
			errMessage: "duplicate key value violates unique constraint \"unique_matched_user_per_user\", Key (user_id, matched_user_id)=(1, 2) already exists.",
		},
		{
			name:       "match no matching user ids",
			errMessage: "new row for relation \"match\" violates check constraint \"no_matching_user_ids\", Failing row contains (1, 1, 1).",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load()
			require.NoError(t, err)
			// TODO don't continue on errorNoChange, it means some migrations haven't run correctly?
			//  make boilerplate code reusable between tests

			// create test DB migrator to set up and teardown test db.
			// You can't drop a database whilst connections still exist, so we authenticate to the postgres DB to run these.
			createTestDBMigrator, err := migrate.New(
				"file://./migrations/create",
				fmt.Sprintf(
					"postgres://%s:%s@%s/postgres?sslmode=disable",
					cfg.DatabaseUser(),
					cfg.DatabasePassword(),
					cfg.DatabaseHost(),
				),
			)
			require.NoError(t, err)
			t.Cleanup(func() {
				err = createTestDBMigrator.Down()
				if !errors.Is(err, migrate.ErrNoChange) {
					require.NoError(t, err)
				}
				sErr, dErr := createTestDBMigrator.Close()
				require.NoError(t, sErr)
				require.NoError(t, dErr)
			})

			err = createTestDBMigrator.Up()
			if !errors.Is(err, migrate.ErrNoChange) {
				require.NoError(t, err)
			}

			// TODO run appMigrator.Drop at the start of each migration to clear all data if db already existed
			// app migrator to run all application migrations.
			appMigrator, err := migrate.New(
				"file://../migrations",
				fmt.Sprintf(
					"postgres://%s:%s@%s/test?sslmode=disable",
					cfg.DatabaseUser(),
					cfg.DatabasePassword(),
					cfg.DatabaseHost(),
				),
			)
			require.NoError(t, err)
			t.Cleanup(func() {
				err = appMigrator.Down()
				require.NoError(t, err)
				sErr, dErr := appMigrator.Close()
				require.NoError(t, sErr)
				require.NoError(t, dErr)
			})

			err = appMigrator.Up()
			if !errors.Is(err, migrate.ErrNoChange) {
				require.NoError(t, err)
			}

			// seed migrator seeds any test data into the database. Golang-migrate requires multiple schema tables to run
			// multiple separate migration folders against the same DB, so we specify a seed schema table for these,
			// &x-migrations-table=\"schema_seed_migrations\".
			// https://github.com/golang-migrate/migrate/issues/395#issuecomment-867133636
			constraintMigrator, err := migrate.New(
				fmt.Sprintf(
					"file://./migrations/constraints/%s",
					strings.ReplaceAll(tc.name, " ", "_"),
				),
				fmt.Sprintf(
					"postgres://%s:%s@%s/test?sslmode=disable&x-migrations-table=schema_constraint_migrations",
					cfg.DatabaseUser(),
					cfg.DatabasePassword(),
					cfg.DatabaseHost(),
				),
			)
			require.NoError(t, err)
			t.Cleanup(func() {
				sErr, dErr := constraintMigrator.Close()
				require.NoError(t, sErr)
				require.NoError(t, dErr)
			})

			err = constraintMigrator.Up()
			require.Error(t, err)
			require.Contains(
				t,
				err.Error(),
				tc.errMessage,
			)
		})
	}
}
