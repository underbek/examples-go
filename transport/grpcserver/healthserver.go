package grpcserver

import (
	"context"

	health_api "google.golang.org/grpc/health/grpc_health_v1"
)

type healthserver struct {
	srv    health_api.HealthServer
	checks []checkHealtsFunc
}

func newHealthserver(srv health_api.HealthServer, checks ...checkHealtsFunc) *healthserver {
	return &healthserver{srv: srv, checks: checks}
}

type checkHealtsFunc func(ctx context.Context) bool

func (h *healthserver) Check(ctx context.Context, request *health_api.HealthCheckRequest) (*health_api.HealthCheckResponse, error) {
	for _, f := range h.checks {
		if !f(ctx) {
			return &health_api.HealthCheckResponse{Status: health_api.HealthCheckResponse_NOT_SERVING}, nil
		}
	}

	return &health_api.HealthCheckResponse{Status: health_api.HealthCheckResponse_SERVING}, nil
}

func (h *healthserver) Watch(request *health_api.HealthCheckRequest, server health_api.Health_WatchServer) error {
	return h.srv.Watch(request, server)
}
