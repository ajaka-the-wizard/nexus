package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	BLACKLISTKEY string = "BLACKLIST"
)

func (c *Cache) AddToBlacklist(ctx context.Context, identifier string, exp time.Duration) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	key := BLACKLISTKEY + identifier
	err := c.db.Set(ctx, key, identifier, exp).Err()
	if err != nil {
		return false
	}
	return true
}

func (c *Cache) CheckBlackList(ctx context.Context, identifier string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	key := BLACKLISTKEY + identifier
	err := c.db.Get(ctx, key).Err()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
