package service

import (
	"bank-api/internal/config"
	"bank-api/internal/models"
	"bank-api/internal/repositories"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(ctx context.Context, user *models.User) error {
	// check: email or username not free
	exists, err := s.repo.IsEmailOrUsernameTaken(ctx, user.Email, user.UserName)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email or username already in use")
	}

	// hash pass until save
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Save user
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	// check pass
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// generate JWT
	return generateJWT(user.ID)
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func generateJWT(userID int64) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := config.AppConfig.JWT.Secret
	return token.SignedString([]byte(secret))
}
