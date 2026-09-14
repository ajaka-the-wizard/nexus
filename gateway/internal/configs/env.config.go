package configs

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	SERVER_ADDR           string `validate:"required"`
	ENVIRONMENT           string `validate:"required"`
	JWT_SHARED_SECRET_KEY []byte `validate:"required"`
	PRODUCTION            bool
}

func LoadEnv(logger *slog.Logger) *Env {
	godotenv.Load()
	env := Env{
		SERVER_ADDR:           os.Getenv("SERVER_ADDR"),
		ENVIRONMENT:           os.Getenv("ENVIRONMENT"),
		JWT_SHARED_SECRET_KEY: []byte(os.Getenv("JWT_SHARED_SECRET_KEY")),
	}
	env.PRODUCTION = env.ENVIRONMENT == "production"

	v := validator.New()
	if err := v.Struct(&env); err != nil {
		logger.Error("Environment validation failed", "error", err)
		panic(err)
	}
	return &env
}
