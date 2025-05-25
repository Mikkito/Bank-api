package payment

import (
	"bank-api/internal/models"
	"bank-api/internal/payment/providers"
	"bank-api/internal/service"
	"context"
	"errors"
	"fmt"
	"time"
)

type PaymentService struct {
	repo               *PaymentRepository
	providers          map[string]providers.PaymentProvider
	transactionService *service.TransactionService
}

func NewPaymentService(repo *PaymentRepository, provider []providers.PaymentProvider, transactionService *service.TransactionService) *PaymentService {
	providerMap := make(map[string]providers.PaymentProvider)
	for _, p := range provider {
		providerMap[p.Name()] = p
	}

	return &PaymentService{
		repo:               repo,
		providers:          providerMap,
		transactionService: transactionService,
	}
}

// ProcessPayment — обработка нового платежа
func (s *PaymentService) ProcessPayment(ctx context.Context, payment models.Payment) (*models.PaymentResult, error) {
	provider, ok := s.providers[payment.Provider]
	if !ok {
		return nil, fmt.Errorf("payment provider '%s' not found", payment.Provider)
	}

	// Запись платежа в базу (со статусом pending)
	payment.Status = "pending"
	if err := s.repo.CreatePayment(ctx, &payment); err != nil {
		return nil, err
	}

	// Отправка на оплату провайдеру
	result, err := provider.ProcessPayment(ctx, payment)
	if err != nil {
		// Ошибка провайдера — обновим статус
		_ = s.repo.UpdatePaymentStatus(ctx, payment.ID, "failed")
		return nil, err
	}

	txn := models.Transaction{
		FromAccount: 0, // внешний источник
		ToAccount:   payment.AccountID,
		Amount:      payment.Amount,
		Type:        "deposit",
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("External payment via %s", payment.Provider),
	}

	transactionID, err := s.transactionService.CreateTransaction(ctx, &txn)

	if err != nil {
		return nil, fmt.Errorf("payment successful, but failed to record transaction: %w", err)
	}

	// Обновим статус и свяжем транзакцию
	if err := s.repo.MarkAsCompleted(ctx, payment.ID, fmt.Sprintf("%d", transactionID)); err != nil {
		return nil, err
	}

	payment.TransactionID = fmt.Sprintf("%d", transactionID)
	result.TransactionID = payment.TransactionID

	return result, nil
}

// RefundPayment — возврат средств
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID int64) error {
	payment, err := s.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if payment.IsRefunded {
		return errors.New("payment already refunded")
	}

	provider, ok := s.providers[payment.Provider]
	if !ok {
		return fmt.Errorf("payment provider '%s' not found", payment.Provider)
	}

	if err := provider.Refund(ctx, payment.TransactionID, payment.Amount); err != nil {
		return err
	}

	// Обратная транзакция
	txn := models.Transaction{
		FromAccount: payment.AccountID,
		ToAccount:   0,
		Amount:      payment.Amount,
		Type:        "refund",
		IsReversal:  true,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Refund to %s for payment %d", payment.Provider, payment.ID),
	}

	_, err = s.transactionService.CreateTransaction(ctx, &txn)
	if err != nil {
		return fmt.Errorf("refund successful, but failed to record reversal: %w", err)
	}

	return s.repo.MarkAsRefunded(ctx, paymentID)
}

func (s *PaymentService) AddPaymentMethod(ctx context.Context, method *models.PaymentMethod) error {
	// простая валидация
	if method.UserID == 0 || method.Type == "" || method.Token == "" {
		return errors.New("invalid payment method")
	}

	method.CreatedAt = time.Now()
	return s.repo.CreatePaymentMethod(ctx, method)
}

func (s *PaymentService) GetPaymentMethods(ctx context.Context, userID int64) ([]models.PaymentMethod, error) {
	return s.repo.GetPaymentMethodsByUser(ctx, userID)
}

func (s *PaymentService) DeletePaymentMethod(ctx context.Context, userID int64, methodID int64) error {
	// Проверим, существует ли метод и принадлежит ли он пользователю
	method, err := s.repo.GetPaymentMethodByID(ctx, methodID)
	if err != nil {
		return err
	}

	if method.UserID != userID {
		return errors.New("unauthorized: payment method does not belong to the user")
	}

	return s.repo.DeletePaymentMethod(ctx, methodID)
}

func (s *PaymentService) GetPaymentByID(ctx context.Context, paymentID int64) (*models.Payment, error) {
	return s.repo.GetPaymentByID(ctx, paymentID)
}
