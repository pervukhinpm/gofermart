package models

import "time"

type Order struct {
	OrderNumber int         `json:"id"`
	Status      OrderStatus `json:"status"`
	ProcessedAt time.Time   `json:"processed_at"`
}
