package model

import "time"

type Order struct {
	ID         int         // идентификатор записи
	Number     string      // номер заказа (уникален)
	UserID     int         // владелец заказа
	Status     OrderStatus // статус: NEW, PROCESSING, INVALID, PROCESSED
	Accrual    *float64    // начисленные баллы (nil, если не рассчитано или не начислено)
	UploadedAt time.Time   // время загрузки
	UpdatedAt  time.Time   // время последнего обновления (для воркера)
}

func (o Order) ToResponse() OrderResponse {
	return OrderResponse{
		Number:     o.Number,
		Status:     o.Status,
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}
}
