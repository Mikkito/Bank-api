package service

import (
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"errors"
)

var ErrAccountAlreadyExists = errors.New("account already exists for this user")

type AccountService struct {
	accountRepo *repositories.AccountRepository
}

func NewAccountService(accountRepo *repositories.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

// Creating account for user
func (s *AccountService) CreateAccount(ctx context.Context, userID int64) (*models.Account, error) {
	exists, err := s.accountRepo.HasAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccountAlreadyExists
	}
	return s.accountRepo.CreateAccount(ctx, userID)
}

func (s *AccountService) GetBalance(ctx context.Context, accountID int64) (float64, error) {
	return s.accountRepo.GetAccountBalance(ctx, accountID)
}
