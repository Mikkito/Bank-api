package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	DB *sqlx.DB
}

// CreateTransaction inserts a new transaction record
func (r *TransactionRepository) CreateTransaction(ctx context.Context, t *models.Transaction) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO transactions (from_account, to_account, amount, type, timestamp, description, is_reversal)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		t.FromAccount, t.ToAccount, t.Amount, t.Type, t.Timestamp, t.Description, t.IsReversal,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetTransactionsByAccountID fetches all transactions where the account is sender or receiver
func (r *TransactionRepository) GetTransactionsByAccountID(ctx context.Context, accountID int64) ([]models.Transaction, error) {
	query := `
		SELECT id, from_account, to_account, amount, type, timestamp, description
		FROM transactions
		WHERE from_account = $1 OR to_account = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var fromAccount, toAccount sql.NullInt64

		err := rows.Scan(
			&tx.ID,
			&fromAccount,
			&toAccount,
			&tx.Amount,
			&tx.Type,
			&tx.Timestamp,
			&tx.Description,
		)
		if err != nil {
			return nil, err
		}

		if fromAccount.Valid {
			tx.FromAccount = fromAccount.Int64
		}
		if toAccount.Valid {
			tx.ToAccount = toAccount.Int64
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// Helper to convert 0 to NULL for optional fields
func nullInt64(val int64) interface{} {
	if val == 0 {
		return nil
	}
	return val
}

func (r *TransactionRepository) UpdateBalancesTx(ctx context.Context, fromID int64, debit float64, toID int64, credit float64) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // пробрасываем панику дальше
		}
	}()

	// Сначала списываем (если fromID задан)
	if fromID > 0 {
		res1, err := tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, debit, fromID)
		if err != nil {
			tx.Rollback()
			return err
		}
		rows1, _ := res1.RowsAffected()
		if rows1 == 0 {
			tx.Rollback()
			return errors.New("source account not found or update failed")
		}
	}

	// Затем зачисляем (если toID задан)
	if toID > 0 {
		res2, err := tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, credit, toID)
		if err != nil {
			tx.Rollback()
			return err
		}
		rows2, _ := res2.RowsAffected()
		if rows2 == 0 {
			tx.Rollback()
			return errors.New("destination account not found or update failed")
		}
	}

	return tx.Commit()
}
func (r *TransactionRepository) GetTransactionByID(ctx context.Context, id int64) (*models.Transaction, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT id, from_account, to_account, amount, type, timestamp, description FROM transactions WHERE id = $1`, id)

	var t models.Transaction
	if err := row.Scan(&t.ID, &t.FromAccount, &t.ToAccount, &t.Amount, &t.Type, &t.Timestamp, &t.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) GetTransactionsByAccount(ctx context.Context, accountID int64) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	query := `
		SELECT 
			id, from_account, to_account, amount, type, description, timestamp, is_reversal 
		FROM 
			transactions 
		WHERE 
			from_account = $1 OR to_account = $1 
		ORDER BY timestamp DESC
	`

	if err := r.DB.SelectContext(ctx, &transactions, query, accountID); err != nil {
		return nil, err
	}

	return transactions, nil
}
