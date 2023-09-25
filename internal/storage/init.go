package storage

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"golang.org/x/exp/slog"
)

func InitStorage(conn string, logger slog.Logger) (*sql.DB, error) {
	const op = "storage.Init"
	db, err := sql.Open("postgres", conn)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = goose.SetDialect("postgres"); err != nil {
		logger.Error("Init DB: failed while goose set dialect, %s", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err = goose.Up(db, "migrations"); err != nil {
		logger.Error("Init DB: failed while goose up, %s", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}
