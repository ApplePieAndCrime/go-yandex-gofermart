package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/utils"
)

func (s *Service) RegisterUser(ctx context.Context, login, password string) (int, string, error) {
	hashed, err := utils.HashPassword(password)
	if err != nil {
		return 0, "", fmt.Errorf("hash password: %w", err)
	}

	userID, err := s.repo.CreateUser(ctx, login, hashed)
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExists) {
			return 0, "", err
		}
		return 0, "", fmt.Errorf("create user: %w", err)
	}

	token, err := s.jwt.GenerateToken(userID)
	if err != nil {
		return 0, "", fmt.Errorf("generate token: %w", err)
	}

	return userID, token, nil
}

func (s *Service) LoginUser(ctx context.Context, login, password string) (int, string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return 0, "", errors.New("invalid credentials")
		}
		return 0, "", fmt.Errorf("get user: %w", err)
	}

	if !utils.CheckPassword(user.PasswordHash, password) {
		return 0, "", errors.New("invalid credentials")
	}

	token, err := s.jwt.GenerateToken(user.ID)
	if err != nil {
		return 0, "", fmt.Errorf("generate token: %w", err)
	}

	return user.ID, token, nil
}
