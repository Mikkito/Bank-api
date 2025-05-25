package models

import "time"

type Payment struct {
	ID            int64     `db:"id" json:"id"`
	UserID        int64     `db:"user_id" json:"user_id"`
	AccountID     int64     `db:"account_id" json:"account_id"`
	Amount        float64   `db:"amount" json:"amount"`
	Currency      string    `db:"currency" json:"currency"`
	Method        string    `db:"method" json:"method"`     // eg: "card", "yoomoney", "stripe"
	Provider      string    `db:"provider" json:"provider"` // eg: "stripe", "yoomoney"
	Status        string    `db:"status" json:"status"`     // "pending", "completed", "failed"
	TransactionID string    `db:"transaction_id" json:"transaction_id"`
	IsRefunded    bool      `db:"is_refunded" json:"is_refunded"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}
