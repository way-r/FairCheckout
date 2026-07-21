package checkout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	domain "FairCheckout/internal/domain"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/paymentintent"
)

type CheckoutProcessor interface {
	AcquireLock(ctx context.Context, attemptKey string, checkoutID string) error
	ReleaseLock(ctx context.Context, attemptKey string, checkoutID string)
	CheckOrder(ctx context.Context, orderKey string) error
	WriteOrder(ctx context.Context, orderKey string, checkoutID string)
}
type InventoryProcessor interface {
	ReserveStock(ctx context.Context, productKey string, quanity int) error
	ReleaseStock(ctx context.Context, productKey string, checkoutID string, quanity int)
}

type Service struct {
	checkout  CheckoutProcessor
	inventory InventoryProcessor
}

func NewService(c CheckoutProcessor, i InventoryProcessor) *Service {
	return &Service{
		checkout:  c,
		inventory: i,
	}
}

type Result struct {
	EventID domain.EventID
}

func (s *Service) ProcessCheckout(ctx context.Context, cr CheckoutRequest) Result {
	ProductID := cr.ProductID
	checkoutID := cr.CheckoutID
	shippingAddress := cr.ShippingAddress

	productKey := fmt.Sprintf("item:%s", ProductID)
	addresskey := shippingAddress.BaseKey()
	attemptKey := fmt.Sprintf("attempt:%s_%s", addresskey, productKey)
	orderSuccessful := false

	// lock the attempt key
	err := s.checkout.AcquireLock(ctx, attemptKey, checkoutID)
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
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s.checkout.ReleaseLock(cleanupCtx, attemptKey, checkoutID)
	}()

	// check if there has been an order at the same address
	orderKey := fmt.Sprintf("order:%s", addresskey)
	err = s.checkout.CheckOrder(ctx, orderKey)
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

	// reserve the stock
	err = s.inventory.ReserveStock(ctx, productKey, cr.Quantity)
	if err != nil {
		slog.Info(
			"Cannot reserve stock",
			"checkoutID", checkoutID,
			"error", err,
		)
		if errors.Is(err, domain.ErrOutOfStock) {
			return Result{EventID: domain.OutOfStock}
		}
		return Result{EventID: domain.InternalError}
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if !orderSuccessful {
			s.inventory.ReleaseStock(cleanupCtx, productKey, checkoutID, cr.Quantity)
		}
	}()

	// create and send the paymentintent to Stripe
	params := &stripe.PaymentIntentParams{
		Amount:        &cr.Amount,
		Currency:      &cr.Currency,
		PaymentMethod: &cr.PaymentMethod,
		Confirm:       stripe.Bool(true),
		CaptureMethod: stripe.String("manual"),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}
	params.SetIdempotencyKey(checkoutID)
	_, err = paymentintent.New(params)
	if err != nil {
		slog.Info(
			"Payment submitted but was not accepted",
			"checkoutID", checkoutID,
			"error", err,
		)
		return Result{EventID: domain.PaymentDecline}
	}
	// order successfully made
	orderSuccessful = true
	finalizedCtx := context.WithoutCancel(ctx)

	// write the order key to cache to prevent future orders at the same address
	s.checkout.WriteOrder(finalizedCtx, orderKey, checkoutID)

	// async DB write

	return Result{
		EventID: domain.PurchaseCompleted,
	}
}
