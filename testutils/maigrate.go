package testutils

import (
	"database/sql"
	"io/fs"

	"github.com/pressly/goose/v3"
)

func Migrate(db *sql.DB, path fs.FS) error {
	goose.SetBaseFS(path)
	return goose.Up(db, "migrations")
}
