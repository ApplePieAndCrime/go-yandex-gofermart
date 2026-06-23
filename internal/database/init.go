package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTables(ctx context.Context, pool *pgxpool.Pool) error {
	sql := `
	-- пользователи
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	ALTER TABLE users ADD COLUMN IF NOT EXISTS current_balance DECIMAL(10,2) NOT NULL DEFAULT 0;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS withdrawn_total DECIMAL(10,2) NOT NULL DEFAULT 0;

	-- заказы
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		number TEXT UNIQUE NOT NULL,
		user_id INT REFERENCES users(id) ON DELETE CASCADE,
		status TEXT NOT NULL,
		accrual DECIMAL(10,2),
		uploaded_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	-- списания
	CREATE TABLE IF NOT EXISTS withdrawals (
		id SERIAL PRIMARY KEY,
		order_number TEXT NOT NULL,
		user_id INT REFERENCES users(id) ON DELETE CASCADE,
		sum DECIMAL(10,2) NOT NULL,
		processed_at TIMESTAMP DEFAULT NOW()
	);
	`
	_, err := pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}
