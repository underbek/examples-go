package grpcserver

import (
	"context"

	healthApi "google.golang.org/grpc/health/grpc_health_v1"
)

type healthServer struct {
	srv    healthApi.HealthServer
	checks []checkHealthFunc
}

func newHealthServer(srv healthApi.HealthServer, checks ...checkHealthFunc) *healthServer {
	return &healthServer{srv: srv, checks: checks}
}

type checkHealthFunc func(ctx context.Context) bool

func (h *healthServer) Check(ctx context.Context, request *healthApi.HealthCheckRequest) (*healthApi.HealthCheckResponse, error) {
	for _, f := range h.checks {
		if !f(ctx) {
			return &healthApi.HealthCheckResponse{Status: healthApi.HealthCheckResponse_NOT_SERVING}, nil
		}
	}

	return &healthApi.HealthCheckResponse{Status: healthApi.HealthCheckResponse_SERVING}, nil
}

func (h *healthServer) Watch(request *healthApi.HealthCheckRequest, server healthApi.Health_WatchServer) error {
	return h.srv.Watch(request, server)
}
