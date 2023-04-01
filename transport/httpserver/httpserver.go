package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/underbek/examples-go/config"
	"github.com/underbek/examples-go/logger"
	"golang.org/x/sync/errgroup"

	mw "github.com/underbek/examples-go/middlewares/http"
)

type HTTPServer struct {
	logger     *logger.Logger
	Server     *http.Server
	serverPort int
}

func New(logger *logger.Logger, cfgServer config.HTTPServer, handler http.Handler) *HTTPServer {
	server := &http.Server{
		Handler:      mw.JaegerTraceMiddleware(mw.Logging(logger, cfgServer.ShowHealthLogs)(handler)),
		Addr:         fmt.Sprintf(":%d", cfgServer.Port),
		WriteTimeout: cfgServer.WriteTimeout,
		ReadTimeout:  cfgServer.ReadTimeout,
	}

	return &HTTPServer{
		logger:     logger,
		Server:     server,
		serverPort: cfgServer.Port,
	}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	s.logger.With("port", s.serverPort).Info("http server listening")

	defer func() {
		s.logger.Info("http server stopped")
	}()

	group, ctx := errgroup.WithContext(ctx)

	// start http
	group.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := s.Server.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("router shutdown error")
			return err
		}

		return nil
	})

	group.Go(func() error {
		if err := s.Server.ListenAndServe(); err != nil {
			s.logger.WithError(err).Error("http server")
			return err
		}
		return nil
	})

	return group.Wait()
}
