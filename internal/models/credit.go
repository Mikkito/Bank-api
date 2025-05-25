package models

import "time"

type Credit struct {
	ID             int64     `json:"id"`
	AccountID      int64     `json:"account_id"`
	Amount         float64   `json:"amount"`
	InterestRate   float64   `json:"interest_rate"`
	TermMonths     int       `json:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment"`
	CreateAt       time.Time `json:"create_at"`
	NextPayment    time.Time `json:"next_payment"`
	Status         string    `json:"status"` // active, closed, overdue
}
