package checkout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	config "FairCheckout/internal/config"
	datastore "FairCheckout/internal/datastore"
	domain "FairCheckout/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/paymentintent"
)

type Service struct {
	redis               *redis.ClusterClient
	attemptLockDuration time.Duration
	orderLockDuration   time.Duration
	maxRetries          int
}

type Result struct {
	EventID domain.EventID // interpreted by handler
}

func NewService(rdsc *config.RedisConfig) *Service {
	return &Service{
		redis:               datastore.RedisClusterClient(rdsc.ClusterAddrs),
		attemptLockDuration: rdsc.AttemptLockDuration,
		orderLockDuration:   rdsc.OrderLockDuration,
		maxRetries:          rdsc.OrderCacheWriteMaxRetries,
	}
}

func (cs *Service) ProcessCheckout(ctx context.Context, checkoutRequest CheckoutRequest) Result {
	shippingAddress := checkoutRequest.ShippingAddress
	baseKey := shippingAddress.BaseKey()
	attemptKey := fmt.Sprintf("attempt:%s", baseKey)
	checkoutID := uuid.New().String()

	// lock the attempt key
	err := cs.acquireLock(ctx, attemptKey, checkoutID)
	if err != nil {
		slog.Info(
			"Can not lock the attempt key",
			"checkoutID", checkoutID,
			"error", err,
		)
		if errors.Is(err, domain.ErrLockBusy) {
			return Result{EventID: domain.DuplicatedProcessing}
		}
		return Result{EventID: domain.InternalError}
	}
	// unlcok the attempt key when the transaction is completed
	defer cs.releaseLock(context.Background(), attemptKey, checkoutID)

	// check if there has been an order at the same address
	orderKey := fmt.Sprintf("order:%s", baseKey)
	err = cs.checkOrder(ctx, orderKey)
	if err != nil {
		slog.Info(
			"Can not make a duplicated order",
			"checkoutID", checkoutID,
			"error", err,
		)
		if errors.Is(err, domain.ErrDupOrder) {
			return Result{EventID: domain.DuplicatedOrder}
		}
		return Result{EventID: domain.InternalError}
	}

	// create the payment intent to stripe
	paymentIntentParams := &stripe.PaymentIntentParams{
		Amount:        &checkoutRequest.Amount,
		Currency:      &checkoutRequest.Currency,
		PaymentMethod: &checkoutRequest.PaymentMethod,
		Confirm:       stripe.Bool(true),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}
	_, err = paymentintent.New(paymentIntentParams)
	if err != nil {
		slog.Info(
			"Payment submitted but was not accepted",
			"checkoutID", checkoutID,
			"error", err,
		)
		return Result{EventID: domain.PaymentDecline}
	}

	// write the order key to cache to prevent future orders at the same address
	cs.writeOrder(ctx, orderKey, checkoutID)

	// async DB write

	return Result{
		EventID: domain.PurchaseCompleted,
	}
}

// Create and lock the attempt Key if it is not yet in cache. Returns an error if it is
func (cs *Service) acquireLock(ctx context.Context, attemptKey string, checkoutID string) error {
	acquired, err := cs.redis.SetNX(ctx, attemptKey, checkoutID, cs.attemptLockDuration).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return domain.ErrLockBusy
	}
	return nil
}

// Delete the attempt key from cache
func (cs *Service) releaseLock(ctx context.Context, attemptKey string, checkoutID string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	cs.redis.Eval(ctx, script, []string{attemptKey}, checkoutID)
}

// Check if the order key is in cache
func (cs *Service) checkOrder(ctx context.Context, orderKey string) error {
	exists, err := cs.redis.Exists(ctx, orderKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return domain.ErrDupOrder
	}
	return nil
}

// Write the order key to cache
func (cs *Service) writeOrder(ctx context.Context, orderKey string, checkoutID string) {
	written := false

	for i := range cs.maxRetries {
		err := cs.redis.Set(ctx, orderKey, checkoutID, cs.orderLockDuration).Err()
		if err == nil {
			written = true
			break
		}
		slog.Warn(
			"Failed to write orderKey to cache",
			"key", orderKey,
			"attempt", i+1,
			"error", err,
		)
		time.Sleep(200 * time.Millisecond)
	}

	if !written {
		slog.Error(
			"Failed to write orderKey to cache after max attempts",
			"key", orderKey,
		)
	}
}
