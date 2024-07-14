package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nickbadlose/muzz/api/handlers"
	"github.com/stretchr/testify/require"
)

func TestPublicRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

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

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%s/%s %s", tc.method, tc.endpoint, tc.description), func(t *testing.T) {
			t.Parallel()

			testDBName := fmt.Sprintf(
				"test_%s_%d_public",
				strings.ReplaceAll(tc.endpoint, "/", ""),
				i,
			)

			srv := newTestServer(t, testDBName)

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

func TestPrivateRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

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

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%s/%s %s", tc.method, tc.endpoint, tc.description), func(t *testing.T) {
			t.Parallel()

			testDBName := fmt.Sprintf(
				"test_%s_%d_private",
				strings.ReplaceAll(tc.endpoint, "/", ""),
				i,
			)

			srv := newTestServer(t, testDBName)

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

func TestPrivateRoutes_Unauthorised(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

	cases := []struct {
		endpoint, method string
	}{
		{
			endpoint: "discover",
			method:   http.MethodGet,
		},
		{
			endpoint: "swipe",
			method:   http.MethodPost,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("POST/%s unauthorised", tc.endpoint), func(t *testing.T) {
			t.Parallel()

			testDBName := fmt.Sprintf(
				"test_%s_%d_unauthorised",
				strings.ReplaceAll(tc.endpoint, "/", ""),
				i,
			)

			srv := newTestServer(t, testDBName)

			resp := makeRequest(
				t,
				tc.method,
				fmt.Sprintf("%s/%s", srv.URL, tc.endpoint),
				nil,
				&header{
					key: "Authorization",
					// expired token
					value: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjEsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6MzAwMCIsImF1ZCI6WyJodHRwOi8vbG9jYWxob3N0OjMwMDAiXSwiZXhwIjoxNzIwODgzODc3LCJuYmYiOjE3MjA4NDA2NzcsImlhdCI6MTcyMDg0MDY3N30.P9cknYtIi5WyfeDH6cYH-9Jdtxjsg_FB-WoNNacSSrs",
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
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

	// password gets encrypted.
	t.Run("POST/user/create success", func(t *testing.T) {
		srv := newTestServer(t, "test_usercreate_success_inconsistent")

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
