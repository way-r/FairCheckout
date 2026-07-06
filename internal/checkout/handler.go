package checkout

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckoutRequest struct {
	Address1 string `json:"address_1" binding:"required"`
	Address2 string `json:"address_2" binding:"required"`
	State    string `json:"state" binding:"required"`
	City     string `json:"city" binding:"required"`
	ZipCode  string `json:"zip_code" binding:"required"`

	PaymentToken   string `json:"payment_token" binding:"required"`
	Amount         int64  `json:"amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

type CheckoutAddress struct {
	StreetAddress string
	State         string
	City          string
	ZipCode       string
}

type CheckoutHandler struct {
	ChekoutService CheckoutService
}

func (chdl *CheckoutHandler) Checkout(gctx *gin.Context) {
	var checkoutRequest CheckoutRequest
	if err := gctx.ShouldBindJSON(&checkoutRequest); err != nil {
		gctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := gctx.Request.Context()
	checkoutAddress := &CheckoutAddress{
		StreetAddress: checkoutRequest.Address1,
		State:         checkoutRequest.State,
		City:          checkoutRequest.City,
		ZipCode:       checkoutRequest.ZipCode,
	}
	chdl.ChekoutService.ProcessCheckout(ctx, *checkoutAddress)

	gctx.JSON(http.StatusOK, gin.H{"status": "accepted"})
}
