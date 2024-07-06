package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/nickbadlose/muzz/api/router"
	"github.com/nickbadlose/muzz/config"
	"github.com/nickbadlose/muzz/internal/auth"
	"github.com/nickbadlose/muzz/internal/database"
	"github.com/nickbadlose/muzz/internal/database/adapter/postgres"
	"github.com/nickbadlose/muzz/internal/service"
	"github.com/stretchr/testify/require"
)

// TODO
//  - have migrations_test.go file in here which test constraints etc.

func newTestServer(t *testing.T) *httptest.Server {
	db := setupDB(t)
	cfg := config.Load()
	matchAdapter := postgres.NewMatchAdapter(db)
	userAdapter := postgres.NewUserAdapter(db)

	authorizer := auth.NewAuthorizer(cfg)

	authService := service.NewAuthService(userAdapter, authorizer)
	matchService := service.NewMatchService(matchAdapter)
	userService := service.NewUserService(userAdapter)

	hlr := handlers.New(authorizer, authService, userService, matchService)

	srv := httptest.NewServer(router.New(hlr, authorizer))
	t.Cleanup(srv.Close)

	return srv
}

func setupDB(t *testing.T) *database.Database {
	// TODO don't continue on errorNoChange, it means some migrations haven't run correctly?

	// TODO auth from cfg
	// create test DB migrator to set up and teardown test db.
	// You can't drop a database whilst connections still exist, so we authenticate to the postgres DB to run these.
	createTestDBMigrator, err := migrate.New(
		"file://./migrations/create",
		"postgres://nickbadlose:password@localhost:5432/postgres?sslmode=disable",
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

	// TODO get auth from env
	// TODO run appMigrator.Drop at the start of each migration to clear all data if db already existed
	// app migrator to run all application migrations.
	appMigrator, err := migrate.New(
		"file://../migrations",
		"postgres://nickbadlose:password@localhost:5432/test?sslmode=disable",
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
	seedMigrator, err := migrate.New(
		"file://./migrations/seed",
		"postgres://nickbadlose:password@localhost:5432/test?sslmode=disable&x-migrations-table=schema_seed_migrations",
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		fmt.Println("seedMigrator.Down()")
		err = seedMigrator.Down()
		require.NoError(t, err)
		sErr, dErr := seedMigrator.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	err = seedMigrator.Up()
	if !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}

	// TODO auth from cfg
	db, err := database.New(context.Background(), &database.Config{
		Username: "nickbadlose",
		Password: "password",
		Name:     "test",
		Host:     "localhost:5432",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

// TODO don't test login here as it will be tested by logging in to call authenticated routes

func TestPublic(t *testing.T) {
	cases := []struct {
		endpoint, method string
		body             interface{}
		expectedCode     int
	}{
		{endpoint: "status", method: http.MethodGet, expectedCode: http.StatusOK},
		{
			endpoint: "user/create",
			method:   http.MethodPost,
			body: &handlers.CreateUserRequest{
				Email:    "test6@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "female",
				Age:      25,
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s/%s", tc.method, tc.endpoint), func(t *testing.T) {
			srv := newTestServer(t)

			resp := makeRequest(t, tc.method, fmt.Sprintf("%s/%s", srv.URL, tc.endpoint), tc.body)

			testDir := getTestDataDirectory()
			expected, err := os.ReadFile(filepath.Join(
				testDir,
				strings.ReplaceAll(fmt.Sprintf("%s.%s.json", tc.endpoint, tc.method), "/", "."),
			))
			require.NoError(t, err)

			got, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			require.JSONEq(t, string(expected), string(got))
			require.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

func TestAuthenticated(t *testing.T) {
	cases := []struct {
		endpoint, method, description string
		body                          interface{}
		expectedCode                  int
	}{
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "all users",
			expectedCode: http.StatusOK,
		},
		{
			endpoint:    "swipe",
			description: "match",
			method:      http.MethodPost,
			body: &handlers.SwipeRequest{
				UserID:     2,
				Preference: true,
			},
			expectedCode: http.StatusOK,
		},
		{
			endpoint:    "swipe",
			description: "no match",
			method:      http.MethodPost,
			body: &handlers.SwipeRequest{
				UserID:     3,
				Preference: true,
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s/%s %s", tc.method, tc.endpoint, tc.description), func(t *testing.T) {
			srv := newTestServer(t)

			loginData := makeRequest(
				t,
				http.MethodPost,
				fmt.Sprintf("%s/%s", srv.URL, "login"),
				&handlers.LoginRequest{
					Email:    "test@test.com",
					Password: "Pa55w0rd!",
				},
			)

			loginRes := &handlers.LoginResponse{}
			err := json.NewDecoder(loginData.Body).Decode(loginRes)
			require.NoError(t, err)
			require.NotEmpty(t, loginRes.Token)

			resp := makeRequest(
				t,
				tc.method,
				fmt.Sprintf("%s/%s", srv.URL, tc.endpoint),
				tc.body,
				&header{
					key:   "Authorization",
					value: fmt.Sprintf("Bearer %s", loginRes.Token),
				},
			)

			testDir := getTestDataDirectory()
			expected, err := os.ReadFile(filepath.Join(
				testDir,
				strings.ReplaceAll(
					fmt.Sprintf(
						"%s.%s.%s.json",
						tc.endpoint,
						tc.method,
						strings.ReplaceAll(tc.description, " ", ""),
					),
					"/",
					".",
				),
			))
			require.NoError(t, err)

			got, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			require.JSONEq(t, string(expected), string(got))
			require.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}

func getTestDataDirectory() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "data")
}

type header struct {
	key   string
	value string
}

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
