package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v86"

	checkout "FairCheckout/internal/checkout"
	config "FairCheckout/internal/config"
	"FairCheckout/internal/datastore"
	logger "FairCheckout/internal/logger"
)

func main() {
	logger.InitLogger()
	appConfig := config.LoadConfigLocal()

	checkoutClient := datastore.ClusterClient(appConfig.Checkout.ClusterAddrs)
	checkoutCache := datastore.NewCheckoutCache(
		checkoutClient,
		appConfig.Checkout.AttemptDuration,
		appConfig.Checkout.OrderDuration,
		appConfig.Checkout.MaxRetries,
	)
	inventoryClient := datastore.Client(appConfig.Inventory.Addr)
	inventoryCache := datastore.NewInventoryCache(
		inventoryClient,
	)
	stripe.Key = appConfig.Stripe.SecretKey

	checkoutService := checkout.NewService(checkoutCache, inventoryCache)
	checkoutHandler := checkout.NewHandler(checkoutService)

	router := gin.Default()
	router.POST("/checkout", checkoutHandler.Checkout)
	router.Run()
}
