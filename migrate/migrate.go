package migrate

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

type options struct {
	driverName string
	path       string
}

type OptionsFunc func(opts *options)

func WithPath(path string) OptionsFunc {
	return func(opts *options) {
		opts.path = path
	}
}

func WithDriver(driver string) OptionsFunc {
	return func(opts *options) {
		opts.driverName = driver
	}
}

func WithFs(fs embed.FS) OptionsFunc {
	return func(opts *options) {
		goose.SetBaseFS(fs)
	}
}

func Run(dsn string, opts ...OptionsFunc) error {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	// default driver. If you want to change it, use option WithDriver
	curDriverName := "pgx"
	if o.driverName != "" {
		curDriverName = o.driverName
	}

	sqlDb, err := sql.Open(curDriverName, dsn)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}
	defer func() { _ = sqlDb.Close() }()

	// migrations default path is 'migrations' directory
	// if you want to change it, use option WithPath
	curPath := "migrations"
	if o.path != "" {
		curPath = o.path
	}

	if err = goose.Up(sqlDb, curPath); err != nil {
		return fmt.Errorf("up migrations: %w", err)
	}

	return nil
}
