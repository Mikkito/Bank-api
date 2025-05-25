package repositories

import (
	"bank-api/internal/models"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	DB *sqlx.DB
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := r.DB.QueryRowContext(ctx, query, user.UserName, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)
	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE email = $1`
	err := r.DB.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) IsEmailOrUsernameTaken(ctx context.Context, email, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1 OR username = $2)`
	err := r.DB.QueryRowContext(ctx, query, email, username).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, created_at FROM users WHERE id = $1`
	err := r.DB.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.UserName, &user.Email, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}
