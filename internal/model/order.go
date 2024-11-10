package model

import "time"

type Order struct {
	OrderNumber string      `json:"id"`
	UserID      string      `json:"user_id"`
	Status      OrderStatus `json:"status"`
	ProcessedAt time.Time   `json:"processed_at"`
	Accrual     int         `json:"accrual"`
}
