package datastore

import (
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

func RedisClusterClient(addresses []string) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addresses,
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
