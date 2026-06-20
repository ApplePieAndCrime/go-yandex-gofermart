package worker

import (
	"context"
	"time"

	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/accrual"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/model"
	"github.com/ApplePieAndCrime/go-yandex-gofermart/internal/repository"
	"go.uber.org/zap"
)

// Worker отвечает за фоновую обработку заказов: опрашивает внешний сервис
// и обновляет статусы заказов в БД.
type Worker struct {
	repo     repository.Repository
	client   *accrual.Client
	logger   *zap.SugaredLogger
	interval time.Duration
	stopCh   chan struct{}
}

func NewWorker(
	repo repository.Repository,
	client *accrual.Client,
	logger *zap.SugaredLogger,
	interval time.Duration,
) *Worker {
	return &Worker{
		repo:     repo,
		client:   client,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopped by context")
			return
		case <-w.stopCh:
			w.logger.Info("worker stopped by stop signal")
			return
		case <-ticker.C:
			if err := w.processPendingOrders(ctx); err != nil {
				w.logger.Errorf("worker processing error: %v", err)
			}
		}
	}
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

func (w *Worker) processPendingOrders(ctx context.Context) error {
	const limit = 10

	orders, err := w.repo.GetPendingOrders(ctx, limit)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	w.logger.Infof("processing %d pending orders", len(orders))

	for _, order := range orders {
		if err := w.processOrder(ctx, order); err != nil {
			w.logger.Errorf("failed to process order %s: %v", order.Number, err)
		}
	}
	return nil
}

func (w *Worker) processOrder(ctx context.Context, order model.Order) error {
	resp, err := w.client.GetAccrual(ctx, order.Number)
	if err != nil {
		return err
	}

	var newStatus model.OrderStatus
	var accrualAmount *float64

	switch resp.Status {
	case model.AccrualStatusProcessed:
		newStatus = model.OrderStatusProcessed
		accrualAmount = resp.Accrual
	case model.AccrualStatusInvalid:
		newStatus = model.OrderStatusInvalid
		accrualAmount = nil
	case model.AccrualStatusProcessing, model.AccrualStatusRegistered:
		if order.Status == model.OrderStatusNew {
			newStatus = model.OrderStatusProcessing
		} else {
			newStatus = order.Status
		}
		accrualAmount = order.Accrual
	default:
		w.logger.Warnf("unknown accrual status %s for order %s", resp.Status, order.Number)
		return nil
	}

	if newStatus != order.Status || (newStatus == model.OrderStatusProcessed && accrualAmount != nil) {
		err = w.repo.ProcessOrderAccrual(ctx, order.Number, order.UserID, newStatus, accrualAmount)
		if err != nil {
			return err
		}
		w.logger.Infof("order %s updated: status=%s, accrual=%v", order.Number, newStatus, accrualAmount)
	}
	return nil
}
