package redis_store

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const shortLink = "s.shorten.com/000000"
const baseURL = "google.com"

func TestSavingAndRetrievingFromRedis(t *testing.T) {
	// IMPORTANT - ENSURE DB IS MEANT ONLY FOR TESTING BECAUSE FLUSHALL RUNS EVERYTIME
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       9, // use only DB 9 for tests
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer client.Close()

	defer cancel()

	pingRedis(t, ctx, client)
	storeString(t, ctx, client, shortLink, baseURL)
	retrieveString(t, ctx, client, shortLink, baseURL)
	client.FlushAll(ctx)
}

func pingRedis(t testing.TB, ctx context.Context, client *redis.Client) {
	t.Helper()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal("unable to ping redis connection")
	}
}

func storeString(t testing.TB, ctx context.Context, client *redis.Client, shortLink, baseURL string) {
	t.Helper()

	err := client.Set(ctx, shortLink, baseURL, 0).Err()
	if err != nil {
		t.Fatalf("unable to store mapping in redis, %v", err)
	}
}

func retrieveString(t testing.TB, ctx context.Context, client *redis.Client, shortLink, baseURL string) {
	t.Helper()

	val, err := client.Get(ctx, shortLink).Result()

	if err != nil {
		t.Fatalf("unable to retrieve short link from redis, %v", err)
	}

	if val != baseURL {
		t.Errorf("expected %v but got %v", baseURL, val)
	}
}

func TestRedisURLStoreImplementation(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       9, // use only DB 9 for tests
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer client.Close()
	defer cancel()

	urlStore := RedisURLStore{client: client}

	err := urlStore.Save(shortLink, baseURL)

	if err != nil {
		t.Fatalf("Save method error, %v", err)
	}
	retrieveString(t, ctx, urlStore.client, shortLink, baseURL)

	val, found := urlStore.Load(shortLink)

	if !found {
		t.Errorf("unable to find baseURL for shortLink %v", shortLink)
	}

	if val != baseURL {
		t.Errorf("expected %v but got %v", baseURL, val)
	}
}
