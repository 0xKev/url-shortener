package redis_store

import (
	"context"
	"testing"
	"time"

	"github.com/0xKev/url-shortener/internal/model"
	"github.com/0xKev/url-shortener/internal/testutil"
	"github.com/redis/go-redis/v9"
)

const (
	shortSuffix = "000000"
	baseURL     = "google.com"
)

func pingRedis(t testing.TB, ctx context.Context, client *redis.Client) {
	t.Helper()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal("unable to ping redis connection")
	}
}

func storeString(t testing.TB, ctx context.Context, client *redis.Client, shortSuffix, baseURL string) {
	t.Helper()

	err := client.Set(ctx, shortSuffix, baseURL, 0).Err()
	if err != nil {
		t.Fatalf("unable to store mapping in redis, %v", err)
	}
}

func retrieveString(t testing.TB, ctx context.Context, client *redis.Client, shortSuffix, baseURL string) {
	t.Helper()

	val, err := client.Get(ctx, shortSuffix).Result()
	if err != nil {
		t.Fatalf("unable to retrieve short link from redis, %v", err)
	}

	if val != baseURL {
		t.Errorf("expected %v but got %v", baseURL, val)
	}
}

func setupClient() (*redis.Client, context.Context, context.CancelFunc) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       9, // use only DB 9 for tests
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return client, ctx, cancel
}

// TODO(LOW): Figure out if setupConfigurableClient() is necessary
func setupConfigurableClient(config *redis.Options) (*redis.Client, context.Context, context.CancelFunc) {
	client := redis.NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return client, ctx, cancel
}

func TestSavingAndRetrievingFromRedis(t *testing.T) {
	// IMPORTANT - ENSURE DB IS MEANT ONLY FOR TESTING BECAUSE FLUSHALL RUNS EVERYTIME
	client, ctx, cancel := setupClient()

	defer client.Close()

	defer cancel()
	client.FlushAll(ctx)

	pingRedis(t, ctx, client)
	storeString(t, ctx, client, shortSuffix, baseURL)
	retrieveString(t, ctx, client, shortSuffix, baseURL)
}

func TestRedisURLStoreImplementation(t *testing.T) {
	client, ctx, cancel := setupClient()
	defer cancel()
	defer client.Close()
	client.FlushAll(ctx)

	urlStore := RedisURLStore{client: client}

	err := urlStore.Save(&model.URLPair{BaseURL: baseURL, ShortSuffix: shortSuffix})
	if err != nil {
		t.Fatalf("Save method error, %v", err)
	}
	retrieveString(t, ctx, urlStore.client, shortSuffix, baseURL)

	val, found := urlStore.Load(shortSuffix)

	if !found {
		t.Errorf("unable to find baseURL for shortSuffix %v", shortSuffix)
	}

	if val != baseURL {
		t.Errorf("expected %v but got %v", baseURL, val)
	}
}

func TestRedisStoreConfig(t *testing.T) {
	t.Run("create redis store with pre set config", func(t *testing.T) {
		config := &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       9, // use only DB 9 for tests
		}

		_, err := NewRedisURLStore(config)

		testutil.AssertNoError(t, err)
	})

	t.Run("expect error when creating redis store with nil config supplied", func(t *testing.T) {
		_, err := NewRedisURLStore(nil)

		testutil.AssertError(t, err)
	})

	t.Run("with new custom config", func(t *testing.T) {
		config := NewRedisConfig("localhost:6379", "", 9)

		_, err := NewRedisURLStore(config)
		testutil.AssertNoError(t, err)
	})
}
