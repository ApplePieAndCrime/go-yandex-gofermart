package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *repository) CreateUser(ctx context.Context, login, hashedPassword string) (int, error) {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
	var userID int
	err := r.pool.QueryRow(ctx, query, login, hashedPassword).Scan(&userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, ErrLoginAlreadyExists
		}
		return 0, fmt.Errorf("create user: %w", err)
	}
	return userID, nil
}

func (r *repository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `
		SELECT id, login, password_hash, current_balance, withdrawn_total, created_at
		FROM users
		WHERE login = $1
	`
	var u model.User
	err := r.pool.QueryRow(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.PasswordHash,
		&u.CurrentBalance,
		&u.WithdrawnTotal,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}
	return &u, nil
}

func (r *repository) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	query := `
		SELECT id, login, password_hash, current_balance, withdrawn_total, created_at
		FROM users
		WHERE id = $1
	`
	var u model.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&u.ID,
		&u.Login,
		&u.PasswordHash,
		&u.CurrentBalance,
		&u.WithdrawnTotal,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
