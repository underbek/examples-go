package grpcserver

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthApi "google.golang.org/grpc/health/grpc_health_v1"
)

func TestHealths(t *testing.T) {
	log, err := logger.New(true)
	require.NoError(t, err)

	mx := &sync.Mutex{}
	data := true

	srv := New(
		log,
		Config{
			Port: 8081,
		},
		func(ctx context.Context) bool {
			mx.Lock()
			defer mx.Unlock()
			return data
		})

	ctx := context.Background()

	go func() {
		err = srv.Run(ctx)
		require.NoError(t, err)
	}()

	con, err := grpc.DialContext(ctx, ":8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer con.Close()

	hsrv := healthApi.NewHealthClient(con)

	var res *healthApi.HealthCheckResponse
	require.Eventually(t, func() bool {
		res, err = hsrv.Check(ctx, &healthApi.HealthCheckRequest{})
		return err == nil
	}, time.Second*30, time.Millisecond*100)

	require.NoError(t, err)
	require.True(t, res.Status == healthApi.HealthCheckResponse_SERVING)

	mx.Lock()
	data = false
	mx.Unlock()

	res, err = hsrv.Check(ctx, &healthApi.HealthCheckRequest{})
	require.NoError(t, err)
	require.True(t, res.Status == healthApi.HealthCheckResponse_NOT_SERVING)
}
