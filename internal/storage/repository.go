package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gsk148/gophermart/internal/models"
	_ "github.com/lib/pq"
	"golang.org/x/exp/slog"
)

var (
	ErrNoDBResult     = errors.New("not found record in DB")
	ErrDuplicateValue = errors.New("duplicate value while insert")
)

type Storage struct {
	DB *sql.DB
}

type User interface {
	GetUserByLogin(login string) (*models.User, error)
	Register(login, password string) (uint, error)
}

type Order interface {
	GetOrderByNumber(orderNumber string) (*models.GetOrdersResponse, error)
	LoadOrder(orderNumber string, userID int) error
	GetOrdersByUserID(userID int) ([]*models.GetOrdersResponse, error)
	GetOrdersForProcessing(poolSize int) ([]string, error)
	UpdateOrderStateProcessed(order *models.GetOrderAccrual) error
	UpdateOrderStateInvalid(order *models.GetOrderAccrual) error
}

type Loyalty interface {
	DeductPoints(w models.WithdrawBalanceRequest, userID int, orderNumber string) error
	GetWithdrawals(userID int) ([]*models.GetWithdrawalsResponse, error)
	GetBalance(userID int) (*models.GetBalanceResponse, error)
}

type Repository struct {
	User
	Order
	Loyalty
}

func NewRepository(ctx context.Context, db *sql.DB, log slog.Logger) *Repository {
	return &Repository{
		User:    NewUserPostgres(ctx, db, log),
		Order:   NewOrderPostgres(ctx, db, log),
		Loyalty: NewLoyaltyPostgres(ctx, db, log),
	}
}
