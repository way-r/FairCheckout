package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type CheckoutRequest struct {
	Address1	string `json:"address_1" binding:"required"`
	Address2	string `json:"address_2" binding:"required"`
	State	string `json:"state" binding: "required"`
	City	string `json:"city" binding: "required`
	ZipCode string `json:"zip_code" binding: "required`

	PaymentToken string `json:"payment_token" binding:"required"`
	Amount int64 `json:"amount" binding:"required"`
	Currency string	`json:"currency" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func submitCheckoutRequest(c *gin.Context) {
	var checkoutRequest CheckoutRequest
	if err := c.ShouldBindJSON(&checkoutRequest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	/*
	main logic
	*/
	
	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

func main() {
	router := gin.Default()

	router.POST("/checkout", submitCheckoutRequest)

	router.Run()
}
