package domain

import (
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type MinimalUserStruct struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}
