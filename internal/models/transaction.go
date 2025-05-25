package models

import "time"

type Transaction struct {
	ID          int64     `json:"id"`
	FromAccount int64     `json:"from_account,omitempty"`
	ToAccount   int64     `json:"to_account,omitempty"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"` // deposit, transfer, withdraw, credit_payment
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description,omitempty"`
	IsReversal  bool      `json:"is_reversal"`
}
