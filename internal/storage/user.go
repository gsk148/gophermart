package storage

import (
	"context"
	"database/sql"

	"github.com/gsk148/gophermart/internal/errors"
	"github.com/gsk148/gophermart/internal/models"
	"golang.org/x/exp/slog"
)

type UserPostgres struct {
	ctx    context.Context
	db     *sql.DB
	logger slog.Logger
}

func NewUserPostgres(ctx context.Context, db *sql.DB, log slog.Logger) *UserPostgres {
	return &UserPostgres{
		ctx:    ctx,
		db:     db,
		logger: log,
	}
}

//func (s *Storage) Login(user models.User) error {
//	userInDB, err := s.GetUserByLogin(user.Login)
//	if err != nil {
//		return err
//	}
//	if !utils.CheckHashAndPassword(userInDB.Password, user.Password) {
//		return err
//	}
//	return nil
//}

func (u *UserPostgres) GetUserByLogin(login string) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password FROM USERS WHERE LOGIN=$1`
	res := u.db.QueryRowContext(u.ctx, query, login)
	err := res.Scan(&user.ID, &user.Login, &user.Password)
	switch {
	case err == sql.ErrNoRows:
		return nil, errors.ErrNoDBResult
	case err != nil:
		return nil, err
	default:
		return &user, nil
	}
}

func (u *UserPostgres) Register(login, password string) (uint, error) {
	var id uint
	query :=
		`INSERT INTO USERS(login, password) VALUES($1, $2) RETURNING Users.id;`
	res := u.db.QueryRowContext(u.ctx, query, login, password)
	err := res.Scan(&id)
	if err != nil {
		u.logger.Error("DB Register: failed to exec query, %s", err)
		return 0, err
	}
	return id, nil
}
