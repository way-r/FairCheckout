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

type CheckoutConfig struct {
	ClusterAddrs    []string
	AttemptDuration time.Duration
	OrderDuration   time.Duration
	MaxRetries      int
}

type InventoryConfig struct {
	Addr string
}

type StripeConfig struct {
	SecretKey string
}

type AppConfig struct {
	Checkout  CheckoutConfig
	Inventory InventoryConfig
	Stripe    StripeConfig
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

	// checkout cache
	checkoutClusterAddrsStr := os.Getenv("CHECKOUT_ADDRS")
	checkoutClusterAddrs := strings.Split(checkoutClusterAddrsStr, ",")
	if len(checkoutClusterAddrs) == 0 {
		slog.Error(
			"Invalid checkoutClusterAddrsStr",
			"checkoutClusterAddrsStr", checkoutClusterAddrsStr,
		)
		os.Exit(1)
	}
	checkoutAttemptDurationStr := os.Getenv("ATTEMPT_DURATION")
	checkoutAttemptDuration, err := time.ParseDuration(checkoutAttemptDurationStr)
	if err != nil {
		slog.Warn(
			"Invalid checkoutAttemptDurationStr, defaulting to 20s",
			"checkoutAttemptDurationStr", checkoutAttemptDurationStr,
		)
		checkoutAttemptDuration = 20 * time.Second
	}
	checkoutOrderDurationStr := os.Getenv("ORDER_DURATION")
	checkoutOrderDuration, err := time.ParseDuration(checkoutOrderDurationStr)
	if err != nil {
		slog.Warn(
			"Invalid checkoutOrderDurationStr, defaulting to 4h",
			"checkoutOrderDurationStr", checkoutOrderDurationStr,
		)
		checkoutOrderDuration = 4 * time.Hour
	}
	checkoutMaxRetriesStr := os.Getenv("MAX_RETRIES")
	checkoutMaxRetries, err := strconv.Atoi(checkoutMaxRetriesStr)
	if err != nil {
		slog.Warn(
			"Invalid checkoutMaxRetriesStr, defaulting to 5",
			"checkoutMaxRetriesStr", checkoutMaxRetriesStr,
		)
		checkoutMaxRetries = 5
	}
	checkout := CheckoutConfig{
		ClusterAddrs:    checkoutClusterAddrs,
		AttemptDuration: checkoutAttemptDuration,
		OrderDuration:   checkoutOrderDuration,
		MaxRetries:      checkoutMaxRetries,
	}

	// inventory cache
	inventoryAddr := os.Getenv("INVENTORY_ADDR")
	inventory := InventoryConfig{
		Addr: inventoryAddr,
	}

	// payment
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecretKey == "" {
		slog.Error("Error reading value for 'STRIPE_SECRET_KEY'")
	}
	stripe := StripeConfig{
		SecretKey: stripeSecretKey,
	}

	return AppConfig{
		Checkout:  checkout,
		Inventory: inventory,
		Stripe:    stripe,
	}
}
