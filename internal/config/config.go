package config

import (
	"FairCheckout/internal/logger"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	RedisClusterAddrs []string
	StripeSecretKey   string
}

// Load config from local env file
func LoadConfigLocal() *AppConfig {
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
		slog.Error(
			"Error reading value for 'REDIS_NODES'",
			"error", err,
		)
	}
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")

	return &AppConfig{
		RedisClusterAddrs: redisClusterAddrs,
		StripeSecretKey:   stripeSecretKey,
	}
}
