package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &redisCache{
		client: client,
	}, nil
}

func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get key %s: %v", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %v", err)
	}

	return nil
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %v", key, err)
	}

	return nil
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key %s: %v", key, err)
	}
	return nil
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

var ErrCacheMiss = fmt.Errorf("cache miss")
