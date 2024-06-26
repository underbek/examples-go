package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/underbek/examples-go/logger"
	mw "github.com/underbek/examples-go/middlewares/grpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthApi "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	logger     *logger.Logger
	Server     *grpc.Server
	serverPort int
}

func New(logger *logger.Logger, cfgServer Config, checks ...checkHealthFunc) *GRPCServer {

	gRPCServer := grpc.NewServer(
		mw.UnaryInterceptors(logger, cfgServer.ShowHealthLogs, cfgServer.ShowPayloadLogs, cfgServer.Timeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: cfgServer.KeepAlive,
		}),
	)

	reflection.Register(gRPCServer)

	baseHealthServer := health.NewServer()
	healthServer := newHealthserver(baseHealthServer, checks...)
	healthApi.RegisterHealthServer(gRPCServer, healthServer)

	return &GRPCServer{
		logger:     logger,
		Server:     gRPCServer,
		serverPort: cfgServer.Port,
	}
}

func (s *GRPCServer) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.serverPort))
	if err != nil {
		s.logger.
			With("port", s.serverPort).
			WithError(err).
			Error("tcp listen failed")
		return err
	}

	s.logger.With("port", s.serverPort).Info("grpc server listening")

	defer func() {
		s.logger.Info("grpc server stopped")
	}()

	group, ctx := errgroup.WithContext(ctx)

	// start grpc
	group.Go(func() error {
		<-ctx.Done()
		s.Server.GracefulStop()
		return nil
	})

	group.Go(func() error {
		if err := s.Server.Serve(listener); err != nil {
			s.logger.WithError(err).Error("grpc serve failed")
			return err
		}
		return nil
	})

	return group.Wait()
}
