package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type MinimalUserStruct struct {
	UserId uuid.UUID `json:"id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
