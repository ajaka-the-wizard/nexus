package configs

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	DATABASE_URL             string  `validate:"required"`
	SERVER_ADDR              string  `validate:"required"`
	ENVIRONMENT              string  `validate:"required"`
	JWT_SHARED_KEY           string  `validate:"required"`
	JWT_REFRESH_KEY          string  `validate:"required"`
	JWT_SESSION_DURATION     float64 `validate:"required,gt=0"`
	JWT_REFRESH_KEY_DURATION float64 `validate:"required,gt=0"`
	REDIS_ADDR               string  `validate:"required"`
	REDIS_PASSWORD           string
	PRODUCTION               bool
}

func LoadEnv(logger *slog.Logger) *Env {
	godotenv.Load()

	env := Env{
		DATABASE_URL:    os.Getenv("DATABASE_URL"),
		SERVER_ADDR:     os.Getenv("SERVER_ADDR"),
		ENVIRONMENT:     os.Getenv("ENVIRONMENT"),
		JWT_SHARED_KEY:  os.Getenv("JWT_SHARED_KEY"),
		JWT_REFRESH_KEY: os.Getenv("JWT_REFRESH_KEY"),
		REDIS_ADDR:      os.Getenv("REDIS_ADDR"),
		REDIS_PASSWORD:  os.Getenv("REDIS_PASSWORD"),
	}

	env.PRODUCTION = env.ENVIRONMENT == "production"

	v := validator.New()
	if err := v.Struct(&env); err != nil {
		logger.Error("Environment validation failed", "error", err)
		os.Exit(1)
	}

	var err error
	env.JWT_SESSION_DURATION, err = strconv.ParseFloat(os.Getenv("JWT_SESSION_DURATION"), 64)
	if err != nil {
		logger.Error("Invalid JWT session duration", "error", err)
		os.Exit(1)
	}
	env.JWT_REFRESH_KEY_DURATION, err = strconv.ParseFloat(os.Getenv("JWT_REFRESH_KEY_DURATION"), 64)
	if err != nil {
		logger.Error("Invalid JWT refresh duration", "error", err)
		os.Exit(1)
	}

	logger.Info("Successfully loaded and validated environment vars")
	return &env
}
