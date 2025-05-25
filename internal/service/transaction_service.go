package service

import (
	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"errors"
	"fmt"
	"time"
)

type TransactionService struct {
	repo        repositories.TransactionRepository
	accountRepo repositories.AccountRepository
}

func NewTransactionService(repo repositories.TransactionRepository, accountRepo repositories.AccountRepository) *TransactionService {
	return &TransactionService{repo: repo, accountRepo: accountRepo}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, txn *models.Transaction) (int64, error) {
	if txn.Amount <= 0 {
		return 0, errors.New("amount must be positive")
	}

	txn.Timestamp = time.Now()
	return s.repo.CreateTransaction(ctx, txn)
}

func (s *TransactionService) TransferBetweenAccounts(ctx context.Context, fromID, toID int64, amount float64, description string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be positive")
	}

	balance, err := s.accountRepo.GetAccountBalance(ctx, fromID)
	if err != nil {
		return 0, err
	}
	if balance < amount {
		return 0, errors.New("insufficient funds")
	}

	err = s.repo.UpdateBalancesTx(ctx, fromID, -amount, toID, amount)
	if err != nil {
		return 0, err
	}

	transaction := &models.Transaction{
		FromAccount: fromID,
		ToAccount:   toID,
		Amount:      amount,
		Type:        "transfer",
		Timestamp:   time.Now(),
		Description: description,
	}
	return s.repo.CreateTransaction(ctx, transaction)
}

func (s *TransactionService) Deposit(ctx context.Context, toAccountID int64, amount float64, description string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	// Используем UpdateBalancesTx: откуда = 0 (внешний источник), зачисление на toAccount
	if err := s.repo.UpdateBalancesTx(ctx, 0, 0, toAccountID, amount); err != nil {
		return 0, err
	}

	transaction := &models.Transaction{
		ToAccount:   toAccountID,
		Amount:      amount,
		Type:        "deposit",
		Timestamp:   time.Now(),
		Description: description,
	}

	return s.repo.CreateTransaction(ctx, transaction)
}

func (s *TransactionService) Withdraw(ctx context.Context, fromAccountID int64, amount float64, description string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}
	if err := s.authorizeAccountOwner(ctx, fromAccountID); err != nil {
		return 0, err
	}
	// Проверка баланса до транзакции (необязательна, но даёт более явную ошибку до начала транзакции)
	balance, err := s.accountRepo.GetAccountBalance(ctx, fromAccountID)
	if err != nil {
		return 0, err
	}
	if balance < amount {
		return 0, errors.New("insufficient funds")
	}

	// Снятие средств, toID = 0 (внешняя система или просто уходит из системы)
	if err := s.repo.UpdateBalancesTx(ctx, fromAccountID, -amount, 0, 0); err != nil {
		return 0, err
	}

	transaction := &models.Transaction{
		FromAccount: fromAccountID,
		Amount:      -amount,
		Type:        "withdraw",
		Timestamp:   time.Now(),
		Description: description,
	}

	return s.repo.CreateTransaction(ctx, transaction)
}

func (s *TransactionService) GetTransactionHistory(ctx context.Context, accountID int64) ([]models.Transaction, error) {
	return s.repo.GetTransactionsByAccountID(ctx, accountID)
}

func (s *TransactionService) Transfer(ctx context.Context, fromID, toID int64, amount float64, description string) (int64, error) {

	if err := s.authorizeAccountOwner(ctx, fromID); err != nil {
		return 0, err
	}

	if amount <= 0 {
		return 0, errors.New("amount must be positive")
	}

	balance, err := s.accountRepo.GetAccountBalance(ctx, fromID)
	if err != nil {
		return 0, err
	}
	if balance < amount {
		return 0, errors.New("insufficient funds")
	}

	if err := s.repo.UpdateBalancesTx(ctx, fromID, -amount, toID, amount); err != nil {
		return 0, err
	}

	tx := &models.Transaction{
		FromAccount: fromID,
		ToAccount:   toID,
		Amount:      amount,
		Type:        "transfer",
		Timestamp:   time.Now(),
		Description: description,
	}
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *TransactionService) CreditPayment(ctx context.Context, fromAccountID int64, amount float64, description string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	balance, err := s.accountRepo.GetAccountBalance(ctx, fromAccountID)
	if err != nil {
		return 0, err
	}
	if balance < amount {
		return 0, errors.New("insufficient funds")
	}

	// Списание в пользу банка/кредита (toID = 0)
	if err := s.repo.UpdateBalancesTx(ctx, fromAccountID, -amount, 0, 0); err != nil {
		return 0, err
	}

	transaction := &models.Transaction{
		FromAccount: fromAccountID,
		Amount:      -amount,
		Type:        "credit_payment",
		Timestamp:   time.Now(),
		Description: description,
	}

	return s.repo.CreateTransaction(ctx, transaction)
}

func (s *TransactionService) ReverseTransaction(ctx context.Context, transactionID int64, description string) (int64, error) {
	// Получаем оригинальную транзакцию
	original, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return 0, err
	}

	// Проверка: поддерживаем только определённые типы
	if original.Type != "transfer" && original.Type != "withdraw" && original.Type != "credit_payment" {
		return 0, errors.New("this transaction type cannot be reversed")
	}

	// Проверка: сумма должна быть положительной
	amount := original.Amount
	if amount < 0 {
		amount = -amount
	}

	var fromID, toID int64
	fromID = original.ToAccount
	toID = original.FromAccount

	if fromID == 0 || toID == 0 {
		return 0, errors.New("reversal requires both source and destination accounts")
	}

	// Проверка баланса перед возвратом
	balance, err := s.accountRepo.GetAccountBalance(ctx, fromID)
	if err != nil {
		return 0, err
	}
	if balance < amount {
		return 0, errors.New("insufficient funds to reverse transaction")
	}

	// Перенос средств назад
	err = s.repo.UpdateBalancesTx(ctx, fromID, -amount, toID, amount)
	if err != nil {
		return 0, err
	}

	// Запись обратной транзакции
	reversal := &models.Transaction{
		FromAccount: fromID,
		ToAccount:   toID,
		Amount:      amount,
		Type:        "reversal",
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Reversal of transaction %d: %s", transactionID, description),
	}

	return s.repo.CreateTransaction(ctx, reversal)
}

func (s *TransactionService) authorizeAccountOwner(ctx context.Context, accountID int64) error {
	userIDRaw := ctx.Value(middleware.UserIDKey)
	if userIDRaw == nil {
		return errors.New("unauthenticated")
	}

	userID, ok := userIDRaw.(int64)
	if !ok {
		return errors.New("invalid user id")
	}

	isOwner, err := s.accountRepo.IsAccountOwnedByUser(ctx, accountID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: account does not belong to user")
	}

	return nil
}

func (s *TransactionService) CreditToAccount(ctx context.Context, accountID int64, amount float64, description string) (int64, error) {
	if amount <= 0 {
		return 0, errors.New("amount must be positive")
	}

	err := s.repo.UpdateBalancesTx(ctx, 0, 0, accountID, amount)
	if err != nil {
		return 0, err
	}

	txn := &models.Transaction{
		ToAccount:   accountID,
		Amount:      amount,
		Type:        "credit_payment",
		Timestamp:   time.Now(),
		Description: description,
	}

	return s.CreateTransaction(ctx, txn)
}
