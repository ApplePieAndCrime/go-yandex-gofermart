package service

import (
	"errors"
	"net/http"
	"time"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"go.uber.org/zap"
)

const (
	OrderNewAccepted     = "new_accepted"
	OrderAlreadyUploaded = "already_uploaded"
)

var (
	ErrLoginAlreadyExists   = errors.New("login already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrOrderAlreadyUploaded = errors.New("order already uploaded by this user")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidOrderNumber   = errors.New("invalid order number")
	ErrOrderAlreadyExists   = errors.New("order already exists (by another user)")
)

type Service struct {
	repo        repository.Repository
	accrualAddr string
	logger      *zap.SugaredLogger
	httpClient  *http.Client
}

func NewService(repo repository.Repository, accrualAddr string, logger *zap.SugaredLogger) *Service {
	return &Service{
		repo:        repo,
		accrualAddr: accrualAddr,
		logger:      logger.With("component", "service"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
