package repository

import (
	"context"
	"errors"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrOrderAlreadyExists = errors.New("order already exists (by another user)")
	ErrOrderNotFound      = errors.New("order not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrUserNotFound       = errors.New("user not found")
)

type Repository interface {
	// Пользователи
	CreateUser(ctx context.Context, login, hashedPassword string) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int) (*model.User, error)

	// Заказы
	InsertOrder(ctx context.Context, number string, userID int) (bool, error)
	UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error
	GetUserOrders(ctx context.Context, userID int) ([]model.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)

	// Баланс и списания
	GetUserBalance(ctx context.Context, userID int) (current float64, withdrawn float64, err error)
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error)

	// Для фонового воркера
	GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error)
	ProcessOrderAccrual(ctx context.Context, orderNumber string, userID int, status model.OrderStatus, accrual *float64) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}
