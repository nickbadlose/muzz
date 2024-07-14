package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

func TestConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

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

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testDBName := fmt.Sprintf("test_%d_constraints", i)

			setupDB(t, cfg, testDBName)

			// trigger constraint errors
			constraintMigrator, err := migrate.New(
				fmt.Sprintf(
					"file://./migrations/constraints/%s",
					strings.ReplaceAll(tc.name, " ", "_"),
				),
				fmt.Sprintf(
					"postgres://%s:%s@%s/%s?sslmode=disable&x-migrations-table=schema_constraint_migrations",
					cfg.DatabaseUser(),
					cfg.DatabasePassword(),
					cfg.DatabaseHost(),
					testDBName,
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
