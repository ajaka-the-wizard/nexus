package cache

import (
	"context"
	"gateway/internal/configs"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
	s   *redis.Script
}

func InitializeRedis(ctx context.Context, env *configs.Env, logger *slog.Logger) *Redis {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr:     env.REDIS_ADDR,
		Password: env.REDIS_PASSWORD,
		DB:       0,
		Protocol: 2,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Could not ping redis", "error", err)
		panic(err)
	}

	script := loadScript(ctx, rdb, logger)

	logger.Info("Redis cache connected successfully")
	return &Redis{
		rdb: rdb,
		s:   script,
	}
}

func loadScript(ctx context.Context, rdb *redis.Client, logger *slog.Logger) *redis.Script {
	script := redis.NewScript(limiter)
	if _, err := script.Load(ctx, rdb).Result(); err != nil {
		logger.Error("Could not load limiter script into redis", "error", err)
		panic(err)
	}

	return script
}

const (
	limiter string = `
    local data = redis.call('MGET', KEYS[1] .. ':t', KEYS[1] .. ':l')
	local redis_time = redis.call('TIME')
	local now = tonumber(redis_time[1]) + (tonumber(redis_time[2]) / 1000000)
    local bucket_size = tonumber(ARGV[1])
    local refill_rate = tonumber(ARGV[2])
    local cost = tonumber(ARGV[3])

    local tokens = tonumber(data[1]) or bucket_size
    local last = tonumber(data[2]) or now

    local elapsed = math.max(0, now - last)
    tokens = math.min(bucket_size, tokens + elapsed * refill_rate)

    if tokens >= cost then
        tokens = tokens - cost
        local ttl = math.ceil(bucket_size / refill_rate)
        redis.call('MSET', KEYS[1] .. ':t', tokens, KEYS[1] .. ':l', now)
        redis.call('EXPIRE', KEYS[1] .. ':t', ttl)
        redis.call('EXPIRE', KEYS[1] .. ':l', ttl)
		return {1, math.floor(tokens), 0}
    else
		return {0, math.floor(tokens), math.ceil((cost - tokens) / refill_rate)}
    end
`
)
