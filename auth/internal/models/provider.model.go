package models

import (
	"time"

	"github.com/google/uuid"
)

type Provider struct {
	Id          uuid.UUID `db:"id"`
	UserId      uuid.UUID `db:"user_id"`
	Provider    string    `db:"provider"`
	ProviderSub string    `db:"provider_sub"`
	Verified    bool      `db:"verified"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
