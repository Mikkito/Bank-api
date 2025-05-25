package payment

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PaymentRepository struct {
	DB *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

// Создание нового платежа
func (r *PaymentRepository) CreatePayment(ctx context.Context, p *models.Payment) error {
	query := `INSERT INTO payments (user_id, provider, amount, currency, status, created_at)
	          VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id`
	return r.DB.QueryRowContext(ctx, query, p.UserID, p.Provider, p.Amount, p.Currency, p.Status).Scan(&p.ID)
}

// Обновление статуса
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE payments SET status = $1 WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, status, id)
	return err
}

// Завершить платеж
func (r *PaymentRepository) MarkAsCompleted(ctx context.Context, id int64, transactionID string) error {
	query := `UPDATE payments SET status = 'completed', transaction_id = $1 WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, transactionID, id)
	return err
}

// Получить платеж по ID
func (r *PaymentRepository) GetPaymentByID(ctx context.Context, id int64) (*models.Payment, error) {
	query := `SELECT id, user_id, provider, amount, currency, status, transaction_id, created_at, refunded_at
	          FROM payments WHERE id = $1`

	row := r.DB.QueryRowContext(ctx, query, id)

	var p models.Payment
	err := row.Scan(
		&p.ID,
		&p.UserID,
		&p.Provider,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.TransactionID,
		&p.CreatedAt,
		&p.IsRefunded,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, err
	}

	return &p, nil
}

// Пометить как возвращённый
func (r *PaymentRepository) MarkAsRefunded(ctx context.Context, id int64) error {
	query := `UPDATE payments SET status = 'refunded', refunded_at = NOW() WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *PaymentRepository) CreatePaymentMethod(ctx context.Context, method *models.PaymentMethod) error {
	query := `
		INSERT INTO payment_methods (user_id, type, token, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return r.DB.QueryRowContext(ctx, query, method.UserID, method.Type, method.Token, method.CreatedAt).
		Scan(&method.ID)
}

func (r *PaymentRepository) GetPaymentMethodsByUser(ctx context.Context, userID int64) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	query := `SELECT id, user_id, type, token, created_at FROM payment_methods WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.DB.SelectContext(ctx, &methods, query, userID)
	return methods, err
}

func (r *PaymentRepository) GetPaymentMethodByID(ctx context.Context, id int64) (*models.PaymentMethod, error) {
	var method models.PaymentMethod
	query := `SELECT id, user_id, provider, card_last4, created_at FROM payment_methods WHERE id = $1`
	err := r.DB.GetContext(ctx, &method, query, id)
	if err != nil {
		return nil, err
	}
	return &method, nil
}

func (r *PaymentRepository) DeletePaymentMethod(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM payment_methods WHERE id = $1`, id)
	return err
}
