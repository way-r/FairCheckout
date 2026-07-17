package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v86"

	checkout "FairCheckout/internal/checkout"
	config "FairCheckout/internal/config"
	logger "FairCheckout/internal/logger"
)

func main() {
	logger.InitLogger()
	appConfig := config.LoadConfigLocal()

	checkoutService := checkout.NewService(&appConfig.Redis)
	checkoutHandler := checkout.NewHandler(*checkoutService)

	stripe.Key = appConfig.Stripe.SecretKey

	router := gin.Default()
	router.POST("/checkout", checkoutHandler.Checkout)
	router.Run()
}
