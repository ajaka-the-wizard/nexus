package configs

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	SERVER_ADDR string `mapstructure:"SERVER_ADDR" validate:"required"`
	ENVIRONMENT string `mapstructure:"ENVIRONMENT" validate:"required"`
	PRODUCTION  bool
}

func LoadEnv(logger *slog.Logger) *Env {
	godotenv.Load()
	env := Env{
		SERVER_ADDR: os.Getenv("SERVER_ADDR"),
		ENVIRONMENT: os.Getenv("ENVIRONMENT"),
	}
	env.PRODUCTION = env.ENVIRONMENT == "production"

	v := validator.New()
	if err := v.Struct(&env); err != nil {
		logger.Error("Environment validation failed", "error", err)
		panic(err)
	}
	return &env
}
