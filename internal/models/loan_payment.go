package models

import "time"

type LoanPayment struct {
	ID        int64     `db:"id" json:"id"`
	LoanID    int64     `db:"loan_id" json:"loan_id"`
	Amount    float64   `db:"amount" json:"amount"`
	PaidAt    time.Time `db:"paid_at" json:"paid_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
