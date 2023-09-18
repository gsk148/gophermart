package storage

import (
	"database/sql"
	"fmt"

	"github.com/gsk148/gophermart/internal/utils"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/gsk148/gophermart/internal/models"
)

func InitStorage(conn string) (*Storage, error) {
	const op = "storage.Init"
	db, err := sql.Open("postgres", conn)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = goose.SetDialect("postgres"); err != nil {
		log.Errorf("Init DB: failed while goose set dialect, %s", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err = goose.Up(db, "migrations"); err != nil {
		log.Errorf("Init DB: failed while goose up, %s", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) GetUserByLogin(login string) (*models.User, error) {
	query := `SELECT id, login, password FROM USERS WHERE LOGIN=$1`
	res := s.DB.QueryRow(query, login)
	var user models.User
	err := res.Scan(&user.ID, &user.Login, &user.Password)
	switch {
	case err == sql.ErrNoRows:
		return nil, ErrNoDBResult
	case err != nil:
		return nil, err
	default:
		return &user, nil
	}
}

func (s *Storage) Login(user models.User) error {
	userInDB, err := s.GetUserByLogin(user.Login)
	if err != nil {
		return err
	}
	if !utils.CheckHashAndPassword(userInDB.Password, user.Password) {
		return err
	}
	return nil
}

func (s *Storage) Register(login, password string) (uint, error) {
	var id uint
	query :=
		`INSERT INTO USERS(login, password) VALUES($1, $2) RETURNING Users.id;`
	res := s.DB.QueryRow(query, login, password)
	err := res.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
