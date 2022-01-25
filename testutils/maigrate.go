package testutils

import (
	"database/sql"
	"github.com/pressly/goose"
)

const migrationPath = "db/migrate"

func migrate(db *sql.DB) error {
	return goose.Up(db, migrationPath)
}
