package checkout

import (
	domain "FairCheckout/internal/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckoutRequest struct {
	PaymentId       string                 `json:"payment_id" binding:"required"`
	ShippingAddress domain.ShippingAddress `json:"shipping_address" binding:"required"`
}

type CheckoutHandler struct {
	ChekoutService CheckoutService
}

func (chdl *CheckoutHandler) Checkout(gctx *gin.Context) {
	var checkoutRequest CheckoutRequest
	if err := gctx.ShouldBindJSON(&checkoutRequest); err != nil {
		slog.Error("CheckoutRequest cannot be parsed", "error", err)
		gctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := gctx.Request.Context()
	paymentId := checkoutRequest.PaymentId
	shippingAddress := checkoutRequest.ShippingAddress

	checkoutResult := chdl.ChekoutService.ProcessCheckout(ctx, paymentId, shippingAddress)

	checkoutStatus := checkoutResult.EventId.StatusCode()
	gctx.JSON(checkoutStatus, gin.H{
		"order_status": checkoutResult.EventId.String(),
	})
}
