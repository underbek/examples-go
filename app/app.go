package app

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/underbek/examples-go/config"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/metrics"
	"github.com/underbek/examples-go/tracing"
	"golang.org/x/sync/errgroup"
)

type Runner = func(ctx context.Context) error
type Defer = func()

type App struct {
	Logger        *logger.Logger
	MetricsServer *metrics.Server
	runners       []Runner
	defers        []Defer
}

func NewApp(cfgApp config.App, cfgJaeger tracing.Config) (*App, error) {
	log, err := logger.New(cfgApp.Debug)
	if err != nil {
		return nil, err
	}

	if cfgApp.Debug {
		log.With("config", cfgApp).Debug("debug mode turned on")
	}

	provider, err := tracing.NewProvider(cfgJaeger)
	if err != nil {
		return nil, err
	}

	app := &App{
		Logger:        log,
		MetricsServer: metrics.New(log, cfgApp.MetricsPort),
	}

	app.AddDefers(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			app.Logger.WithError(err).Error("jaeger provider shutdown error")
		}
	})

	return app, nil
}

func (a *App) AddRunners(runners ...Runner) {
	a.runners = append(a.runners, runners...)
}

func (a *App) AddDefers(defers ...Defer) {
	a.defers = append(a.defers, defers...)
}

func (a *App) Run(ctx context.Context) error {
	a.Logger.Info("Starting application")

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	// start metrics
	group.Go(func() error {
		if err := a.MetricsServer.Run(ctx); err != nil {
			a.Logger.WithError(err).Error("metrics server")
			return err
		}
		return nil
	})

	// run all runners
	for _, runner := range a.runners {
		rn := runner
		group.Go(func() error {
			return rn(ctx)
		})
	}

	defer func() {
		a.Logger.Info("Application stopped")
		_ = a.Logger.Sync()
	}()

	// defer all
	for _, d := range a.defers {
		defer d()
	}

	return group.Wait()
}
