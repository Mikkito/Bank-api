package models

import "time"

type PaymentMethod struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	Type        string    `db:"type"`     // card, bank_account, crypto, wallet и т.п.
	Provider    string    `db:"provider"` // Visa, Mastercard, YooMoney и т.п.
	Token       string    `db:"token"`    // токенизированные данные (PCI DSS безопасный storage)
	CreatedAt   time.Time `db:"created_at"`
	IsActive    bool      `db:"is_active"`
	Description string    `db:"description"`
}
