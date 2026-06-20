package model

import "time"

type Withdrawal struct {
	ID          int       // идентификатор
	OrderNumber string    // номер заказа, в счёт которого списаны баллы
	UserID      int       // владелец списания
	Sum         float64   // списанная сумма
	ProcessedAt time.Time // время списания
}

func (w Withdrawal) ToResponse() WithdrawalResponse {
	return WithdrawalResponse{
		Order:       w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	}
}
