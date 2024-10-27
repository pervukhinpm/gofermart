package models

import "time"

type Withdrawal struct {
	UserUUID    string
	OrderID     string    `json:"order"`
	Amount      float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
