package cache

import (
	"auth/internal/configs"
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	db *redis.Client
}

func Initcache(ctx context.Context, env *configs.Env, logger *slog.Logger) *Cache {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr:     env.REDIS_ADDR,
		Password: env.REDIS_PASSWORD,
		DB:       0,
		Protocol: 2,
	})
	err := rdb.Set(ctx, "ping", "pong", 0).Err()
	if err != nil {
		os.Exit(1)
	}
	logger.Info("Redis cache connected successfully")
	return &Cache{
		rdb,
	}
}
