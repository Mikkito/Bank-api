package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type LoanRepository struct {
	DB *sqlx.DB
}

func NewLoanRepository(db *sqlx.DB) *LoanRepository {
	return &LoanRepository{DB: db}
}

func (r *LoanRepository) CreateLoan(ctx context.Context, loan *models.Loan) (int64, error) {
	query := `
		INSERT INTO loans (user_id, account_id, principal, interest_rate, created_at, is_repaid)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	now := time.Now()
	loan.CreatedAt = now
	loan.IsRepaid = false

	var id int64
	err := r.DB.QueryRowContext(ctx, query,
		loan.UserID, loan.AccountID, loan.Principal,
		loan.InterestRate, loan.CreatedAt, loan.IsRepaid,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	loan.ID = id
	return id, nil
}

func (r *LoanRepository) GetLoanByID(ctx context.Context, id int64) (*models.Loan, error) {
	query := `
		SELECT id, user_id, account_id, principal, interest_rate, created_at, is_repaid
		FROM loans
		WHERE id = $1
	`
	row := r.DB.QueryRowContext(ctx, query, id)

	var loan models.Loan
	err := row.Scan(
		&loan.ID, &loan.UserID, &loan.AccountID, &loan.Principal,
		&loan.InterestRate, &loan.CreatedAt, &loan.IsRepaid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepository) MarkAsRepaid(ctx context.Context, loanID int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE loans SET is_repaid = TRUE WHERE id = $1`,
		loanID,
	)
	return err
}

func (r *LoanRepository) ListLoansByUserID(ctx context.Context, userID int64) ([]*models.Loan, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, user_id, account_id, principal, interest_rate, created_at, is_repaid
		FROM loans
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []*models.Loan
	for rows.Next() {
		var loan models.Loan
		if err := rows.Scan(
			&loan.ID, &loan.UserID, &loan.AccountID,
			&loan.Principal, &loan.InterestRate,
			&loan.CreatedAt, &loan.IsRepaid,
		); err != nil {
			return nil, err
		}
		loans = append(loans, &loan)
	}
	return loans, rows.Err()
}

// Удалить кредит по ID (если потребуется, можно проверить IsRepaid перед удалением)
func (r *LoanRepository) DeleteLoan(ctx context.Context, loanID int64) error {
	_, err := r.DB.ExecContext(ctx, `
		DELETE FROM loans WHERE id = $1
	`, loanID)
	return err
}

// Получить активные кредиты по account_id
func (r *LoanRepository) GetActiveLoansByAccount(ctx context.Context, accountID int64) ([]*models.Loan, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, user_id, account_id, principal, interest_rate, created_at, is_repaid
		FROM loans
		WHERE account_id = $1 AND is_repaid = FALSE
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []*models.Loan
	for rows.Next() {
		var loan models.Loan
		if err := rows.Scan(
			&loan.ID, &loan.UserID, &loan.AccountID,
			&loan.Principal, &loan.InterestRate,
			&loan.CreatedAt, &loan.IsRepaid,
		); err != nil {
			return nil, err
		}
		loans = append(loans, &loan)
	}
	return loans, rows.Err()
}

// Получить общую сумму активных кредитов (principal) по user_id
func (r *LoanRepository) GetTotalOutstandingPrincipal(ctx context.Context, userID int64) (float64, error) {
	var total float64
	err := r.DB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(principal), 0)
		FROM loans
		WHERE user_id = $1 AND is_repaid = FALSE
	`, userID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// Добавить выплату
func (r *LoanRepository) AddPayment(ctx context.Context, loanID int64, amount float64) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO loan_payments (loan_id, amount) 
		VALUES ($1, $2)
	`, loanID, amount)
	return err
}

// Получить выплаты
func (r *LoanRepository) GetPayments(ctx context.Context, loanID int64) ([]models.LoanPayment, error) {
	var payments []models.LoanPayment
	err := r.DB.SelectContext(ctx, &payments, `
		SELECT * FROM loan_payments WHERE loan_id = $1
	`, loanID)
	return payments, err
}

func (r *LoanRepository) UpdateNextPaymentDate(ctx context.Context, loanID int64, nextPayment time.Time) error {
	query := `
		UPDATE loans
		SET next_payment_due = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.DB.ExecContext(ctx, query, nextPayment, loanID)
	if err != nil {
		return err
	}
	return nil
}
