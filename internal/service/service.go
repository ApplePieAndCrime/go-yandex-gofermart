package service

import (
	"net/http"
	"time"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"go.uber.org/zap"
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
