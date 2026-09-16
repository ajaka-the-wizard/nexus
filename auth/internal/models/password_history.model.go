package models

import (
	"time"

	"github.com/google/uuid"
)

type PasswordHistory struct {
	Id           uuid.UUID `db:"id"`
	UserId       uuid.UUID `db:"user_id"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	RetiredAt    time.Time `db:"retired_at"`
}
