package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	health "github.com/hellofresh/health-go/v5"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/config"
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
		config.HTTPServer{
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
