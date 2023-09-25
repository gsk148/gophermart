package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gsk148/gophermart/internal/enums"
	"github.com/gsk148/gophermart/internal/errors"
	"github.com/gsk148/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"golang.org/x/exp/slog"
)

type OrderPostgres struct {
	ctx context.Context
	db  *sql.DB
	log slog.Logger
}

func NewOrderPostgres(ctx context.Context, db *sql.DB, log slog.Logger) *OrderPostgres {
	return &OrderPostgres{
		ctx: ctx,
		db:  db,
		log: log,
	}
}

func (p OrderPostgres) GetOrderByNumber(orderNumber string) (*models.GetOrdersResponse, error) {
	p.log.Info("GetOrderByNumber: provided order num is ", orderNumber)
	var (
		order models.GetOrdersResponse
	)
	query := `SELECT id, number, user_id, uploaded_time FROM orders WHERE number=$1`
	res := p.db.QueryRow(query, orderNumber)
	err := res.Scan(&order.ID, &order.Number, &order.UserID, &order.UploadedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			{
				p.log.Info("GetOrderByNumber err is", err)
				return nil, errors.ErrNoDBResult
			}
		default:
			{
				p.log.Info("GetOrderByNumber err is", err)
				return nil, err
			}
		}
	}
	return &order, nil
}

func (p OrderPostgres) LoadOrder(orderNumber string, userID uint) error {
	queryAccrual := `INSERT INTO orders(number, user_id, uploaded_time) VALUES($1, $2, $3)`
	queryWithdrawal := `INSERT INTO orders(number, user_id, uploaded_time, operation_type) VALUES($1, $2, $3, $4)`

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = fmt.Errorf("LoadOrder: failed to rollback %s", txErr.Error())
			}
		}
	}()

	_, err = tx.Exec(queryAccrual,
		orderNumber,
		userID,
		time.Now())
	if err != nil {
		if pgerrcode.IsIntegrityConstraintViolation(string(err.(*pq.Error).Code)) {
			tx.Rollback()
			return errors.ErrDuplicateValue
		}
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(queryWithdrawal,
		orderNumber,
		userID,
		time.Now(),
		enums.Withdrawal)
	if err != nil {
		if pgerrcode.IsIntegrityConstraintViolation(string(err.(*pq.Error).Code)) {
			return errors.ErrDuplicateValue
		}
		p.log.Error("DB LoadOrder: failed to exec query insert into orders, %s", err)
		return err
	}
	return tx.Commit()
}

func (p OrderPostgres) GetOrdersByUserID(userID uint) ([]*models.GetOrdersResponse, error) {
	var (
		order  models.GetOrdersResponse
		orders []*models.GetOrdersResponse
		err    error
	)
	query := "SELECT number, status, amount, uploaded_time from orders where user_id=$1 and  operation_type=$2"
	rows, err := p.db.Query(query, userID, enums.Accrual)
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

func (p OrderPostgres) GetOrdersForProcessing(poolSize int) ([]string, error) {
	var orders []string
	rows, err := p.db.Query(
		"SELECT number FROM orders WHERE status IN ($1, $2, $3) and operation_type=$4 ORDER BY uploaded_time LIMIT $5",
		enums.StatusNew,
		enums.StatusProcessing,
		enums.StatusRegistered,
		enums.Accrual,
		poolSize,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
	}()

	for rows.Next() {
		var orderID string
		if err = rows.Scan(&orderID); err != nil {
			return orders, err
		}
		orders = append(orders, orderID)
	}
	err = rows.Err()
	return orders, err
}

func (p OrderPostgres) UpdateOrderStateInvalid(order *models.GetOrderAccrual) error {
	query := "UPDATE orders set status=$1 where number=$2"
	_, err := p.db.Exec(query, enums.StatusInvalid, order.Order)
	if err != nil {
		p.log.Error("UpdateOrderStateInvalid: %s", err)
		return err
	}
	return nil
}

func (p OrderPostgres) UpdateOrderStateProcessed(order *models.GetOrderAccrual) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = fmt.Errorf("LoadOrder: failed to rollback %s", txErr.Error())
			}
		}
	}()

	var status string
	res := tx.QueryRow("SELECT status FROM orders WHERE number=$1 FOR UPDATE", order.Order)
	err = res.Scan(&status)

	if status == enums.StatusProcessed {
		tx.Commit()
		return nil
	}

	_, err = tx.Exec(
		"UPDATE orders SET status=$1, amount=$2 WHERE number = $3 and operation_type=$4",
		order.Status,
		order.Accrual,
		order.Order,
		enums.Accrual,
	)

	_, err = tx.Exec("UPDATE users SET balance=(select sum(amount) from orders where status=$1) WHERE id=(select distinct user_id from orders where number=$2)",
		enums.StatusProcessed,
		order.Order)
	return tx.Commit()
}
