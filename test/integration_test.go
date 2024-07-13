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
//  - search fmt.Println and log and clear them if not for production
//  - check each handler for potential errors, query errors etc.

// TODO for temp file - check if Create truncates if exists? I think it does. CreateTempDir?
//// generate a temp file.
//	f, err := os.CreateTemp("", "logger-test-*")
//	if err != nil {
//		panic(err)
//	}
//
//	// defer that we remove the temp file and close the logger.
//	t.Cleanup(func() {
//		require.NoError(t, Close())
//		require.NoError(t, os.Remove(f.Name()))
//	})

// TODO use actual geoip data for integration tests, since we cache data, it shouldn't be an issue spamming the service.
//  It would be nice for it to not be dependant on an external service though.

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

	db := setupDB(t, cfg)
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

func setupDB(t *testing.T, cfg *config.Config) *database.Database {
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
		require.NoError(t, err)
		sErr, dErr := createTestDBMigrator.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	err = createTestDBMigrator.Up()
	require.NoError(t, err)

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
	require.NoError(t, err)

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
		err = seedMigrator.Down()
		require.NoError(t, err)
		sErr, dErr := seedMigrator.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	err = seedMigrator.Up()
	require.NoError(t, err)

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
		endpoint, method, description string
		body                          interface{}
		expectedCode                  int
	}{
		{
			endpoint:     "status",
			method:       http.MethodGet,
			description:  "success",
			expectedCode: http.StatusOK,
		},
		{
			endpoint:     "user/create",
			method:       http.MethodPost,
			description:  "bad request",
			body:         "bad request",
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:     "user/create",
			method:       http.MethodPost,
			description:  "invalid input",
			body:         &handlers.CreateUserRequest{Email: "invalidemail"},
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:     "login",
			method:       http.MethodPost,
			description:  "bad request",
			body:         "bad request",
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:     "login",
			method:       http.MethodPost,
			description:  "invalid input",
			body:         &handlers.LoginRequest{Email: "test@test.com"},
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:     "login",
			method:       http.MethodPost,
			description:  "incorrect credentials",
			body:         &handlers.LoginRequest{Email: "test@test.com", Password: "wrongPassword"},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s/%s %s", tc.method, tc.endpoint, tc.description), func(t *testing.T) {
			srv := newTestServer(t)

			resp := makeRequest(t, tc.method, fmt.Sprintf("%s/%s", srv.URL, tc.endpoint), tc.body)

			testDir := getTestDataDirectory()
			expected, err := os.ReadFile(filepath.Join(
				testDir,
				strings.ReplaceAll(
					fmt.Sprintf(
						"%s.%s.%d.%s.json",
						tc.endpoint,
						tc.method,
						tc.expectedCode,
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

func TestAuthenticated(t *testing.T) {
	cases := []struct {
		endpoint, method, description, queryParams, token string
		body                                              interface{}
		expectedCode                                      int
		headers                                           []*header
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
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "no results",
			queryParams:  "maxAge=70&minAge=70&genders=female",
			expectedCode: http.StatusNotFound,
		},
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "bad params",
			queryParams:  "maxAge=notAnInt",
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "invalidparams",
			queryParams:  "minAge=25&maxAge=20",
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:     "discover",
			method:       http.MethodGet,
			description:  "no authorisation",
			expectedCode: http.StatusUnauthorized,
			headers: []*header{
				{
					key:   "Authorization",
					value: fmt.Sprintf("Bearer %s", ""),
				},
			},
		},
		{
			endpoint:    "swipe",
			description: "match",
			method:      http.MethodPost,
			body: &handlers.SwipeRequest{
				UserID:     2,
				Preference: true,
			},
			expectedCode: http.StatusCreated,
		},
		{
			endpoint:    "swipe",
			description: "no match",
			method:      http.MethodPost,
			body: &handlers.SwipeRequest{
				UserID:     3,
				Preference: true,
			},
			expectedCode: http.StatusCreated,
		},
		{
			endpoint:     "swipe",
			method:       http.MethodPost,
			description:  "bad request",
			body:         "bad request",
			expectedCode: http.StatusBadRequest,
		},
		{
			endpoint:    "swipe",
			method:      http.MethodPost,
			description: "invalid input",
			body: &handlers.SwipeRequest{
				UserID:     1,
				Preference: true,
			},
			expectedCode: http.StatusBadRequest,
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

			headers := []*header{
				{
					key:   "Authorization",
					value: fmt.Sprintf("Bearer %s", loginRes.Token),
				},
			}
			if tc.headers != nil {
				headers = append(headers, tc.headers...)
			}

			resp := makeRequest(
				t,
				tc.method,
				fmt.Sprintf("%s/%s?%s", srv.URL, tc.endpoint, tc.queryParams),
				tc.body,
				headers...,
			)

			testDir := getTestDataDirectory()
			expected, err := os.ReadFile(filepath.Join(
				testDir,
				strings.ReplaceAll(
					fmt.Sprintf(
						"%s.%s.%d.%s.json",
						tc.endpoint,
						tc.method,
						tc.expectedCode,
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

func TestAuthenticated_Unauthorised(t *testing.T) {
	cases := []struct {
		endpoint, method, token string
	}{
		{
			endpoint: "discover",
			method:   http.MethodGet,
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjEsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6MzAwMCIsImF1ZCI6WyJodHRwOi8vbG9jYWxob3N0OjMwMDAiXSwiZXhwIjoxNzIwODgzODc3LCJuYmYiOjE3MjA4NDA2NzcsImlhdCI6MTcyMDg0MDY3N30.P9cknYtIi5WyfeDH6cYH-9Jdtxjsg_FB-WoNNacSSrs",
		},
		{
			endpoint: "swipe",
			method:   http.MethodPost,
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjEsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6MzAwMCIsImF1ZCI6WyJodHRwOi8vbG9jYWxob3N0OjMwMDAiXSwiZXhwIjoxNzIwODgzODc3LCJuYmYiOjE3MjA4NDA2NzcsImlhdCI6MTcyMDg0MDY3N30.P9cknYtIi5WyfeDH6cYH-9Jdtxjsg_FB-WoNNacSSrs",
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("POST/%s unauthorised", tc.endpoint), func(t *testing.T) {
			srv := newTestServer(t)

			resp := makeRequest(
				t,
				tc.method,
				fmt.Sprintf("%s/%s", srv.URL, tc.endpoint),
				nil,
				&header{
					key:   "Authorization",
					value: fmt.Sprintf("Bearer %s", tc.token),
				},
			)

			got, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			require.JSONEq(
				t,
				`{"status": 401,"error": "token has invalid claims: token is expired"}`,
				string(got),
			)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
}

// Tests for any endpoints where the response isn't consistent so can't be asserted like the other requests.
func TestInconsistentResponse(t *testing.T) {
	// password gets encrypted.
	t.Run("POST/user/create success", func(t *testing.T) {
		srv := newTestServer(t)

		resp := makeRequest(
			t,
			http.MethodPost,
			fmt.Sprintf("%s/%s", srv.URL, "user/create"),
			&handlers.CreateUserRequest{
				Email:    "test8@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   "female",
				Age:      25,
				Location: handlers.Location{Lat: 50.266, Lon: -5.0527},
			},
		)

		got := &handlers.UserResponse{}
		err := json.NewDecoder(resp.Body).Decode(got)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// since password in encrypted, we cannot assert it.
		require.NotEmpty(t, got.Result.Password)
		got.Result.Password = ""
		require.Equal(
			t,
			&handlers.UserResponse{
				Result: &handlers.User{
					ID:       8,
					Email:    "test8@test.com",
					Password: "",
					Name:     "test",
					Gender:   "female",
					Age:      25,
					Location: handlers.Location{Lat: 50.266, Lon: -5.0527},
				},
			},
			got,
		)
	})
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
