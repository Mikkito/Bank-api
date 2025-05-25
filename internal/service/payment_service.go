package service

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"errors"
)

type PaymentService struct {
	repo *repositories.PaymentMethodRepository
}

func NewPaymentService(repo *repositories.PaymentMethodRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// AddPaymentMethod добавляет новый метод оплаты
func (s *PaymentService) AddPaymentMethod(ctx context.Context, method *models.PaymentMethod) error {
	if method.UserID == 0 || method.Type == "" || method.Provider == "" || method.Token == "" {
		return errors.New("missing required payment method fields")
	}
	method.IsActive = true
	return s.repo.Create(ctx, method)
}

// GetActiveMethodsByUser возвращает все активные методы пользователя
func (s *PaymentService) GetActiveMethodsByUser(ctx context.Context, userID int64) ([]models.PaymentMethod, error) {
	if userID == 0 {
		return nil, errors.New("invalid user id")
	}
	return s.repo.GetByUserID(ctx, userID)
}

// GetPaymentMethodByID возвращает конкретный метод по ID
func (s *PaymentService) GetPaymentMethodByID(ctx context.Context, id int64) (*models.PaymentMethod, error) {
	if id == 0 {
		return nil, errors.New("invalid payment method id")
	}
	return s.repo.GetByID(ctx, id)
}

// DeactivateMethod отключает метод оплаты
func (s *PaymentService) DeactivateMethod(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("invalid payment method id")
	}
	return s.repo.Deactivate(ctx, id)
}
