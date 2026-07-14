package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v86"

	checkout "FairCheckout/internal/checkout"
	config "FairCheckout/internal/config"
	datastore "FairCheckout/internal/datastore"
	logger "FairCheckout/internal/logger"
)

func main() {
	logger.InitLogger()
	appConfig := config.LoadConfigLocal()

	redisClusterClient := datastore.RedisClusterClient(appConfig)
	checkoutService := &checkout.CheckoutService{
		RedisClusterClient: redisClusterClient,
	}
	checkoutHandler := &checkout.CheckoutHandler{
		ChekoutService: *checkoutService,
	}

	stripe.Key = appConfig.StripeSecretKey

	router := gin.Default()
	router.POST("/checkout", checkoutHandler.Checkout)
	router.Run()
}
