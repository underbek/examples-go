package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/underbek/examples-go/logger"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	logger *logger.Logger
	server *http.Server
	port   int
}

func New(logger *logger.Logger, port int) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 15,
	}

	return &Server{
		logger: logger,
		server: server,
		port:   port,
	}
}

func (ms *Server) Run(ctx context.Context) error {
	ms.logger.With("port", ms.port).Info("metric server listening")

	defer func() {
		ms.logger.Info("metric server stopped")
	}()

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ms.logger.WithError(err).Error("metric server")
			return err
		}
		return nil
	})

	group.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return ms.server.Shutdown(shutdownCtx)
	})

	return group.Wait()
}
