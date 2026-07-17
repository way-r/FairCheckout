package checkout

import (
	domain "FairCheckout/internal/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckoutRequest struct {
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	PaymentMethod string `json:"payment_mehtod" binding:"required"`

	ShippingAddress domain.ShippingAddress `json:"shipping_address" binding:"required"`
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (hdl *Handler) Checkout(gctx *gin.Context) {
	var checkoutRequest CheckoutRequest
	if err := gctx.ShouldBindJSON(&checkoutRequest); err != nil {
		slog.Error(
			"CheckoutRequest cannot be parsed",
			"error", err,
		)
		gctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := gctx.Request.Context()
	checkoutResponse := hdl.svc.ProcessCheckout(ctx, checkoutRequest)

	checkoutStatus := checkoutResponse.EventID.StatusCode()
	gctx.JSON(checkoutStatus, gin.H{
		"order_status": checkoutResponse.EventID.String(),
	})
}
