package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/api/router"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/auth"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/database/adapter/postgres"
	"github.com/nickbadlose/muzz/internal/service"
	"github.com/nickbadlose/muzz/internal/tracer"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

// prevents race conditions from the viper package when running tests in parallel.
var cfg *config.Config

func init() {
	cfg = config.MustLoad()
}

// mockLocation to mock getting the location from IP address. This is the only part of our integration tests we mock,
// so we don't spam the geoip service.
//
// It also allows us to use a static location for the logged-in user, so test results are static.
type mockLocation struct{}

func (*mockLocation) ByIP(_ context.Context, _ string) (orb.Point, error) {
	return orb.Point{-5.0527, 50.266}, nil
}

// newTestServer configures and runs a test server to make requests against.
//
// all services are configured as if running in development environment, except for the location service, which is
// mocked.
func newTestServer(t *testing.T, dbName string) *httptest.Server {
	setupDB(t, cfg, dbName)
	seedTestData(t, cfg, dbName)

	db, err := database.New(context.Background(), &database.Credentials{
		Username: cfg.DatabaseUser(),
		Password: cfg.DatabasePassword(),
		Name:     dbName,
		Host:     cfg.DatabaseHost(),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	matchAdapter, err := postgres.NewSwipeAdapter(db)
	require.NoError(t, err)
	userAdapter, err := postgres.NewUserAdapter(db)
	require.NoError(t, err)

	authorizer, err := auth.NewAuthoriser(cfg, userAdapter)
	require.NoError(t, err)

	authService, err := service.NewAuthService(authorizer, userAdapter)
	require.NoError(t, err)
	matchService, err := service.NewSwipeService(matchAdapter)
	require.NoError(t, err)
	userService, err := service.NewUserService(userAdapter)
	require.NoError(t, err)

	hlr, err := handlers.New(cfg, authorizer, &mockLocation{}, authService, userService, matchService)
	require.NoError(t, err)

	tp, err := tracer.New(cfg, "muzz")
	require.NoError(t, err)

	route, err := router.New(cfg, hlr, authorizer, tp)
	require.NoError(t, err)
	srv := httptest.NewServer(route)
	t.Cleanup(srv.Close)

	return srv
}

// seedTestData seeds the given test database with test data.
func seedTestData(t *testing.T, cfg *config.Config, dbName string) {
	seedMigrator, err := migrate.New(
		"file://./migrations/seed",
		fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable&x-migrations-table=schema_seed_migrations",
			cfg.DatabaseUser(),
			cfg.DatabasePassword(),
			cfg.DatabaseHost(),
			dbName,
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = seedMigrator.Down()
		require.NoError(t, err)
		sErr, dErr := seedMigrator.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	err = seedMigrator.Up()
	require.NoError(t, err)
}

// setupDB handles the creation and destruction of the provided test database to run integration tests against.
func setupDB(t *testing.T, cfg *config.Config, dbName string) {
	createTestDB(t, cfg, dbName)

	// app migrator to run all application migrations.
	appMigrator, err := migrate.New(
		"file://../migrations",
		fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			cfg.DatabaseUser(),
			cfg.DatabasePassword(),
			cfg.DatabaseHost(),
			dbName,
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
	require.NoError(t, err)
}

// createTestDB is idempotent to handle the cases where previous tests are ungracefully shut down and leave the
// migrations schema in a 'dirty' state, which would cause errors when running migrator.Up().
func createTestDB(t *testing.T, cfg *config.Config, dbName string) {
	migrationPath := createTempMigrations(t, dbName)

	// since we use the default DB to create our test DB, we need unique
	// migration table names per test to avoid dirty database errors.
	migrationTable := fmt.Sprintf("schema_migrations_%s", dbName)

	// create test DB migrator to set up and teardown test db.
	// You can't drop a database whilst connections still exist, so we authenticate to the postgres DB to run these.
	createTestDBMigrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		fmt.Sprintf(
			"postgres://%s:%s@%s/postgres?sslmode=disable&x-migrations-table=%s",
			cfg.DatabaseUser(),
			cfg.DatabasePassword(),
			cfg.DatabaseHost(),
			migrationTable,
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		cErr := createTestDBMigrator.Down()
		require.NoError(t, cErr)
		require.NoError(t, cErr)

		body := io.NopCloser(strings.NewReader(fmt.Sprintf(`DROP TABLE IF EXISTS "%s";`, migrationTable)))
		cleanupMigration, cErr := migrate.NewMigration(body, "drop_schema_table", 1, -1)
		require.NoError(t, cErr)
		// intentionally ignore this error, as we delete the migration table from the default DB
		// for cleanup of tests and the migrator errors when trying to write the new version to it.
		_ = createTestDBMigrator.Run(cleanupMigration)

		sErr, dErr := createTestDBMigrator.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	// force idempotency by clearing any old test data and reset test db migration version handling any ungraceful
	// shutdowns in previous tests.
	body := io.NopCloser(strings.NewReader(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s";`, dbName)))
	idempotencyMigration, err := migrate.NewMigration(body, "drop_test_database", 1, -1)
	require.NoError(t, err)
	err = createTestDBMigrator.Run(idempotencyMigration)
	require.NoError(t, err)

	// create our test database.
	err = createTestDBMigrator.Up()
	require.NoError(t, err)
}

func getTestDataDirectory() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "data")
}

type header struct {
	key   string
	value string
}

// makeRequest builds and makes a http request and returns the response for testing against.
func makeRequest(t *testing.T, method, url string, data interface{}, headers ...*header) *http.Response {
	var body []byte

	if data != nil {
		var err error
		body, err = json.Marshal(data)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// createTempMigrations dynamically creates our temp migrations for the given test. This function is idempotent. on
// cleanup, it removes any created resources.
//
// If using running in parallel, there are some conditions:
//   - dbName MUST be unique across ALL tests, otherwise migrations and test data will clash.
//   - The database the user connects to, to run these setup test database migrations,
//     must not be dropped, as it could affect other tests being run.
func createTempMigrations(t *testing.T, dbName string) string {
	tempDirPath := os.TempDir() + dbName

	// clear any old test files if execution was ungracefully shut down previously without cleanup.
	err := os.RemoveAll(tempDirPath)
	require.NoError(t, err)

	// create temp migration directory.
	err = os.Mkdir(tempDirPath, 0777)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Remove(tempDirPath))
	})

	// create dynamic migrations for creating and destroying the test database.
	files := []struct{ name, content string }{
		{
			name:    "1_create_test_db.up.sql",
			content: fmt.Sprintf(`CREATE DATABASE "%s";`, dbName),
		},
		{
			name:    "1_create_test_db.down.sql",
			content: fmt.Sprintf(`DROP DATABASE IF EXISTS "%s";`, dbName),
		},
	}

	for _, f := range files {
		data := []byte(f.content)
		migrationFile, fErr := os.CreateTemp(tempDirPath, f.name)
		require.NoError(t, fErr)
		t.Cleanup(func() {
			require.NoError(t, os.Remove(migrationFile.Name()))
		})

		_, fErr = migrationFile.Write(data)
		require.NoError(t, fErr)
	}

	return tempDirPath
}
