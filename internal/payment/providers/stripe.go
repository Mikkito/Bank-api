package providers

import (
	"bank-api/internal/models"
	"context"
	"errors"
	"fmt"
)

type StripeProvider struct {
	APIKey string
}

func NewStripeProvider(apiKey string) *StripeProvider {
	return &StripeProvider{
		APIKey: apiKey,
	}
}

func (y *StripeProvider) ProcessPayment(ctx context.Context, payment models.Payment) (*models.PaymentResult, error) {
	if payment.Amount <= 0 {
		return nil, errors.New("invalid payment amount")
	}

	// Имитируем успешный платеж
	transactionID := fmt.Sprintf("stripe_txn_%d", payment.ID)

	return &models.PaymentResult{
		TransactionID: transactionID,
		Status:        "success",
	}, nil
}

func (s *StripeProvider) Refund(ctx context.Context, paymentID string, amount float64) error {
	// Тут тоже могла бы быть интеграция с Stripe API для возврата
	if amount <= 0 {
		return errors.New("invalid refund amount")
	}
	fmt.Printf("Stripe: refunded %.2f for %s\n", amount, paymentID)
	return nil
}

func (s *StripeProvider) Name() string {
	return "stripe"
}
