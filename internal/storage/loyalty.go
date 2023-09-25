package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/exp/slog"

	"github.com/gsk148/gophermart/internal/enums"
	"github.com/gsk148/gophermart/internal/errors"
	"github.com/gsk148/gophermart/internal/models"
)

type LoyaltyPostgres struct {
	ctx context.Context
	db  *sql.DB
	log slog.Logger
}

func NewLoyaltyPostgres(ctx context.Context, db *sql.DB, log slog.Logger) *LoyaltyPostgres {
	return &LoyaltyPostgres{
		ctx: ctx,
		db:  db,
		log: log,
	}
}

func (p LoyaltyPostgres) DeductPoints(w models.WithdrawBalanceRequest, userID uint, orderNumber string) error {
	queryUpdateCurrentBalance := `UPDATE users SET balance=balance-$1 WHERE id=$2`
	queryUpdateWithdrawal := `UPDATE orders SET amount=$1 WHERE user_id=$2 AND number=$3 AND operation_type=$4`

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
	_, err = tx.ExecContext(p.ctx, queryUpdateCurrentBalance,
		w.Sum,
		userID)

	_, err = tx.ExecContext(p.ctx, queryUpdateWithdrawal,
		w.Sum,
		userID,
		orderNumber,
		enums.Withdrawal)
	return tx.Commit()
}

func (p LoyaltyPostgres) GetWithdrawals(userID uint) ([]*models.GetWithdrawalsResponse, error) {
	var (
		withdrawal models.GetWithdrawalsResponse
		response   []*models.GetWithdrawalsResponse
		err        error
	)
	queryGet := `SELECT number, amount, uploaded_time
			  FROM orders
			  WHERE user_id=$1 AND operation_type=$2 ORDER BY uploaded_time`

	res, err := p.db.Query(queryGet, userID, enums.Withdrawal)
	if res.Err() != nil {
		return nil, res.Err()
	}
	if err != nil {
		return nil, err
	}
	for res.Next() {
		err = res.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawal.ProcessedAt.Format(time.RFC3339)
		if withdrawal.Sum == 0 {
			continue
		}
		response = append(response, &withdrawal)
	}
	return response, nil
}

func (p LoyaltyPostgres) GetBalance(userID uint) (*models.GetBalanceResponse, error) {
	var balance models.GetBalanceResponse
	current := `SELECT balance FROM users WHERE id=$1`
	withdrawn := `SELECT sum(amount) FROM orders where user_id=$1 and operation_type=$2`

	resAcc := p.db.QueryRow(current, userID)
	if resAcc.Err() != nil {
		return nil, resAcc.Err()
	}
	err := resAcc.Scan(&balance.Current)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.ErrNoDBResult
		default:
			return nil, err
		}
	}

	resWith := p.db.QueryRow(withdrawn, userID, enums.Withdrawal)
	if resWith.Err() != nil {
		return nil, resWith.Err()
	}
	err = resWith.Scan(&balance.Withdrawn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.ErrNoDBResult
		default:
			return nil, err
		}
	}
	return &balance, nil
}
