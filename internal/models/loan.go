package models

import "time"

type Loan struct {
	ID             int64     `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	AccountID      int64     `db:"account_id" json:"account_id"`
	Principal      float64   `db:"principal" json:"principal"`         // Основная сумма кредита
	InterestRate   float64   `db:"interest_rate" json:"interest_rate"` // Годовая ставка
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	IsRepaid       bool      `db:"is_repaid" json:"is_repaid"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
	StartDate      time.Time `db:"start_date" json:"start_date"`             // дата
	NextPaymentDue time.Time `db:"next_payment_due" json:"next_payment_due"` // дата следующего платежа
}
