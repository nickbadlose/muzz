package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nickbadlose/muzz/internal/store"
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
	"github.com/nickbadlose/muzz/internal/app"
	"github.com/nickbadlose/muzz/internal/pkg/database"
	"github.com/nickbadlose/muzz/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) database.Database {
	// test migrator to set up and teardown test db.
	testM, err := migrate.New(
		"file://./migrations",
		"postgres://nick:password@localhost:5432/postgres?sslmode=disable",
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = testM.Down()
		if !errors.Is(err, migrate.ErrNoChange) {
			require.NoError(t, err)
		}
		sErr, dErr := testM.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	err = testM.Up()
	if !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}

	// TODO run appM.Drop at the start of each migration to clear all data if db already existed
	// app migrator to run application migrations.
	appM, err := migrate.New(
		"file://../migrations",
		"postgres://nick:password@localhost:5432/test?sslmode=disable",
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = appM.Down()
		require.NoError(t, err)
		sErr, dErr := appM.Close()
		require.NoError(t, sErr)
		require.NoError(t, dErr)
	})

	err = appM.Up()
	if !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err)
	}

	db, err := database.New(context.Background(), &database.Config{
		Username: "nick",
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

func TestSuccess(t *testing.T) {
	cases := []struct {
		endpoint, method string
		body             interface{}
		res              interface{}
		expectedCode     int
	}{
		{endpoint: "status", method: http.MethodGet, expectedCode: http.StatusOK},
		{
			endpoint: "user/create",
			method:   http.MethodPost,
			body: &app.CreateUserRequest{
				Email:    "test@test.com",
				Password: "Pa55w0rd!",
				Name:     "test",
				Gender:   app.GenderFemale,
				Age:      25,
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {
			db := setupDB(t)
			storage := store.New(db)
			service := app.NewService(storage)
			handlers := app.NewHandlers(service)

			srv := httptest.NewServer(router.New(handlers))
			t.Cleanup(srv.Close)

			resp := makeRequest(t, tc.method, fmt.Sprintf("%s/%s", srv.URL, tc.endpoint), tc.body)
			require.Equal(t, tc.expectedCode, resp.StatusCode)

			testDir := getTestDataDirectory()
			expected, err := os.ReadFile(filepath.Join(
				testDir,
				strings.ReplaceAll(fmt.Sprintf("%s.%s.json", tc.endpoint, tc.method), "/", "."),
			))
			require.NoError(t, err)

			got, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			assert.JSONEq(t, string(expected), string(got))
		})
	}
}

func getTestDataDirectory() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "data")
}

func makeRequest(t *testing.T, method string, path string, data interface{}) *http.Response {
	var body []byte

	if data != nil {
		var err error
		body, err = json.Marshal(data)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}
