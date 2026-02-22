package migrations

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func Run(db *sql.DB) error {
	goose.SetBaseFS(Files)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
