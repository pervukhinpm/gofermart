package repository

import "time"

type Withdrawal struct {
	OrderID     string    `json:"order"`
	Amount      int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
