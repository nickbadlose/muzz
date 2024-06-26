package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nickbadlose/muzz/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

func TestSuccess(t *testing.T) {
	cases := []struct {
		endpoint, method string
		expectedCode     int
	}{
		{"status", http.MethodGet, http.StatusOK},
	}

	srv := httptest.NewServer(router.New())
	defer srv.Close()
	for _, tc := range cases {
		t.Run(tc.endpoint, func(t *testing.T) {

			resp := makeRequest(t, tc.method, fmt.Sprintf("%s/%s", srv.URL, tc.endpoint), nil)
			require.Equal(t, tc.expectedCode, resp.StatusCode)

			testDir := getTestDataDirectory()
			expected, err := os.ReadFile(filepath.Join(
				testDir,
				strings.ReplaceAll(fmt.Sprintf("%s.%d.json", tc.endpoint, tc.expectedCode), "/", "."),
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
