package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if c == nil || c.client == nil {
		return nil, redis.Nil
	}
	return c.client.Get(ctx, key).Bytes()
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(ctx, key).Err()
}
