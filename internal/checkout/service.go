package checkout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	domain "FairCheckout/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CheckoutService struct {
	RedisClusterClient *redis.ClusterClient
}

type CheckoutResult struct {
	EventId domain.EventId
}

func (cs *CheckoutService) ProcessCheckout(ctx context.Context, paymentId string, shippingAddress domain.ShippingAddress) CheckoutResult {
	// lock the attempt key
	baseKey := shippingAddress.BaseKey()
	attemptKey := fmt.Sprintf("attempt:%s", baseKey)
	checkoutId := uuid.New().String()
	err := cs.acquireLock(ctx, attemptKey, checkoutId)
	if err != nil {
		slog.Info("Can not lock the attempt key", "key", baseKey, "error", err)
		if errors.Is(err, domain.ErrLockBusy) {
			return CheckoutResult{EventId: domain.DuplicatedProcessing}
		}
		return CheckoutResult{EventId: domain.InternalError}
	}
	// unlcok the attempt key when the transaction is completed
	defer cs.releaseLock(context.Background(), attemptKey, checkoutId)

	// check if there has been an order at the same address
	orderKey := fmt.Sprintf("order:%s", baseKey)
	err = cs.checkOrder(ctx, orderKey)
	if err != nil {
		slog.Info("Can not make a duplicated order", "key", baseKey, "error", err)
		if errors.Is(err, domain.ErrDupOrder) {
			return CheckoutResult{EventId: domain.DuplicatedOrder}
		}
		return CheckoutResult{EventId: domain.InternalError}
	}

	// payment processor and async db write

	// write the order key to cache to prevent future orders at the same address
	cs.writeOrder(ctx, 5, orderKey, checkoutId)

	return CheckoutResult{
		EventId: domain.PurchaseCompleted,
	}
}

// Create and lock the attempt Key if it is not yet in cache. Returns an error if it is
func (cs *CheckoutService) acquireLock(ctx context.Context, attemptKey string, checkoutId string) error {
	acquired, err := cs.RedisClusterClient.SetNX(ctx, attemptKey, checkoutId, 20*time.Second).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return domain.ErrLockBusy
	}
	return nil
}

// Delete the attempt key from cache
func (cs *CheckoutService) releaseLock(ctx context.Context, attemptKey string, checkoutId string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	cs.RedisClusterClient.Eval(ctx, script, []string{attemptKey}, checkoutId)
}

// Check if the order key is in cache
func (cs *CheckoutService) checkOrder(ctx context.Context, orderKey string) error {
	exists, err := cs.RedisClusterClient.Exists(ctx, orderKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return domain.ErrDupOrder
	}
	return nil
}

// Write the order key to cache
func (cs *CheckoutService) writeOrder(ctx context.Context, maxRetries int, orderKey string, checkoutId string) {
	written := false

	for i := range maxRetries {
		err := cs.RedisClusterClient.Set(ctx, orderKey, checkoutId, 2*time.Hour).Err()
		if err == nil {
			written = true
			break
		}
		slog.Warn("Failed to write orderKey to cache", "key", orderKey, "attempt", i+1, "error", err)
		time.Sleep(200 * time.Millisecond)
	}

	if !written {
		slog.Error("Failed to write orderKey to cache after max attempts", "key", orderKey)
	}
}
