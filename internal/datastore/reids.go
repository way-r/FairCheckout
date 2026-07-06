package datastore

import (
	config "FairCheckout/internal/config"
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

func RedisClusterClient(cfg *config.AppConfig) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: cfg.RedisClusterAddrs,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error(
			"Error connecting to the redis cluster",
			"error", err,
		)
		os.Exit(1)
	}
	return client
}
