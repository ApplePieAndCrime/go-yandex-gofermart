package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/utils"
)

func (s *Service) UploadOrder(ctx context.Context, userID int, number string) (string, error) {
	if !utils.IsValidLuhn(number) {
		return "", ErrInvalidOrderNumber
	}
	isNew, err := s.repo.InsertOrder(ctx, number, userID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderAlreadyExists) {
			return "", ErrOrderAlreadyExists
		}
		return "", err
	}
	if !isNew {
		return OrderAlreadyUploaded, nil
	}

	return OrderNewAccepted, nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int) ([]model.OrderResponse, error) {
	orders, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, o := range orders {
		responses[i] = o.ToResponse()
	}
	return responses, nil
}
