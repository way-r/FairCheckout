package config

import (
	"FairCheckout/internal/logger"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type RedisConfig struct {
	ClusterAddrs              []string
	AttemptLockDuration       time.Duration
	OrderLockDuration         time.Duration
	OrderCacheWriteMaxRetries int
}

type StripeConfig struct {
	SecretKey string
}

type AppConfig struct {
	Redis  RedisConfig
	Stripe StripeConfig
}

// Load config from local env file
func LoadConfigLocal() AppConfig {
	logger.InitLogger()
	err := godotenv.Load()
	if err != nil {
		slog.Error(
			"Error reading local .env file",
			"error", err,
		)
		os.Exit(1)
	}

	redisClusterAddrsStr := os.Getenv("REDIS_NODES")
	redisClusterAddrs := strings.Split(redisClusterAddrsStr, ",")
	if len(redisClusterAddrs) == 0 {
		slog.Error("Error reading value for 'REDIS_NODES'")
	}
	attemptLockDurationStr := os.Getenv("ATTEMPT_LOCK_DURATION")
	attemptLockDuration, err := time.ParseDuration(attemptLockDurationStr)
	if err != nil {
		slog.Warn(
			"Invalid attemptLockDurationStr, defaulting to 20s",
			"attemptLockDurationStr", attemptLockDurationStr,
		)
		attemptLockDuration = 20 * time.Second
	}
	orderLockDurationStr := os.Getenv("ORDER_LOCK_DURATION")
	orderLockDuration, err := time.ParseDuration(orderLockDurationStr)
	if err != nil {
		slog.Warn(
			"Invalid orderLockDurationStr, defaulting to 4h",
			"orderLockDurationStr", orderLockDurationStr,
		)
		orderLockDuration = 4 * time.Hour
	}
	orderCacheWriteMaxRetriesStr := os.Getenv("ORDER_CACHE_WRITE_MAX_RETRIES")
	orderCacheWriteMaxRetries, err := strconv.Atoi(orderCacheWriteMaxRetriesStr)
	if err != nil {
		slog.Warn(
			"Invalid orderCacheWriteMaxRetriesStr, defaulting to 5",
			"orderCacheWriteMaxRetriesStr", orderCacheWriteMaxRetriesStr,
		)
		orderCacheWriteMaxRetries = 5
	}

	redis := RedisConfig{
		ClusterAddrs:              redisClusterAddrs,
		AttemptLockDuration:       attemptLockDuration,
		OrderLockDuration:         orderLockDuration,
		OrderCacheWriteMaxRetries: orderCacheWriteMaxRetries,
	}

	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecretKey == "" {
		slog.Error("Error reading value for 'STRIPE_SECRET_KEY'")
	}
	stripe := StripeConfig{
		SecretKey: stripeSecretKey,
	}

	return AppConfig{
		Redis:  redis,
		Stripe: stripe,
	}
}
