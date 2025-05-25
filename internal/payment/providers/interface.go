package providers

import (
	"bank-api/internal/models"
	"context"
)

type PaymentProvider interface {
	Name() string
	ProcessPayment(ctx context.Context, payment models.Payment) (*models.PaymentResult, error)
	Refund(ctx context.Context, transactionID string, amount float64) error
}
