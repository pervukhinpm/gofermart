package model

import "time"

type Withdrawal struct {
	UserID      string
	OrderID     string    `json:"order"`
	Amount      float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
