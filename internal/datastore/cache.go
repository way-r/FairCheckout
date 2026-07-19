package datastore

import (
	"FairCheckout/internal/domain"
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// creates a redis client for inventory
func Client(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error(
			"Error connecting to the redis client",
			"error", err,
		)
		os.Exit(1)
	}
	return client
}

type InventoryCache struct {
	client *redis.Client
}

func NewInventoryCache(c *redis.Client) *InventoryCache {
	return &InventoryCache{
		client: c,
	}
}

var reserveInventoryScript = redis.NewScript(`
	local requested = tonumber(ARGV[1])
	local stock = tonumber(redis.call("GET", KEYS[1]))
	if stock == nil or stock < requested then 
		return 0 
	end
	redis.call("DECRBY", KEYS[1], requested)
	return 1
`)

// reserves stock from the inventory
func (i *InventoryCache) ReserveStock(ctx context.Context, productKey string, quanity int) error {
	result, err := reserveInventoryScript.Run(ctx, i.client, []string{productKey}, quanity).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return domain.ErrOutOfStock
	}
	return nil
}

// releases stock from the inventory
func (i *InventoryCache) ReleaseStock(ctx context.Context, productKey string, checkoutID string, quanity int) {
	err := i.client.IncrBy(ctx, productKey, int64(quanity)).Err()
	if err != nil {
		slog.Error(
			"Failed to release stock",
			"productKey", productKey,
			"checkoutID", checkoutID,
			"err", err,
		)
	}
}

// creates a redis cluster client for checkout
func ClusterClient(addrs []string) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addrs,
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

type CheckoutCache struct {
	client          *redis.ClusterClient
	attemptDuration time.Duration
	orderDuration   time.Duration
	maxRetries      int
}

func NewCheckoutCache(c *redis.ClusterClient, ad time.Duration, od time.Duration, mr int) *CheckoutCache {
	return &CheckoutCache{
		client:          c,
		attemptDuration: ad,
		orderDuration:   od,
		maxRetries:      mr,
	}
}

// locks the address' attempt key with TTL
func (c *CheckoutCache) AcquireLock(ctx context.Context, attemptKey string, checkoutID string) error {
	acquired, err := c.client.SetNX(ctx, attemptKey, checkoutID, c.attemptDuration).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return domain.ErrLockBusy
	}
	return nil
}

var releaseAttemptScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

// releases the address' attempt key
func (c *CheckoutCache) ReleaseLock(ctx context.Context, attemptKey string, checkoutID string) {
	err := releaseAttemptScript.Run(ctx, c.client, []string{attemptKey}, checkoutID).Err()
	if err != nil {
		slog.Error(
			"Failed to release attempt key",
			"attemptKey", attemptKey,
			"checkoutID", checkoutID,
			"err", err,
		)
	}
}

// checks the address' order key
func (c *CheckoutCache) CheckOrder(ctx context.Context, orderKey string) error {
	exists, err := c.client.Exists(ctx, orderKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return domain.ErrDupOrder
	}
	return nil
}

// writes the address' order key with TTL
func (c *CheckoutCache) WriteOrder(ctx context.Context, orderKey string, checkoutID string) {
	written := false

	for i := range c.maxRetries {
		err := c.client.Set(ctx, orderKey, checkoutID, c.orderDuration).Err()
		if err == nil {
			written = true
			break
		}
		slog.Warn(
			"Failed to write orderKey to cache",
			"key", orderKey,
			"checkoutID", checkoutID,
			"attempt", i+1,
			"error", err,
		)
		time.Sleep(200 * time.Millisecond)
	}

	if !written {
		slog.Error(
			"Failed to write orderKey to cache after max attempts",
			"key", orderKey,
			"checkoutID", checkoutID,
		)
	}
}
