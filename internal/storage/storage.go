package storage

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var (
	ErrNoDBResult = errors.New("not found record in DB")
)

type Storage struct {
	DB *sql.DB
}
