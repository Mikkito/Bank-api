package models

import "time"

type CardResponse struct {
	ID             int64     `json:"id"`
	AccountID      int64     `json:"account_id"`
	CardNumber     string    `json:"card_number"`
	ExpirationDate string    `json:"expiration_date"` // формат MM/YY
	CreatedAt      time.Time `json:"created_at"`
}
