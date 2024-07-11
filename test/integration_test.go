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
	"github.com/nickbadlose/muzz/internal/tracer"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

// TODO
//  - have migrations_test.go file in here which test constraints etc.
//  - try t.parallel to speed these up, I think it might fail if all accessing same db, is there a way to fix that?
//  - try test main func

// mockLocation to mock getting the location from IP address. This is the only part of our integration tests we mock,
// so we don't spam the geoip service.
//
// It also allows us to use a static location for the logged-in user, so test results are static.
type mockLocation struct{}

func (*mockLocation) ByIP(_ context.Context, _ string) (orb.Point, error) {
	return orb.Point{-5.0527, 50.266}, nil
}

func newTestServer(t *testing.T) *httptest.Server {
	cfg, err := config.Load()
	require.NoError(t, err)

	fmt.Println(cfg.DatabaseUser(), cfg.DatabasePassword())

	db := setupDB(t, cfg)
	matchAdapter := postgres.NewMatchAdapter(db)
	userAdapter := postgres.NewUserAdapter(db)

	authorizer := auth.NewAuthorizer(cfg)

	authService := service.NewAuthService(userAdapter, authorizer)
	matchService := service.NewMatchService(matchAdapter)
	userService := service.NewUserService(userAdapter)

	hlr := handlers.New(cfg, authorizer, &mockLocation{}, authService, userService, matchService)

	tp, err := tracer.New(context.Background(), cfg, "muzz")
	require.NoError(t, err)

	srv := httptest.NewServer(router.New(hlr, cfg, authorizer, tp))
	t.Cleanup(srv.Close)

	return srv
}

func setupDB(t *testing.T, cfg *config.Config) *database.Database {
	// TODO don't continue on errorNoChange, it means some migrations haven't run correctly?

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
	seedMigrator, err := migrate.New(
		"file://./migrations/seed",
		fmt.Sprintf(
			"postgres://%s:%s@%s/test?sslmode=disable&x-migrations-table=schema_seed_migrations",
			cfg.DatabaseUser(),
			cfg.DatabasePassword(),
			cfg.DatabaseHost(),
		),
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

	db, err := database.New(context.Background(), &database.Credentials{
		Username: cfg.DatabaseUser(),
		Password: cfg.DatabasePassword(),
		Name:     "test",
		Host:     cfg.DatabaseHost(),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

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
				Email:    "test8@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "female",
				Age:      25,
				Location: handlers.Location{Lat: 50.266, Lon: -5.0527},
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
		endpoint, method, description, queryParams string
		body                                       interface{}
		expectedCode                               int
	}{
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "all users",
			expectedCode: http.StatusOK,
		},
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "females 20 - 30",
			queryParams:  "maxAge=30&minAge=20&genders=female",
			expectedCode: http.StatusOK,
		},
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "males and unspecified",
			queryParams:  "genders=male,unspecified",
			expectedCode: http.StatusOK,
		},
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "sort type attractiveness",
			queryParams:  "sort=attractiveness",
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
				fmt.Sprintf("%s/%s?%s", srv.URL, tc.endpoint, tc.queryParams),
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
