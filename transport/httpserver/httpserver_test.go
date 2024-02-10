package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/hellofresh/health-go/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
)

func TestHttpServer(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	healthCheck, err := health.New(health.WithComponent(health.Component{
		Name:    "go-kit",
		Version: os.Getenv("CI_COMMIT_SHORT_SHA"),
	}))
	require.NoError(t, err)

	srv := New(
		l,
		Config{
			Port:         8080,
			WriteTimeout: time.Second,
			ReadTimeout:  time.Second,
		},
		healthCheck.Handler())

	testServer := httptest.NewServer(srv.Server.Handler)
	defer testServer.Close()

	resp, err := testServer.Client().Get(testServer.URL + "/status")
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, err)

	var payload health.Check
	err = json.NewDecoder(resp.Body).Decode(&payload)
	require.NoError(t, err)
	require.Equal(t, "go-kit", payload.Name)
}

func TestBodies(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	requestBody := []byte("request body")
	responseBody := []byte("response body")

	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "status code 200",
			statusCode: http.StatusOK,
		},
		{
			name:       "status code 500",
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := New(
				l,
				Config{
					Port:         8080,
					WriteTimeout: time.Second,
					ReadTimeout:  time.Second,
				},
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					actualRequestBody, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.Equal(t, requestBody, actualRequestBody)
					w.WriteHeader(tt.statusCode)

					_, err = w.Write(responseBody)
					assert.NoError(t, err)
				}))

			testServer := httptest.NewServer(srv.Server.Handler)
			defer testServer.Close()

			resp, err := testServer.Client().Post(testServer.URL, "text", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			require.Equal(t, tt.statusCode, resp.StatusCode)

			actualResponseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, responseBody, actualResponseBody)
		})
	}
}
