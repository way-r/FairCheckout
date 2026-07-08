package checkout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	status "FairCheckout/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CheckoutService struct {
	RedisClusterClient *redis.ClusterClient
}

type CheckoutResult struct {
	EventId status.EventId
}

var ErrLockBusy = errors.New("Lock is currently held by another transaction")
var ErrDupOrder = errors.New("An order has been made to the address")

func (cs *CheckoutService) ProcessCheckout(ctx context.Context, zipCode string, streetAddress string, city string, state string) CheckoutResult {
	// lock the attempt key
	baseKey := fmt.Sprintf(":{%s}%s_%s_%s", zipCode, streetAddress, city, state)
	attemptKey := fmt.Sprintf("attempt:%s", baseKey)
	userId := uuid.New().String()
	err := cs.acquireLock(ctx, attemptKey, userId)
	if err != nil {
		slog.Info("Can not lock the attempt key", "key", baseKey, "error", err)
		if errors.Is(err, ErrLockBusy) {
			return CheckoutResult{EventId: status.OrderProcessing}
		}
		return CheckoutResult{EventId: status.InternalError}
	}
	// unlcok the attempt key when the transaction is completed
	defer cs.releaseLock(context.Background(), attemptKey, userId)

	// check if there has been an order at the same address
	orderKey := fmt.Sprintf("order:%s", baseKey)
	err = cs.checkOrder(ctx, orderKey)
	if err != nil {
		slog.Info("Can not make a duplicated order", "key", baseKey, "error", err)
		if errors.Is(err, ErrDupOrder) {
			return CheckoutResult{EventId: status.DuplicatedAddress}
		}
		return CheckoutResult{EventId: status.InternalError}
	}

	// payment processor and async db write

	// write the order key to cache to prevent future orders at the same address
	cs.writeOrder(ctx, 5, orderKey, userId)

	return CheckoutResult{
		EventId: status.PurchaseCompleted,
	}
}

// Create and lock the attempt Key if it is not yet in cache. Returns an error if it is
func (cs *CheckoutService) acquireLock(ctx context.Context, attemptKey string, userId string) error {
	acquired, err := cs.RedisClusterClient.SetNX(ctx, attemptKey, userId, 20*time.Second).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return ErrLockBusy
	}
	return nil
}

// Delete the attempt key from cache
func (cs *CheckoutService) releaseLock(ctx context.Context, attemptKey string, userId string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	cs.RedisClusterClient.Eval(ctx, script, []string{attemptKey}, userId)
}

// Check if the order key is in cache
func (cs *CheckoutService) checkOrder(ctx context.Context, orderKey string) error {
	exists, err := cs.RedisClusterClient.Exists(ctx, orderKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrDupOrder
	}
	return nil
}

// Write the order key to cache
func (cs *CheckoutService) writeOrder(ctx context.Context, maxAttempts int, orderKey string, userId string) {
	written := false

	for i := 0; i < maxAttempts; i++ {
		err := cs.RedisClusterClient.Set(ctx, orderKey, userId, 2*time.Hour).Err()
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
