package models

import "time"

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type AccountBalance struct {
	AccountID int64     `json:"account_id"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}
