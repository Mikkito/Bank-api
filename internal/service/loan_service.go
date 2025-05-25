package service

import (
	"bank-api/internal/cbr"
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

type LoanService struct {
	repo               repositories.LoanRepository
	accountRepo        repositories.AccountRepository
	cbService          cbr.CBRService
	transactionService *TransactionService
}

func NewLoanService(
	repo *repositories.LoanRepository,
	accountRepo *repositories.AccountRepository,
	cb *cbr.CBRService,
	transactionService *TransactionService,
) *LoanService {
	return &LoanService{
		repo:               *repo,
		accountRepo:        *accountRepo,
		cbService:          *cb,
		transactionService: transactionService,
	}
}

func (s *LoanService) TakeLoan(ctx context.Context, userID, accountID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("loan amount must be positive")
	}

	rate, err := s.cbService.GetKeyRate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get key rate: %w", err)
	}

	// Регистрируем как кредитную транзакцию
	txn := &models.Transaction{
		ToAccount:   accountID,
		Amount:      amount,
		Type:        "credit_payment", // вместо "deposit"
		Description: fmt.Sprintf("Loan issued with interest %.2f%%", rate),
		Timestamp:   time.Now(),
	}
	_, err = s.transactionService.CreateTransaction(ctx, txn)
	if err != nil {
		return fmt.Errorf("failed to create credit transaction: %w", err)
	}

	// Сохраняем кредит
	loan := &models.Loan{
		UserID:       userID,
		AccountID:    accountID,
		Principal:    amount,
		InterestRate: rate,
		CreatedAt:    time.Now(),
		IsRepaid:     false,
	}
	if _, err := s.repo.CreateLoan(ctx, loan); err != nil {
		return fmt.Errorf("failed to create loan: %w", err)
	}

	// Обновляем баланс через транзакцию (именно с зачислением)
	err = s.transactionService.repo.UpdateBalancesTx(ctx, 0, 0, accountID, amount)
	if err != nil {
		return fmt.Errorf("failed to credit account balance: %w", err)
	}

	return nil
}

func (s *LoanService) GetUserLoans(ctx context.Context, userID int64) ([]*models.Loan, error) {
	return s.repo.ListLoansByUserID(ctx, userID)
}

// Получить конкретный займ
func (s *LoanService) GetLoanByID(ctx context.Context, loanID int64) (*models.Loan, error) {
	return s.repo.GetLoanByID(ctx, loanID)
}

func (s *LoanService) MarkLoanAsRepaid(ctx context.Context, loanID int64) error {
	return s.repo.MarkAsRepaid(ctx, loanID)
}

func (s *LoanService) RepayLoan(ctx context.Context, loanID int64) error {
	// Получаем кредит
	loan, err := s.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return fmt.Errorf("failed to get loan: %w", err)
	}
	if loan.IsRepaid {
		return errors.New("loan is already repaid")
	}

	// Считаем текущую задолженность с учётом процентов по времени
	debt, err := s.CalculateOutstandingDebt(ctx, loan)
	if err != nil {
		return fmt.Errorf("failed to calculate debt: %w", err)
	}
	if debt <= 0 {
		// Если уже погашен, обновляем статус
		if err := s.repo.MarkAsRepaid(ctx, loan.ID); err != nil {
			return fmt.Errorf("failed to mark loan as repaid: %w", err)
		}
		return nil
	}

	// Проверка баланса
	balance, err := s.accountRepo.GetAccountBalance(ctx, loan.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}
	if balance < debt {
		return fmt.Errorf("insufficient funds: need %.2f, available %.2f", debt, balance)
	}

	// Списание денег
	_, err = s.transactionService.CreditPayment(
		ctx,
		loan.AccountID,
		debt,
		fmt.Sprintf("Loan repayment for loan ID %d", loan.ID),
	)
	if err != nil {
		return fmt.Errorf("failed to perform repayment transaction: %w", err)
	}

	// Обновляем дату следующего платежа
	nextDue := time.Now().Add(30 * 24 * time.Hour)
	if err := s.repo.UpdateNextPaymentDate(ctx, loan.ID, nextDue); err != nil {
		return fmt.Errorf("failed to update next payment date: %w", err)
	}

	// Перепроверяем задолженность после списания
	newDebt, err := s.CalculateOutstandingDebt(ctx, loan)
	if err != nil {
		return fmt.Errorf("failed to re-check loan debt: %w", err)
	}
	if newDebt <= 0 {
		if err := s.repo.MarkAsRepaid(ctx, loan.ID); err != nil {
			return fmt.Errorf("failed to mark as repaid: %w", err)
		}
	}

	// Записываем транзакцию
	txn := &models.Transaction{
		FromAccount: loan.AccountID,
		Amount:      debt,
		Type:        "credit_payment",
		Description: fmt.Sprintf("Repayment for loan ID %d", loan.ID),
		Timestamp:   time.Now(),
	}
	_, err = s.transactionService.CreateTransaction(ctx, txn)
	if err != nil {
		return fmt.Errorf("failed to log repayment transaction: %w", err)
	}

	return nil
}

// Расчёт задолженности
func (s *LoanService) CalculateOutstandingDebt(ctx context.Context, loan *models.Loan) (float64, error) {
	daysPassed := time.Since(loan.StartDate).Hours() / 24
	dailyRate := loan.InterestRate / 365 / 100
	rawDebt := loan.Principal * (1 + dailyRate*daysPassed)

	// Вычитаем все выплаты
	payments, err := s.repo.GetPayments(ctx, loan.ID)
	if err != nil {
		return 0, err
	}
	var totalPaid float64
	for _, p := range payments {
		totalPaid += p.Amount
	}

	debt := rawDebt - totalPaid
	if debt < 0 {
		debt = 0
	}
	return math.Round(debt*100) / 100, nil
}

// Частичное погашение долга
func (s *LoanService) RepayPartialLoan(ctx context.Context, loanID int64, amount float64) error {
	loan, err := s.repo.GetLoanByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan.IsRepaid {
		return errors.New("loan already repaid")
	}
	return s.repo.AddPayment(ctx, loanID, amount)
}
