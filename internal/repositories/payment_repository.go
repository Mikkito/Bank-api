package repositories

import (
	"bank-api/internal/models"
	"context"

	"github.com/jmoiron/sqlx"
)

type PaymentMethodRepository struct {
	DB *sqlx.DB
}

func NewPaymentMethodRepository(db *sqlx.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{DB: db}
}

// Create сохраняет новый метод оплаты
func (r *PaymentMethodRepository) Create(ctx context.Context, method *models.PaymentMethod) error {
	query := `
		INSERT INTO payment_methods (user_id, type, provider, token, description, is_active)
		VALUES (:user_id, :type, :provider, :token, :description, :is_active)
		RETURNING id, created_at
	`
	stmt, err := r.DB.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.GetContext(ctx, method, method)
}

// GetByUserID возвращает все активные методы пользователя
func (r *PaymentMethodRepository) GetByUserID(ctx context.Context, userID int64) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	err := r.DB.SelectContext(ctx, &methods, `
		SELECT * FROM payment_methods
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`, userID)
	return methods, err
}

// GetByID возвращает метод по ID
func (r *PaymentMethodRepository) GetByID(ctx context.Context, id int64) (*models.PaymentMethod, error) {
	var method models.PaymentMethod
	err := r.DB.GetContext(ctx, &method, `SELECT * FROM payment_methods WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &method, nil
}

// Deactivate делает метод неактивным
func (r *PaymentMethodRepository) Deactivate(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE payment_methods SET is_active = false WHERE id = $1`, id)
	return err
}
