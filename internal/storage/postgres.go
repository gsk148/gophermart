package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/labstack/gommon/log"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/gsk148/gophermart/internal/models"
	"github.com/gsk148/gophermart/internal/utils"
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

func (s *Storage) GetOrderByNumber(orderNumber string) (*models.GetOrdersResponse, error) {
	var (
		order models.GetOrdersResponse
	)
	query := `SELECT id, number, user_id, uploaded_time FROM orders WHERE number=$1`
	res := s.DB.QueryRow(query, orderNumber)
	err := res.Scan(&order.ID, &order.Number, &order.UserID, &order.UploadedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			{
				return nil, ErrNoDBResult
			}
		default:
			{
				return nil, err
			}
		}
	}
	return &order, nil
}

func (s *Storage) LoadOrder(orderNumber string, userID int) error {
	queryAccrual := `INSERT INTO ORDERS(number, user_id, uploaded_time) VALUES($1, $2, $3)`
	_, err := s.DB.Exec(queryAccrual,
		orderNumber,
		userID,
		time.Now())
	if err != nil {
		if pgerrcode.IsIntegrityConstraintViolation(string(err.(*pq.Error).Code)) {
			return ErrDuplicateValue
		}
		return err
	}

	return nil
}

func (s *Storage) GetOrdersByUserID(userID int) ([]*models.GetOrdersResponse, error) {
	var (
		order  models.GetOrdersResponse
		orders []*models.GetOrdersResponse
		err    error
	)
	query := "SELECT number, status, amount, uploaded_time from orders where user_id=$1 and  OPERATION_TYPE=$2"
	rows, err := s.DB.Query(query, userID, "ACCRUAL")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	for rows.Next() {
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
