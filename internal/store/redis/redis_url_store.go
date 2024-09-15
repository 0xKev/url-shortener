package redis_store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// gotta implement these types
// type URLStore interface {
// 	Save(shortLink, baseURL string)
// 	Load(shortLink string) (string, bool)
// }

type RedisURLStore struct {
	client *redis.Client
}

func (r *RedisURLStore) Save(shortLink, baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.client.Set(ctx, shortLink, baseURL, 0).Err()

	if err != nil {
		return fmt.Errorf("error when saving short link to redis, %v", err)
	}

	return nil
}

func (r *RedisURLStore) Load(shortLink string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, shortLink).Result()

	if err != nil {
		return "", false
	}

	return val, true
}
