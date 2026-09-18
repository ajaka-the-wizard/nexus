package configs

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(ctx context.Context, logger *slog.Logger, databaseUrl string) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		logger.Error("unable to parse database url", "error", err.Error())
		os.Exit(1)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Error("unable to create connection pool", "error", err.Error())
		os.Exit(1)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		logger.Error("unable to ping database", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("successfully connected to PostgreSQL database")
	return pool
}
