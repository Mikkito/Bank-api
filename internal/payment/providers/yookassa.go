package providers

import (
	"bank-api/internal/models"
	"context"
	"errors"
	"fmt"
)

type YooMoneyProvider struct {
	Token string
}

func NewYooMoneyProvider(token string) *YooMoneyProvider {
	return &YooMoneyProvider{
		Token: token,
	}
}

func (y *YooMoneyProvider) ProcessPayment(ctx context.Context, payment models.Payment) (*models.PaymentResult, error) {
	if payment.Amount <= 0 {
		return nil, errors.New("invalid payment amount")
	}

	// Имитируем успешный платеж
	transactionID := fmt.Sprintf("yoomoney_txn_%d", payment.ID)

	return &models.PaymentResult{
		TransactionID: transactionID,
		Status:        "success",
	}, nil
}

func (y *YooMoneyProvider) Refund(ctx context.Context, paymentID string, amount float64) error {
	if amount <= 0 {
		return errors.New("invalid refund amount")
	}
	fmt.Printf("YooMoney: refunded %.2f for %s\n", amount, paymentID)
	return nil
}

func (y *YooMoneyProvider) Name() string {
	return "yoomoney"
}
