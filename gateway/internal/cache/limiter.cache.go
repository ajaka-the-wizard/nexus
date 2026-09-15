package cache

import (
	"context"
	"gateway/internal/domain"
	"log/slog"
	"time"
)

func (r *Redis) Allow(ctx context.Context, logger *slog.Logger, key string, cost uint) (*domain.LimiterResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := r.s.Run(ctx, r.rdb, []string{key}, domain.BUCKET_SIZE, domain.REFILL_RATE_PER_SECOND, cost).Result()
	if err != nil {
		return nil, err
	}

	values, ok := result.([]any)
	if !ok || len(values) != 3 {
		logger.Error("Redis returned an invalid limiter response", "result", result)
		return nil, domain.ErrInvalidRedisReturnType
	}
	return castRedisLimiterValue(logger, values)
}

func castRedisLimiterValue(logger *slog.Logger, values []any) (*domain.LimiterResponse, error) {
	allowed, ok := values[0].(uint64)
	remaining, ok := values[1].(float64)
	retryAfter, ok := values[2].(float64)
	if !ok {
		logger.Error("Redis returned an invalid limiter retry-after value", "values", values)
		return nil, domain.ErrInvalidRedisReturnType
	}

	return &domain.LimiterResponse{
		Allowed:    allowed == 1,
		Remaining:  uint64(remaining),
		RetryAfter: uint64(retryAfter),
	}, nil
}
