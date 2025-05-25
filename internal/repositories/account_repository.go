package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	DB *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

// CreateAccount inserts a new account for a given user ID
func (r *AccountRepository) CreateAccount(ctx context.Context, userID int64) (*models.Account, error) {
	account := &models.Account{
		UserID:  userID,
		Balance: 0,
	}

	query := `
		INSERT INTO accounts (user_id, balance)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.DB.QueryRowContext(ctx, query, account.UserID, account.Balance).
		Scan(&account.ID, &account.CreatedAt)

	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *AccountRepository) HasAccountByUserID(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM accounts WHERE user_id = $1)`
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&exists)
	return exists, err
}

func (r *AccountRepository) GetAccountBalance(ctx context.Context, accountID int64) (float64, error) {
	var balance float64
	err := r.DB.QueryRowContext(ctx, `
		SELECT balance FROM accounts WHERE id = $1
	`, accountID).Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, errors.New("account not found")
	}
	return balance, err
}

func (r *AccountRepository) IsAccountOwnedByUser(ctx context.Context, accountID int64, userID int64) (bool, error) {
	var count int
	err := r.DB.GetContext(ctx, &count, `SELECT COUNT(*) FROM accounts WHERE id = $1 AND user_id = $2`, accountID, userID)
	return count > 0, err
}
