package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/jackc/pgx/v5"
)

func (r *repository) InsertOrder(ctx context.Context, number string, userID int) (bool, error) {
	query := `
        INSERT INTO orders (number, user_id, status, uploaded_at, updated_at)
        VALUES ($1, $2, 'NEW', NOW(), NOW())
        ON CONFLICT (number) DO NOTHING
    `
	cmdTag, err := r.pool.Exec(ctx, query, number, userID)
	if err != nil {
		return false, fmt.Errorf("insert order: %w", err)
	}

	if cmdTag.RowsAffected() == 1 {
		return true, nil
	}

	existing, err := r.GetOrderByNumber(ctx, number)
	if err != nil {
		return false, fmt.Errorf("check existing order: %w", err)
	}

	if existing.UserID == userID {
		return false, nil
	}
	return false, ErrOrderAlreadyExists
}

func (r *repository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE number = $1
	`
	var o model.Order
	err := r.pool.QueryRow(ctx, query, number).Scan(
		&o.ID,
		&o.Number,
		&o.UserID,
		&o.Status,
		&o.Accrual,
		&o.UploadedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by number: %w", err)
	}
	return &o, nil
}

func (r *repository) UpdateOrderStatus(ctx context.Context, number string, status string, accrual *float64) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE number = $3
	`
	_, err := r.pool.Exec(ctx, query, status, accrual, number)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func (r *repository) GetUserOrders(ctx context.Context, userID int) ([]model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		err := rows.Scan(
			&o.ID,
			&o.Number,
			&o.UserID,
			&o.Status,
			&o.Accrual,
			&o.UploadedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return orders, nil
}

func (r *repository) GetPendingOrders(ctx context.Context, limit int) ([]model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at, updated_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		  AND updated_at < NOW() - INTERVAL '10 seconds'
		ORDER BY uploaded_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		err := rows.Scan(
			&o.ID,
			&o.Number,
			&o.UserID,
			&o.Status,
			&o.Accrual,
			&o.UploadedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pending order: %w", err)
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return orders, nil
}

// ProcessOrderAccrual обновляет статус заказа и начисляет баллы пользователю в одной транзакции.
// Используется воркером при получении финального статуса PROCESSED.
// Если accrual == nil или статус не PROCESSED – просто обновляет статус (без начисления).
func (r *repository) ProcessOrderAccrual(ctx context.Context, orderNumber string, userID int, status model.OrderStatus, accrual *float64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	queryOrder := `
		UPDATE orders
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE number = $3 AND user_id = $4
	`
	_, err = tx.Exec(ctx, queryOrder, status, accrual, orderNumber, userID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	// Если статус PROCESSED и есть начисление – начислить баллы пользователю
	if status == model.OrderStatusProcessed && accrual != nil && *accrual > 0 {
		queryBalance := `
			UPDATE users
			SET current_balance = current_balance + $1
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, queryBalance, *accrual, userID)
		if err != nil {
			return fmt.Errorf("update user balance: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
