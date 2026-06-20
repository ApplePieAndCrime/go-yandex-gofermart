package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/jackc/pgx/v5"
)

func (r *repository) GetUserBalance(ctx context.Context, userID int) (float64, float64, error) {
	query := `SELECT current_balance, withdrawn_total FROM users WHERE id = $1`
	var current, withdrawn float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&current, &withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, ErrUserNotFound
		}
		return 0, 0, fmt.Errorf("get user balance: %w", err)
	}
	return current, withdrawn, nil
}

func (r *repository) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentBalance float64
	queryLock := `SELECT current_balance FROM users WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryLock, userID).Scan(&currentBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("lock user: %w", err)
	}

	if currentBalance < sum {
		return ErrInsufficientFunds
	}

	queryUpdate := `
		UPDATE users
		SET current_balance = current_balance - $1,
		    withdrawn_total = withdrawn_total + $1
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, queryUpdate, sum, userID)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	queryInsert := `
		INSERT INTO withdrawals (order_number, user_id, sum, processed_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err = tx.Exec(ctx, queryInsert, orderNumber, userID, sum)
	if err != nil {
		return fmt.Errorf("insert withdrawal: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *repository) GetUserWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	query := `
		SELECT id, order_number, user_id, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		err := rows.Scan(
			&w.ID,
			&w.OrderNumber,
			&w.UserID,
			&w.Sum,
			&w.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return withdrawals, nil
}
