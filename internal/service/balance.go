package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/utils"
)

func (s *Service) GetBalance(ctx context.Context, userID int) (*model.BalanceResponse, error) {
	current, withdrawn, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	return &model.BalanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}, nil
}

// Withdraw выполняет списание баллов.
// Возвращает ошибку, если недостаточно средств или номер невалидный.
func (s *Service) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {

	if !utils.IsValidLuhn(orderNumber) {
		return ErrInvalidOrderNumber
	}

	if sum <= 0 {
		return errors.New("sum must be positive")
	}

	err := s.repo.Withdraw(ctx, userID, orderNumber, sum)
	if err != nil {
		return err
	}
	return nil
}

// GetWithdrawals возвращает историю списаний пользователя.
func (s *Service) GetWithdrawals(ctx context.Context, userID int) ([]model.WithdrawalResponse, error) {
	withdrawals, err := s.repo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}

	responses := make([]model.WithdrawalResponse, len(withdrawals))
	for i, w := range withdrawals {
		responses[i] = w.ToResponse()
	}
	return responses, nil
}
