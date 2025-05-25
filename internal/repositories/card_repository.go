package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type CardRepository struct {
	DB *sqlx.DB
}

func (r *CardRepository) CreateCard(ctx context.Context, card *models.Card) error {
	query := `
		INSERT INTO cards (account_id, encrypted_data, hmac, cvv, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	err := r.DB.QueryRowContext(
		ctx,
		query,
		card.AccountID,
		card.EncryptedData,
		card.HMAC,
		card.CVV,
	).Scan(&card.ID, &card.CreatedAt)
	return err
}

func (r *CardRepository) GetCardByID(ctx context.Context, cardID int64) (*models.Card, error) {
	query := `
		SELECT id, account_id, encrypted_data, hmac, cvv, created_at
		FROM cards
		WHERE id = $1
	`
	var card models.Card
	err := r.DB.QueryRowContext(ctx, query, cardID).Scan(
		&card.ID,
		&card.AccountID,
		&card.EncryptedData,
		&card.HMAC,
		&card.CVV,
		&card.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // карта не найдена
		}
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) GetCardsByAccountID(ctx context.Context, accountID int64) ([]*models.Card, error) {
	query := `
		SELECT id, account_id, encrypted_data, hmac, cvv, created_at
		FROM cards
		WHERE account_id = $1
	`
	rows, err := r.DB.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.EncryptedData,
			&card.HMAC,
			&card.CVV,
			&card.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

func (r *CardRepository) GetCardsByUser(ctx context.Context, userID int64) ([]*models.Card, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, account_id, encrypted_data, hmac, cvv, created_at FROM cards WHERE account_id IN (SELECT id FROM accounts WHERE user_id = $1)`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(&card.ID, &card.AccountID, &card.EncryptedData, &card.HMAC, &card.CVV, &card.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return cards, nil
}

func (r *CardRepository) DeleteCard(ctx context.Context, cardID int64) error {
	result, err := r.DB.ExecContext(ctx, `DELETE FROM cards WHERE id = $1`, cardID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("card not found")
	}
	return nil
}

func (r *CardRepository) SetCardStatus(ctx context.Context, cardID int64, status string) error {
	query := `UPDATE cards SET status = $1 WHERE id = $2`
	result, err := r.DB.ExecContext(ctx, query, status, cardID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("card not found")
	}
	return nil
}
