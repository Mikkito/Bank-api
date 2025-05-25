package models

import "time"

type Card struct {
	ID             int64     `json:"id"`
	AccountID      int64     `json:"account_id"`
	EncryptedData  string    `json:"-"`
	HMAC           string    `json:"-"`
	CVV            string    `json:"cvv,omitempty"` // показывать только в нужных случаях
	CardNumber     string    `json:"card_number,omitempty"`
	ExpirationDate time.Time `json:"expiration_date,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
