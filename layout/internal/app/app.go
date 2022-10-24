package app

import (
	"fmt"
	"net/http"

	"layout/internal/config"
	"layout/internal/handler"
	"layout/internal/router"
	"layout/internal/storage"
	"layout/internal/usecase"
	"layout/pkg/logger"

	"go.uber.org/zap"
)

type App struct {
	HTTPServer *http.Server
	logger     *zap.Logger
}

func New(cfg config.Config) (*App, error) {
	lg, err := logger.New(true)
	if err != nil {
		return nil, err
	}

	repo := storage.New(lg)
	useCase := usecase.New(lg, repo)
	h := handler.New(lg, useCase)

	srv := &http.Server{
		Handler: router.New(h),
		Addr:    fmt.Sprintf(":%d", cfg.AppPort),
	}

	return &App{
		HTTPServer: srv,
		logger:     lg,
	}, nil
}

func (app *App) Run() error {
	app.logger.Info("server started", zap.String("addr", app.HTTPServer.Addr))
	return app.HTTPServer.ListenAndServe()
}
