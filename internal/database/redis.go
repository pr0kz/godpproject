package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() error {
	db := 0
	if value := os.Getenv("REDIS_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid REDIS_DB: %w", err)
		}
		db = parsed
	}
	host := envOr("REDIS_HOST", "localhost")
	port := envOr("REDIS_PORT", "6379")
	client := redis.NewClient(&redis.Options{Addr: host + ":" + port, Password: os.Getenv("REDIS_PASSWORD"), DB: db})
	var err error
	for attempt := 0; attempt < 8; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = client.Ping(ctx).Err()
		cancel()
		if err == nil {
			RedisClient = client
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	_ = client.Close()
	return fmt.Errorf("connect to Redis: %w", err)
}
func GetRedis() *redis.Client { return RedisClient }
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
