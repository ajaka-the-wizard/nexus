package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id                uuid.UUID `json:"id" db:"id" redis:"id"`
	FullName          string    `json:"full_name" db:"full_name" redis:"full_name"`
	Email             string    `json:"email" db:"email" redis:"email"`
	Verified          bool      `json:"verified" db:"verified" redis:"verified"`
	Paid              bool      `json:"paid" db:"paid" redis:"paid"`
	Active            bool      `json:"active" db:"active" redis:"active"`
	Role              string    `json:"role" db:"role" redis:"role"`
	Password          string    `json:"-" db:"password" redis:"password"`
	PasswordUpdatedAt time.Time `json:"-" db:"password_updated_at" redis:"password_updated_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at" redis:"updated_at"`
}
