package model

import "time"

type Order struct {
	OrderNumber string      `json:"id"`
	UserID      string      `json:"user_id"`
	Status      OrderStatus `json:"status"`
	ProcessedAt time.Time   `json:"processed_at"`
	Accrual     float64     `json:"accrual"`
}

type OrderResponse struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    float64     `json:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at"`
}
