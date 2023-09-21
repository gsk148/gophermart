package storage

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var (
	ErrNoDBResult     = errors.New("not found record in DB")
	ErrDuplicateValue = errors.New("duplicate value while insert")
)

type Storage struct {
	DB *sql.DB
}
