package models

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"username" validate:"required,alphanum"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
